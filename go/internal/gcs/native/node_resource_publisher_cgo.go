//go:build cgo

package native

/*
#include <stdlib.h>
#include "ray/core_worker/lib/go/gcs_client_bridge.h"
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
	"unsafe"

	"github.com/ray-project/ray/go/pkg/gcs"
)

// Compile-time assertion: *cgoClient satisfies the pkg-level
// gcs.NodeResourceUsagePublisher interface, so the reporter module's
// resolvePublisher can select it as the publisher via the
// cfg.GCS.(gcs.NodeResourceUsagePublisher) type assertion.
var _ gcs.NodeResourceUsagePublisher = (*cgoClient)(nil)

// PublishNodeResourceUsage publishes a node resource usage report to GCS.
// keyID is the GCS key (with the RAY_REPORTER: prefix) under which the usage is
// stored, and jsonPayload is the JSON-serialized resource usage payload. It
// returns an error if the publication fails.
func (c *cgoClient) PublishNodeResourceUsage(ctx context.Context, keyID, jsonPayload string) error {
	if c == nil {
		return fmt.Errorf("gcs client is not initialized")
	}
	if strings.TrimSpace(keyID) == "" {
		return fmt.Errorf("node resource usage key must not be empty")
	}
	if strings.TrimSpace(jsonPayload) == "" {
		return fmt.Errorf("node resource usage payload must not be empty")
	}

	var cErr *C.char
	cKeyID := C.CString(keyID)
	defer C.free(unsafe.Pointer(cKeyID))
	cPayload := C.CString(jsonPayload)
	defer C.free(unsafe.Pointer(cPayload))

	_, err := waitContext(ctx, runAsync(ctx, func() (struct{}, error) {
		ok := C.ray_gcs_client_publisher_publish_node_resource_usage(
			c.getPtr(),
			cKeyID,
			cPayload,
			&cErr,
		)
		if cErr != nil {
			errMsg := C.GoString(cErr)
			C.free(unsafe.Pointer(cErr))
			return struct{}{}, fmt.Errorf("publish node resource usage failed: %s", errMsg)
		}
		if ok == 0 {
			return struct{}{}, fmt.Errorf("publish node resource usage failed")
		}
		return struct{}{}, nil
	}))
	if err != nil {
		return err
	}
	return nil
}
