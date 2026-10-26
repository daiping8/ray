// Copyright 2025 The Ray Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ray-project/ray/go/pkg/gcs"
	"github.com/ray-project/ray/go/pkg/ids"
	"google.golang.org/grpc"
)

// fakeKVClient is a minimal in-memory gcs.Client used to observe KV writes.
// It embeds the gcs.Client interface so it satisfies the full interface while
// only overriding the operations exercised by the tests.
type fakeKVClient struct {
	gcs.Client
	kv map[string]map[string][]byte
}

func (f *fakeKVClient) Put(_ context.Context, ns, key string, value []byte, overwrite bool) (bool, error) {
	if f.kv == nil {
		f.kv = map[string]map[string][]byte{}
	}
	if f.kv[ns] == nil {
		f.kv[ns] = map[string][]byte{}
	}
	f.kv[ns][key] = value
	return true, nil
}

func (f *fakeKVClient) Get(_ context.Context, ns, key string) ([]byte, error) {
	if v, ok := f.kv[ns][key]; ok {
		return v, nil
	}
	return nil, gcs.ErrKeyNotFound
}

type stubModule struct{ started bool }

func (s *stubModule) Start(ctx context.Context) error {
	s.started = true
	return nil
}

func (s *stubModule) RegisterHTTP(mux *http.ServeMux) error { return nil }
func (s *stubModule) RegisterGRPC(srv *grpc.Server) error   { return nil }

func TestRegisterAgentToGCS(t *testing.T) {
	fake := &fakeKVClient{}
	nodeID, err := ids.NodeIDFromHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c")
	if err != nil {
		t.Fatalf("NodeIDFromHex: %v", err)
	}
	cfg := Config{
		NodeIP: "1.2.3.4",
		NodeID: nodeID,
		GCS:    fake,
	}
	if err := registerAgentToGCS(context.Background(), cfg, 8888, 8080); err != nil {
		t.Fatalf("registerAgentToGCS: %v", err)
	}
	// The two keys carry different shapes, mirroring Python agent.py:
	// NODE_ID_PREFIX:<id> -> [ip, http, grpc], IP_PREFIX:<ip> -> [id, http, grpc].
	// Head-side consumers resolve the missing field through the other key —
	// e.g. the dashboard head's log module reads vals[0] of the ip-keyed value
	// as a node id, so that slot must never carry the ip.
	if got, want := string(fake.kv[KVNamespaceDashboard][DASHBOARD_AGENT_ADDR_NODE_ID_PREFIX+cfg.NodeID.Hex()]), `["1.2.3.4", 8080, 8888]`; got != want {
		t.Errorf("node-id key value = %q, want %q", got, want)
	}
	if got, want := string(fake.kv[KVNamespaceDashboard][DASHBOARD_AGENT_ADDR_IP_PREFIX+"1.2.3.4"]), fmt.Sprintf("[%q, 8080, 8888]", cfg.NodeID.Hex()); got != want {
		t.Errorf("ip key value = %q, want %q", got, want)
	}
}

func TestRunStartsAllModules(t *testing.T) {
	m1, m2 := &stubModule{}, &stubModule{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Run(ctx, Config{NodeID: ids.NewNodeID(), GCS: &fakeKVClient{}}, []Module{m1, m2}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !m1.started || !m2.started {
		t.Errorf("Run did not start both modules: m1=%v m2=%v", m1.started, m2.started)
	}
}

// TestNewHTTPServerBindsBeforeServe verifies the listen socket is bound by
// the constructor, so the socket accepts connections before the agent
// publishes its address in GCS.
func TestNewHTTPServerBindsBeforeServe(t *testing.T) {
	httpSrv, err := newHTTPServer(Config{NodeIP: "127.0.0.1", ListenPort: 0}, []Module{&stubModule{}})
	if err != nil {
		t.Fatalf("newHTTPServer: %v", err)
	}
	conn, err := net.Dial("tcp", httpSrv.listener.Addr().String())
	if err != nil {
		t.Fatalf("dial bound listener before serve: %v", err)
	}
	conn.Close()
}

// TestRunDegradesWhenHTTPPortTaken verifies the Python parity contract for an
// HTTP bind failure: the agent stays alive (Run returns nil on cancellation),
// the address is not registered in GCS, and module startup still proceeds.
func TestRunDegradesWhenHTTPPortTaken(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("occupy port: %v", err)
	}
	defer lis.Close()
	port := lis.Addr().(*net.TCPAddr).Port

	fake := &fakeKVClient{}
	mod := &stubModule{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, Config{
			NodeIP:     "127.0.0.1",
			ListenPort: port,
			GCS:        fake,
		}, []Module{mod})
	}()

	// Give Run time to fail the HTTP bind and reach the serving state.
	time.Sleep(200 * time.Millisecond)
	if keys := fake.kv[KVNamespaceDashboard]; len(keys) != 0 {
		t.Errorf("agent address registered despite HTTP bind failure: %v", keys)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run = %v, want nil (degraded, not failed)", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancellation")
	}
}

// readPersistedPort reads a port-discovery file written by the agent. The file
// name is "{port_name}_{node_id_hex}" and the content is the decimal port,
// mirroring ray::PersistPort.
func readPersistedPort(t *testing.T, dir string, nodeID ids.NodeID, name string) int {
	t.Helper()
	path := filepath.Join(dir, name+"_"+nodeID.Hex())
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	port, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatalf("parse %s content %q: %v", name, data, err)
	}
	return port
}

// TestRunPersistsDashboardAgentPorts verifies the agent publishes the port
// files raylet's WaitForDashboardAgentPorts waits on when a port is 0
// (dynamic allocation). raylet blocks for these files and CHECK-fails after
// 15s if they never appear, so failing to write them kills startup.
func TestRunPersistsDashboardAgentPorts(t *testing.T) {
	dir := t.TempDir()
	nodeID := ids.NewNodeID()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Run(ctx, Config{
		NodeIP:     "127.0.0.1",
		NodeID:     nodeID,
		SessionDir: dir,
		GCS:        &fakeKVClient{},
	}, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got := readPersistedPort(t, dir, nodeID, "metrics_agent_port"); got <= 0 {
		t.Errorf("metrics_agent_port = %d, want a bound ephemeral port", got)
	}
	if got := readPersistedPort(t, dir, nodeID, "metrics_export_port"); got != -1 {
		t.Errorf("metrics_export_port = %d, want -1 (the Go agent does not export metrics)", got)
	}
	if got := readPersistedPort(t, dir, nodeID, "dashboard_agent_listen_port"); got <= 0 {
		t.Errorf("dashboard_agent_listen_port = %d, want a bound ephemeral port", got)
	}
}

// TestPersistAgentPorts verifies the exact file names and decimal contents for
// injected bound ports. The Go agent does not export metrics, so
// metrics_export_port must be -1.
func TestPersistAgentPorts(t *testing.T) {
	dir := t.TempDir()
	nodeID, err := ids.NodeIDFromHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c")
	if err != nil {
		t.Fatalf("NodeIDFromHex: %v", err)
	}
	if err := persistAgentPorts(Config{SessionDir: dir, NodeID: nodeID}, 12345, 54321); err != nil {
		t.Fatalf("persistAgentPorts: %v", err)
	}
	for _, c := range []struct {
		name string
		want string
	}{
		{"metrics_agent_port", "12345"},
		{"metrics_export_port", "-1"},
		{"dashboard_agent_listen_port", "54321"},
	} {
		path := filepath.Join(dir, c.name+"_"+nodeID.Hex())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", c.name, err)
		}
		if got := string(data); got != c.want {
			t.Errorf("%s content = %q, want %q", c.name, got, c.want)
		}
	}
}
