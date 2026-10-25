// Package native unit tests for *cgoClient.PublishNodeResourceUsage that do
// not require a live GCS. They exercise the pure-Go validation paths (nil
// client, blank keyID / jsonPayload) before any CGO call is attempted.
package native

import (
	"context"
	"strings"
	"testing"

	"github.com/ray-project/ray/go/pkg/gcs"
)

// TestNodeResourceUsagePublisherInterface asserts that *cgoClient satisfies the
// gcs.NodeResourceUsagePublisher interface, which is the contract closed by the
// reporter module's resolvePublisher (a *cgoClient must be selectable as the
// publisher).
func TestNodeResourceUsagePublisherInterface(t *testing.T) {
	var _ gcs.NodeResourceUsagePublisher = (*cgoClient)(nil)
}

// TestPublishNodeResourceUsageNilClient verifies that a nil *cgoClient is
// rejected before any CGO call is attempted.
func TestPublishNodeResourceUsageNilClient(t *testing.T) {
	var c *cgoClient
	err := c.PublishNodeResourceUsage(context.Background(), "RAY_REPORTER:node", `{}`)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPublishNodeResourceUsageRejectsEmptyKeyID verifies that an empty keyID is
// rejected before any CGO call is attempted.
func TestPublishNodeResourceUsageRejectsEmptyKeyID(t *testing.T) {
	c := &cgoClient{}
	err := c.PublishNodeResourceUsage(context.Background(), "", `{}`)
	if err == nil {
		t.Fatal("expected error for empty keyID")
	}
	if !strings.Contains(err.Error(), "key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestPublishNodeResourceUsageRejectsEmptyPayload verifies that an empty
// jsonPayload is rejected before any CGO call is attempted.
func TestPublishNodeResourceUsageRejectsEmptyPayload(t *testing.T) {
	c := &cgoClient{}
	err := c.PublishNodeResourceUsage(context.Background(), "RAY_REPORTER:node", "")
	if err == nil {
		t.Fatal("expected error for empty jsonPayload")
	}
	if !strings.Contains(err.Error(), "payload") {
		t.Fatalf("unexpected error: %v", err)
	}
}
