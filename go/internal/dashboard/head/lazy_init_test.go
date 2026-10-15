package head

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ray-project/ray/go/pkg/gcs"
	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/proto"
)

// fakeGCSClient stubs the gcs.Client surface the dashboard head's
// initWithConfig exercises (GetNodeToConnect + NextJobID). It lets tests drive
// the fail-closed branches without a live GCS or a loaded go_runtime.so.
type fakeGCSClient struct {
	gcs.Client
	getNodeInfo func(ctx context.Context, nodeIP string) (*proto.GcsNodeInfo, error)
	nextJobID   func(ctx context.Context) (ids.JobID, error)
	closed      bool
}

func (f *fakeGCSClient) GetNodeToConnect(ctx context.Context, nodeIP string) (*proto.GcsNodeInfo, error) {
	if f.getNodeInfo != nil {
		return f.getNodeInfo(ctx, nodeIP)
	}
	return nil, errors.New("GetNodeToConnect not stubbed")
}

func (f *fakeGCSClient) NextJobID(ctx context.Context) (ids.JobID, error) {
	if f.nextJobID != nil {
		return f.nextJobID(ctx)
	}
	return ids.NilJobID(), errors.New("NextJobID not stubbed")
}

func (f *fakeGCSClient) Close() error   { f.closed = true; return nil }
func (f *fakeGCSClient) IsClosed() bool { return f.closed }

// runWithFakeGCS runs fn with loadGCSClientFn swapped for a stub that hands
// fn the given fake client, restoring the original after.
func runWithFakeGCS(fake gcs.Client, fn func()) {
	orig := loadGCSClientFn
	loadGCSClientFn = func(opts gcs.ClientOptions) (gcs.Client, error) { return fake, nil }
	defer func() { loadGCSClientFn = orig }()
	fn()
}

// TestInitWithConfigFailsClosedOnGetNodeToConnectError verifies that a
// GetNodeToConnect RPC failure surfaces an error from initWithConfig and the
// preflight (and thus any dial) is never reached: without node information the
// sockets stay empty and the node manager port stays 0, which base.Initialize
// would hand to CoreWorkerProcess as an empty raylet socket (RAY_LOG(FATAL))
// or a port-0 address (QuickExit) killing the whole dashboard.
func TestInitWithConfigFailsClosedOnGetNodeToConnectError(t *testing.T) {
	runWithFakeGCS(&fakeGCSClient{
		getNodeInfo: func(ctx context.Context, nodeIP string) (*proto.GcsNodeInfo, error) {
			return nil, errors.New("gcs rpc failed")
		},
		nextJobID: func(ctx context.Context) (ids.JobID, error) {
			return mustJobID(t, "01020304"), nil
		},
	}, func() {
		origPre := preflightConnectFn
		called := false
		preflightConnectFn = func(addr string, timeout time.Duration) error { called = true; return nil }
		defer func() { preflightConnectFn = origPre }()

		cfg := &HeadConfig{GCSAddress: "127.0.0.1:6379", ClusterIDHex: "abc", NodeIPAddress: "127.0.0.1"}
		err := initWithConfig(cfg)
		if err == nil {
			t.Fatal("initWithConfig returned nil, want error for GetNodeToConnect failure")
		}
		if !errors.Is(err, errNoNodeInfo) {
			t.Fatalf("initWithConfig error = %v, want ErrNoNodeInfo", err)
		}
		if called {
			t.Fatal("preflightConnectFn called despite GetNodeToConnect failure; want fail-closed before any dial")
		}
	})
}

// TestInitWithConfigFailsClosedOnNilNodeInfo verifies that a successful
// GetNodeToConnect returning no node info also fails closed: an empty node
// would leave sockets empty and NodeManagerPort=0, the same QuickExit/FATAL
// hazards as an RPC error.
func TestInitWithConfigFailsClosedOnNilNodeInfo(t *testing.T) {
	runWithFakeGCS(&fakeGCSClient{
		getNodeInfo: func(ctx context.Context, nodeIP string) (*proto.GcsNodeInfo, error) {
			return nil, nil
		},
		nextJobID: func(ctx context.Context) (ids.JobID, error) {
			return mustJobID(t, "01020304"), nil
		},
	}, func() {
		cfg := &HeadConfig{GCSAddress: "127.0.0.1:6379", ClusterIDHex: "abc", NodeIPAddress: "127.0.0.1"}
		err := initWithConfig(cfg)
		if err == nil {
			t.Fatal("initWithConfig returned nil for empty node info, want error (fail-closed)")
		}
		if !errors.Is(err, errNoNodeInfo) {
			t.Fatalf("initWithConfig error = %v, want ErrNoNodeInfo", err)
		}
	})
}

// TestEnsureInitializedCallsInitOnce verifies the uninitialized path invokes
// the initializer exactly once per call (the test binary does not load
// go_runtime.so, so api.IsInitialized() is always false here).
func TestEnsureInitializedCallsInitOnce(t *testing.T) {
	var calls atomic.Int32
	orig := initWithConfigFn
	initWithConfigFn = func(cfg *HeadConfig) error {
		calls.Add(1)
		return nil
	}
	defer func() { initWithConfigFn = orig }()

	cfg := &HeadConfig{GCSAddress: "127.0.0.1:6379", ClusterIDHex: "abc", NodeIPAddress: "127.0.0.1"}
	if err := EnsureInitialized(cfg); err != nil {
		t.Fatalf("EnsureInitialized returned error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("initWithConfig called %d times, want 1", calls.Load())
	}
	if err := EnsureInitialized(cfg); err != nil {
		t.Fatalf("second EnsureInitialized returned error: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("initWithConfig called %d times total, want 2", calls.Load())
	}
}

// TestEnsureInitializedConcurrentNoPanic verifies that concurrent callers
// serialize through initOnce and never trigger the setHandle "already
// initialized" panic, even when every caller observes an uninitialized runtime
// and would otherwise race into InitWithOptions.
func TestEnsureInitializedConcurrentNoPanic(t *testing.T) {
	orig := initWithConfigFn
	initWithConfigFn = func(cfg *HeadConfig) error { return nil }
	defer func() { initWithConfigFn = orig }()

	cfg := &HeadConfig{GCSAddress: "127.0.0.1:6379", ClusterIDHex: "abc", NodeIPAddress: "127.0.0.1"}
	var wg sync.WaitGroup
	errs := make([]error, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = EnsureInitialized(cfg)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent EnsureInitialized[%d] returned error: %v", i, err)
		}
	}
}

// TestEnsureInitializedPropagatesInitError verifies an initialization failure
// is returned to the caller so the calling module can degrade.
func TestEnsureInitializedPropagatesInitError(t *testing.T) {
	sentinel := errors.New("gcs not ready")
	orig := initWithConfigFn
	initWithConfigFn = func(cfg *HeadConfig) error { return sentinel }
	defer func() { initWithConfigFn = orig }()

	cfg := &HeadConfig{GCSAddress: "127.0.0.1:6379", ClusterIDHex: "abc", NodeIPAddress: "127.0.0.1"}
	err := EnsureInitialized(cfg)
	if !errors.Is(err, sentinel) {
		t.Fatalf("EnsureInitialized error = %v, want sentinel %v", err, sentinel)
	}
}

// TestPreflightNodeManagerRejectsUnreachable verifies the preflight fails
// closed: a node_manager_port that refuses TCP connections (the stale
// GCS-recovered head node case) must surface an error so the caller degrades
// instead of letting base.Initialize dial a dead port and QuickExit.
func TestPreflightNodeManagerRejectsUnreachable(t *testing.T) {
	sentinel := errors.New("connection refused")
	orig := preflightConnectFn
	preflightConnectFn = func(addr string, timeout time.Duration) error { return sentinel }
	defer func() { preflightConnectFn = orig }()

	err := preflightNodeManager("10.0.0.1", 12345)
	if err == nil {
		t.Fatal("preflightNodeManager returned nil, want error for unreachable port")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("preflightNodeManager error = %v, want sentinel %v", err, sentinel)
	}
}

// TestPreflightNodeManagerAllowsReachable verifies a reachable port passes.
func TestPreflightNodeManagerAllowsReachable(t *testing.T) {
	orig := preflightConnectFn
	preflightConnectFn = func(addr string, timeout time.Duration) error { return nil }
	defer func() { preflightConnectFn = orig }()

	if err := preflightNodeManager("10.0.0.1", 12345); err != nil {
		t.Fatalf("preflightNodeManager returned error for reachable port: %v", err)
	}
}

func mustJobID(t *testing.T, hex string) ids.JobID {
	t.Helper()
	id, err := ids.JobIDFromHex(hex)
	if err != nil {
		t.Fatalf("JobIDFromHex(%q): %v", hex, err)
	}
	return id
}

// TestPreflightNodeManagerSkipsUnset verifies the preflight is a no-op when
// GCS returned no usable port (so callers that intentionally skip dialing are
// not blocked).
func TestPreflightNodeManagerSkipsUnset(t *testing.T) {
	called := false
	orig := preflightConnectFn
	preflightConnectFn = func(addr string, timeout time.Duration) error {
		called = true
		return nil
	}
	defer func() { preflightConnectFn = orig }()

	for _, tc := range []struct {
		name string
		ip   string
		port int32
	}{
		{"zero port", "10.0.0.1", 0},
		{"empty ip", "", 12345},
	} {
		if err := preflightNodeManager(tc.ip, tc.port); err != nil {
			t.Fatalf("%s: preflightNodeManager returned error: %v", tc.name, err)
		}
	}
	if called {
		t.Fatal("preflightConnectFn called for unset node_manager_port; want no-op")
	}
}
