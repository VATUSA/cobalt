package config

import (
	"fmt"
	"os"
	"strings"
)

// Azure Blob Storage is the Azure-hosted equivalent of DigitalOcean Spaces
// (see spaces.go), used once StorageProvider() selects "azure_blob". Unlike
// Spaces it isn't S3-compatible, so it has its own auth/signing path in
// storage/azure_blob.go.
//
// STORAGE_PROVIDER exists because the same built image is deployed to both
// DOKS (Spaces) and AKS (Blob) at once during the migration coexistence
// window — see gitops' migration-steps.md §3 — so the choice has to be an
// explicit switch read at request time, not a compile-time or one-shot
// decision.

// StorageProvider selects which object storage backend UploadEventBanner and
// UploadPolicyDocument use. Defaults to "spaces" so existing DOKS overlays
// that don't set this var keep working unchanged.
func StorageProvider() string {
	val, ok := os.LookupEnv("STORAGE_PROVIDER")
	if !ok || val == "" {
		return "spaces"
	}
	return val
}

func AzureStorageAccount() string {
	return os.Getenv("AZURE_STORAGE_ACCOUNT")
}

func AzureStorageKey() string {
	return os.Getenv("AZURE_STORAGE_KEY")
}

// AzureEndpoint is the origin for the storage account, e.g.
// https://vatusadevstorage.blob.core.windows.net. Derived from the account
// name unless AZURE_STORAGE_ENDPOINT overrides it.
func AzureEndpoint() string {
	val, ok := os.LookupEnv("AZURE_STORAGE_ENDPOINT")
	if ok && val != "" {
		return strings.TrimSuffix(val, "/")
	}
	return fmt.Sprintf("https://%s.blob.core.windows.net", AzureStorageAccount())
}

// Event banners live in AZURE_STORAGE_CONTAINER, the Blob equivalent of the
// vatusa-events Spaces bucket.

func AzureContainer() string {
	return os.Getenv("AZURE_STORAGE_CONTAINER")
}

func AzurePublicBaseURL() string {
	val, ok := os.LookupEnv("AZURE_STORAGE_PUBLIC_BASE_URL")
	if ok && val != "" {
		return strings.TrimSuffix(val, "/")
	}
	return fmt.Sprintf("%s/%s", AzureEndpoint(), AzureContainer())
}

// IsAzureBlobConfigured reports whether event banner uploads can be served
// from Azure Blob. When false (and Spaces is also unconfigured) the event
// endpoints fall back to accepting a caller-supplied banner URL only.
func IsAzureBlobConfigured() bool {
	return AzureStorageAccount() != "" && AzureStorageKey() != "" && AzureContainer() != ""
}

// Policy documents live in a separate container, AZURE_STORAGE_DOCS_CONTAINER
// (the Blob equivalent of the vatusa-storage Spaces bucket), but under the
// same storage account, so they reuse AzureStorageAccount()/AzureStorageKey().

func AzureDocsContainer() string {
	return os.Getenv("AZURE_STORAGE_DOCS_CONTAINER")
}

func AzureDocsPublicBaseURL() string {
	val, ok := os.LookupEnv("AZURE_STORAGE_DOCS_PUBLIC_BASE_URL")
	if ok && val != "" {
		return strings.TrimSuffix(val, "/")
	}
	return fmt.Sprintf("%s/%s", AzureEndpoint(), AzureDocsContainer())
}

// IsAzureDocsConfigured reports whether document uploads can be served from
// Azure Blob. When false (and Spaces docs is also unconfigured) the policy
// endpoints fall back to accepting a caller-supplied document_url only.
func IsAzureDocsConfigured() bool {
	return AzureStorageAccount() != "" && AzureStorageKey() != "" && AzureDocsContainer() != ""
}
