package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"math"

	// Registers the decoders used by image.DecodeConfig below.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// ErrInvalidImage marks a rejection caused by what the user uploaded rather
// than by a storage failure, so callers can answer 400 instead of 500.
var ErrInvalidImage = errors.New("invalid image")

// MaxBannerBytes caps an uploaded banner. A 1920x1080 JPEG is well under 1 MB;
// 8 MB leaves room for a lossless PNG at that size without letting anyone use
// the bucket as general-purpose file hosting.
const MaxBannerBytes = 8 << 20

// bannerAspectRatio and bannerAspectTolerance mirror the check the staff app
// runs in the browser before submitting, so a crafted request can't slip past
// it. Keep the two in sync.
const (
	bannerAspectRatio     = 16.0 / 9.0
	bannerAspectTolerance = 0.02
)

type imageInfo struct {
	format      string
	contentType string
	extension   string
	width       int
	height      int
}

var imageFormats = map[string]struct {
	contentType string
	extension   string
}{
	"jpeg": {"image/jpeg", "jpg"},
	"png":  {"image/png", "png"},
	"gif":  {"image/gif", "gif"},
	"webp": {"image/webp", "webp"},
}

// inspectImage decodes just enough of data to prove it really is an image we
// recognise and to learn its dimensions. This is the security-relevant step:
// without it the endpoint would happily store HTML or SVG on a public bucket
// and hand back a URL, which is a stored-XSS primitive.
func inspectImage(data []byte) (imageInfo, error) {
	if len(data) == 0 {
		return imageInfo{}, fmt.Errorf("%w: file is empty", ErrInvalidImage)
	}
	if len(data) > MaxBannerBytes {
		return imageInfo{}, fmt.Errorf("%w: file is larger than %d MB", ErrInvalidImage, MaxBannerBytes>>20)
	}

	format, width, height, err := decodeDimensions(data)
	if err != nil {
		return imageInfo{}, err
	}

	meta, known := imageFormats[format]
	if !known {
		return imageInfo{}, fmt.Errorf("%w: unsupported image format %q", ErrInvalidImage, format)
	}
	if width <= 0 || height <= 0 {
		return imageInfo{}, fmt.Errorf("%w: could not determine image dimensions", ErrInvalidImage)
	}

	ratio := float64(width) / float64(height)
	if math.Abs(ratio-bannerAspectRatio) > bannerAspectTolerance {
		return imageInfo{}, fmt.Errorf(
			"%w: banner must be 16:9 (got %dx%d)", ErrInvalidImage, width, height,
		)
	}

	return imageInfo{
		format:      format,
		contentType: meta.contentType,
		extension:   meta.extension,
		width:       width,
		height:      height,
	}, nil
}

func decodeDimensions(data []byte) (format string, width, height int, err error) {
	// image.DecodeConfig covers png/jpeg/gif from the standard library. WebP
	// has no stdlib decoder, so read its RIFF header directly rather than
	// pulling in golang.org/x/image for four bytes of width and height.
	if w, h, ok := webpDimensions(data); ok {
		return "webp", w, h, nil
	}

	cfg, format, decodeErr := image.DecodeConfig(bytes.NewReader(data))
	if decodeErr != nil {
		return "", 0, 0, fmt.Errorf("%w: file is not a supported image (png, jpg, gif or webp)", ErrInvalidImage)
	}
	return format, cfg.Width, cfg.Height, nil
}

// webpDimensions parses the canvas size out of a WebP container. It handles
// the three chunk types in the wild: lossy (VP8 ), lossless (VP8L) and
// extended (VP8X).
func webpDimensions(data []byte) (width, height int, ok bool) {
	if len(data) < 30 || !bytes.Equal(data[0:4], []byte("RIFF")) || !bytes.Equal(data[8:12], []byte("WEBP")) {
		return 0, 0, false
	}

	switch string(data[12:16]) {
	case "VP8 ":
		// Frame header: 3-byte tag, 3-byte start code, then 16-bit width and
		// height whose top two bits are scaling hints.
		if len(data) < 30 || !bytes.Equal(data[23:26], []byte{0x9d, 0x01, 0x2a}) {
			return 0, 0, false
		}
		w := int(binary.LittleEndian.Uint16(data[26:28]) & 0x3fff)
		h := int(binary.LittleEndian.Uint16(data[28:30]) & 0x3fff)
		return w, h, true

	case "VP8L":
		if len(data) < 25 || data[20] != 0x2f {
			return 0, 0, false
		}
		bits := binary.LittleEndian.Uint32(data[21:25])
		w := int(bits&0x3fff) + 1
		h := int((bits>>14)&0x3fff) + 1
		return w, h, true

	case "VP8X":
		if len(data) < 30 {
			return 0, 0, false
		}
		w := int(uint32(data[24]) | uint32(data[25])<<8 | uint32(data[26])<<16)
		h := int(uint32(data[27]) | uint32(data[28])<<8 | uint32(data[29])<<16)
		return w + 1, h + 1, true
	}

	return 0, 0, false
}
