// Copyright 2025 The Ray Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"sync"
	"sync/atomic"

	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/pkg/runtime/contract"
	"github.com/ray-project/ray/go/pkg/runtime/submitter"
)

// releaseQueue receives object IDs whose local reference must be removed after
// their ObjectRef is garbage collected. ObjectRef finalizers must NOT call CGO
// directly: Go runs finalizers on a GC-specialized goroutine where a CGO call
// into the C++ CoreWorker can segfault. Instead, finalizers push the object ID
// onto this channel and a dedicated worker goroutine performs the (serialized)
// CGO RemoveLocalReference outside the GC context.
var releaseQueue chan ids.ObjectID

// releaseWorkerStop is closed to signal the release worker goroutine to exit.
var releaseWorkerStop chan struct{}

// initReleaseWorker starts the background goroutine that drains releaseQueue.
// It is idempotent; calling it multiple times is safe.
func initReleaseWorker() {
	if releaseQueue != nil {
		return
	}
	releaseQueue = make(chan ids.ObjectID, 1024)
	releaseWorkerStop = make(chan struct{})
	go func() {
		for {
			select {
			case objectID := <-releaseQueue:
				removeLocalRefSafe(objectID)
			case <-releaseWorkerStop:
				// Drain any remaining items before exiting.
				for {
					select {
					case objectID := <-releaseQueue:
						removeLocalRefSafe(objectID)
					default:
						return
					}
				}
			}
		}
	}()
}

// stopReleaseWorker stops the release worker goroutine. Called during shutdown
// after all finalizers have been blocked (finalizerMu write lock held).
func stopReleaseWorker() {
	if releaseWorkerStop != nil {
		close(releaseWorkerStop)
	}
}

// removeLocalRefSafe performs the CGO RemoveLocalReference for one object ID,
// guarding against runtime teardown the same way the finalizer body did.
func removeLocalRefSafe(objectID ids.ObjectID) {
	if shutdownComplete.Load() {
		return
	}
	h, ok := tryGetHandle()
	if !ok || h == nil {
		return
	}
	rr := h.Runtime()
	if rr != nil && rr.GetObjectStore() != nil {
		_ = rr.GetObjectStore().RemoveLocalReference(&objectID)
	}
}

// currentHandle is the current runtime handle.
// Set by InitWithOptions, cleared by Shutdown.
// Uses atomic.Value for thread-safe access without explicit locks.
var currentHandle atomic.Value // stores contract.RuntimeHandle

// initMu protects concurrent initialization.
// This prevents multiple goroutines from calling InitWithOptions simultaneously.
var initMu sync.Mutex

// initialized tracks whether the runtime has been initialized.
var initialized atomic.Bool

// shutdownComplete tracks whether shutdown has completed.
// This is used by finalizers to skip RemoveLocalReference calls after shutdown.
var shutdownComplete atomic.Bool

// finalizerMu prevents finalizers from running concurrently with shutdown.
// - Write lock: acquired by shutdown to block all finalizers
// - Read lock: acquired by finalizers to wait for shutdown to complete
//
// This mutex solves the race condition where:
// 1. Finalizer checks shutdownComplete=false (shutdown not started)
// 2. Shutdown runs and sets shutdownComplete=true, releases C++ objects
// 3. Finalizer calls RemoveLocalReference on released C++ objects → SIGSEGV
//
// With this mutex:
// - Shutdown acquires write lock, blocking all finalizers
// - Finalizers acquire read lock, waiting for shutdown to complete
// - After shutdown releases write lock, finalizers see shutdownComplete=true and return safely
var finalizerMu sync.RWMutex

// getHandle returns the current runtime handle.
// Panics if runtime is not initialized.
//
// Deprecated: Use tryGetHandle() for comma-ok style error handling.
// This function is reserved for internal APIs where the caller must
// ensure initialization. Callers that need to handle the "not initialized"
// case gracefully should use tryGetHandle() instead.
//
// Example usage:
//
//	// Public API - return error (preferred)
//	handle, ok := tryGetHandle()
//	if !ok {
//	    return nil, errors.ErrRuntimeNotInitialized
//	}
//
//	// Internal API - panic (use getHandle or handle tryGetHandle result)
//	handle, ok := tryGetHandle()
//	if !ok {
//	    panic("runtime not initialized")
//	}
//
//	// Or use deprecated getHandle() for internal APIs
//	handle := getHandle() // panics if not initialized
func getHandle() contract.RuntimeHandle {
	h := currentHandle.Load()
	if h == nil {
		panic("Ray runtime not initialized. Call api.Init() or api.InitWithOptions() first.")
	}
	return h.(contract.RuntimeHandle)
}

// tryGetHandle returns the handle and true if initialized, or nil and false otherwise.
// This is the preferred way to access the runtime handle in code that needs
// to gracefully handle the "not initialized" case.
//
// Callers that need to return errors (rather than silently handling the error)
// can wrap the result:
//
//	handle, ok := tryGetHandle()
//	if !ok {
//	    return nil, fmt.Errorf("runtime not initialized: call api.Init() first")
//	}
func tryGetHandle() (contract.RuntimeHandle, bool) {
	h := currentHandle.Load()
	if h == nil {
		return nil, false
	}
	return h.(contract.RuntimeHandle), true
}

// tryGetTaskSubmitter returns the task submitter and true if runtime is initialized,
// or nil and false otherwise.
func tryGetTaskSubmitter() (submitter.TaskSubmitter, bool) {
	handle, ok := tryGetHandle()
	if !ok || handle == nil {
		return nil, false
	}

	runtime := handle.Runtime()
	if runtime == nil {
		return nil, false
	}

	return runtime.GetTaskSubmitter(), true
}

// setHandle sets the current runtime handle.
// It provides protection against concurrent initialization.
// Only called by InitWithOptions.
func setHandle(h contract.RuntimeHandle) {
	initMu.Lock()
	defer initMu.Unlock()

	if initialized.Load() {
		panic("Ray runtime already initialized. Multiple calls to InitWithOptions are not allowed.")
	}
	currentHandle.Store(h)
	initialized.Store(true)
	// Start the background release worker before any ObjectRef finalizer can run.
	initReleaseWorker()
}

// clearHandle clears the handle after shutdown.
//
// This function acquires a write lock on finalizerMu to prevent finalizers
// from running concurrently with shutdown. This solves the race condition where
// finalizers might access C++ objects that are being released during shutdown.
func clearHandle() {
	// Acquire write lock - this blocks all finalizers from running until shutdown completes
	finalizerMu.Lock()
	defer finalizerMu.Unlock()

	// Note: atomic.Value cannot store nil, so we don't clear currentHandle.
	// Instead, we rely on initialized=false to indicate shutdown.
	// tryGetHandle() checks initialized first, so it will return nil after shutdown.
	// currentHandle.Store(contract.RuntimeHandle(nil)) // This would panic!
	initialized.Store(false)
	shutdownComplete.Store(true)
	// Stop the release worker; any queued releases are drained before exit.
	stopReleaseWorker()
}
