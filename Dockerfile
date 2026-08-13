# --platform=$BUILDPLATFORM keeps the build stage on the runner's native arch and
# lets Go cross-compile to $TARGETARCH instead of emulating the whole toolchain
# under QEMU.
FROM --platform=$BUILDPLATFORM golang:1.25.6-alpine3.23 AS build
WORKDIR /go/src/github.com/VATUSA/cobalt
COPY src/ ./
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/server ./cmd/server/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/background ./cmd/background/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/cli ./cmd/cli/

FROM alpine:3.23 AS app
WORKDIR /app
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/server ./
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/background ./
COPY --from=build /go/src/github.com/VATUSA/cobalt/bin/cli ./
COPY sql/ ./sql
ENTRYPOINT ["/app/server"]