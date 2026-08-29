/*
Copyright 2026 The Ray Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// cluster_driver is a minimal driver executable that connects to a real local
// Ray cluster through the CGO CoreWorker bridge. It registers the native
// (cluster-mode) runtime initializer, initializes a driver CoreWorker, and
// performs an in-process Put/Get round-trip against the object store.
//
// Worker task execution is not in scope for this example (the Go worker
// entrypoint is not yet migrated); the Remote submit is only asserted to not
// panic inside the driver.
package main

import (
	"fmt"
	"os"

	_ "github.com/ray-project/ray/go/internal/runtime/native"
	"github.com/ray-project/ray/go/pkg/options"
	"github.com/ray-project/ray/go/pkg/runtime/api"
)

func add(a int, b int) int { return a + b }

func main() {
	gcsAddress := os.Getenv("RAY_ADDRESS")
	if gcsAddress == "" {
		gcsAddress = "127.0.0.1:6379"
	}
	jobID := os.Getenv("RAY_JOB_ID")
	if jobID == "" {
		jobID = "01000000"
	}

	initOpts := &options.InitializeOptions{
		WorkerType: options.WorkerTypeDriver,
		Network:    options.NetworkOptions{GcsAddress: gcsAddress},
		Job:        options.JobOptions{JobID: jobID},
	}
	if err := api.InitWithOptions(initOpts); err != nil {
		fmt.Fprintf(os.Stderr, "INIT FAILED: %v\n", err)
		os.Exit(1)
	}
	defer api.Instance().Shutdown()
	fmt.Printf("INIT OK: native/CoreWorker driver initialized (gcs=%s)\n", gcsAddress)

	// Put/Get round-trip inside the driver process (no worker required).
	data := map[string]int{"a": 1, "b": 2, "c": 3}
	ref, err := api.Instance().Put(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "PUT FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PUT OK: %v\n", ref.ObjectID())

	result, err := api.Instance().Get(ref)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GET FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("GET OK: %v\n", result)

	// Submit a remote task. The worker runtime is not migrated in this fork,
	// so execution is expected to not complete; we only assert the submit
	// action itself does not panic inside the driver.
	rref, err := api.Instance().Remote(add).Call(1, 2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "REMOTE CALL FAILED (expected boundary): %v\n", err)
	} else {
		fmt.Printf("REMOTE SUBMIT OK: %v\n", rref.ObjectID())
	}

	fmt.Println("CLUSTER_DRIVER_DONE")
}
