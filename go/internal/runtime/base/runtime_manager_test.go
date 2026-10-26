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

package base

import (
	"syscall"
	"testing"

	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/pkg/runtime/contract"
	"github.com/ray-project/ray/go/pkg/runtime/function"
	"github.com/ray-project/ray/go/pkg/runtime/object"
	"github.com/ray-project/ray/go/pkg/runtime/submitter"
)

// shutdownSignalRuntime is a minimal contract.Runtime whose only observable
// behavior is recording whether Shutdown was called. It lets the shutdown-hook
// routing tests observe the Shutdown path without a live runtime.
type shutdownSignalRuntime struct {
	shutdownCalled bool
}

func (r *shutdownSignalRuntime) Start() error                          { return nil }
func (r *shutdownSignalRuntime) Shutdown() error                       { r.shutdownCalled = true; return nil }
func (r *shutdownSignalRuntime) Run() error                            { return nil }
func (r *shutdownSignalRuntime) IsInitialized() bool                   { return true }
func (r *shutdownSignalRuntime) WorkerContext() contract.WorkerContext { return nil }
func (r *shutdownSignalRuntime) GetRunMode() contract.RunMode          { return contract.RunModeCluster }
func (r *shutdownSignalRuntime) IsLocalMode() bool                     { return false }
func (r *shutdownSignalRuntime) WasCurrentActorRestarted() bool        { return false }
func (r *shutdownSignalRuntime) GetAllNodeInfo() []contract.NodeInfo   { return nil }
func (r *shutdownSignalRuntime) GetAllActorInfo() []contract.ActorInfo { return nil }
func (r *shutdownSignalRuntime) GetAllActorInfoByFilter(jobID ids.JobID, state contract.ActorState) []contract.ActorInfo {
	return nil
}
func (r *shutdownSignalRuntime) GetGpuIds() []string { return nil }
func (r *shutdownSignalRuntime) GetResourceIds() map[string][]contract.ResourceValue {
	return nil
}
func (r *shutdownSignalRuntime) GetCurrentActorHandle() submitter.ActorHandle { return nil }
func (r *shutdownSignalRuntime) GetObjectStore() object.ObjectStore           { return nil }
func (r *shutdownSignalRuntime) GetTaskSubmitter() submitter.TaskSubmitter    { return nil }
func (r *shutdownSignalRuntime) GetFunctionManager() function.Manager         { return nil }
func (r *shutdownSignalRuntime) CreateRuntimeEnv() contract.RuntimeEnv        { return contract.RuntimeEnv{} }
func (r *shutdownSignalRuntime) DeserializeRuntimeEnv(serializedRuntimeEnv string) (contract.RuntimeEnv, error) {
	return contract.UnmarshalRuntimeEnv(serializedRuntimeEnv)
}

// setupShutdownSignalTest installs a fake runtime as the global handle, sets
// the worker-mode marker, and injects a counting fake graceful-exit bridge.
// It restores the package-global shutdown-hook state on cleanup.
func setupShutdownSignalTest(t *testing.T, workerMode bool) (*shutdownSignalRuntime, *int) {
	t.Helper()

	rt := &shutdownSignalRuntime{}
	// Explicit Runtime type parameter: Shutdown() casts the handle to
	// *RuntimeHandleImpl[Runtime], and the real Initialize() also handles a
	// Runtime interface, so the fake must be wrapped the same way.
	handle := NewRuntimeHandle[Runtime](rt)

	manager.globalLock.Lock()
	manager.globalRuntime = rt
	manager.cachedHandle = handle
	manager.handleRegistry.Store(handle)
	manager.globalLock.Unlock()

	workerModeInitialized.Store(workerMode)

	requests := 0
	SetGracefulExitRequester(func() { requests++ })

	t.Cleanup(func() {
		workerModeInitialized.Store(false)
		SetGracefulExitRequester(nil)
		manager.globalLock.Lock()
		manager.globalRuntime = nil
		manager.cachedHandle = nil
		manager.globalLock.Unlock()
	})

	return rt, &requests
}

// TestHandleShutdownSignalWorkerModeRequestsGracefulExit pins the SIGTERM
// worker-mode routing: the injected graceful-exit bridge is invoked and the
// Go-side Shutdown path is NOT taken, so the core worker owns the exit
// (INTENDED_USER_EXIT) instead of being torn down from Go.
func TestHandleShutdownSignalWorkerModeRequestsGracefulExit(t *testing.T) {
	rt, requests := setupShutdownSignalTest(t, true)

	handleShutdownSignal(syscall.SIGTERM)

	if *requests != 1 {
		t.Fatalf("graceful exit bridge called %d times, want 1", *requests)
	}
	if rt.shutdownCalled {
		t.Fatal("worker-mode SIGTERM must not run the Go-side Shutdown path")
	}
}

// TestHandleShutdownSignalDriverModeShutsDown pins that SIGTERM in driver mode
// keeps the pre-existing behavior: the runtime is shut down from Go and the
// graceful-exit bridge is not consulted.
func TestHandleShutdownSignalDriverModeShutsDown(t *testing.T) {
	rt, requests := setupShutdownSignalTest(t, false)

	handleShutdownSignal(syscall.SIGTERM)

	if *requests != 0 {
		t.Fatalf("graceful exit bridge called %d times, want 0", *requests)
	}
	if !rt.shutdownCalled {
		t.Fatal("driver-mode SIGTERM must run the Go-side Shutdown path")
	}
}

// TestHandleShutdownSignalSigIntKeepsShutdownPath pins that SIGINT is not
// routed to the graceful-exit bridge in any mode: driver Ctrl-C semantics are
// unchanged and the runtime is shut down from Go.
func TestHandleShutdownSignalSigIntKeepsShutdownPath(t *testing.T) {
	rt, requests := setupShutdownSignalTest(t, true)

	handleShutdownSignal(syscall.SIGINT)

	if *requests != 0 {
		t.Fatalf("graceful exit bridge called %d times, want 0", *requests)
	}
	if !rt.shutdownCalled {
		t.Fatal("SIGINT must keep the Go-side Shutdown path")
	}
}

// TestHandleShutdownSignalWorkerModeWithoutBridgeFallsBack pins the fallback
// when no graceful-exit bridge is installed (e.g. a build without the cgo
// layer): worker-mode SIGTERM degrades to the pre-existing Shutdown path
// instead of dropping the signal.
func TestHandleShutdownSignalWorkerModeWithoutBridgeFallsBack(t *testing.T) {
	rt, requests := setupShutdownSignalTest(t, true)
	SetGracefulExitRequester(nil)

	handleShutdownSignal(syscall.SIGTERM)

	if *requests != 0 {
		t.Fatalf("graceful exit bridge called %d times, want 0", *requests)
	}
	if !rt.shutdownCalled {
		t.Fatal("worker-mode SIGTERM without a bridge must fall back to the Shutdown path")
	}
}
