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
	"fmt"

	"github.com/ray-project/ray/go/pkg/options"
	"github.com/ray-project/ray/go/pkg/runtime/contract"
)

// InitFunc is the type of the initialization function.
// It accepts options and returns a RuntimeHandle.
type InitFunc func(*options.InitializeOptions) (contract.RuntimeHandle, error)

var (
	// initFunc is the registered initialization function.
	// Set automatically by internal packages via init().
	initFunc InitFunc
)

// RegisterInitializer registers an initialization function.
// This function is typically called from init() in internal packages.
// Example:
//
//	func init() {
//	    api.RegisterInitializer(func(opts *options.InitializeOptions) (api.RuntimeHandle, error) {
//	        // Create and initialize runtime...
//	    })
//	}
func RegisterInitializer(fn InitFunc) {
	initFunc = fn
}

// getInitFunc returns the registered initialization function.
// Returns nil if no function has been registered.
func getInitFunc() InitFunc {
	return initFunc
}

// defaultFactory is a fallback factory that returns an error if no initializer is registered.
type defaultFactory struct{}

func (f *defaultFactory) Initialize(opts *options.InitializeOptions) (contract.RuntimeHandle, error) {
	if initFunc != nil {
		return initFunc(opts)
	}
	return nil, fmt.Errorf("no runtime initializer registered. " +
		"Make sure to import github.com/ray-project/ray/go/internal/runtime/native")
}

// globalFactory is the global factory instance.
var globalFactory = &defaultFactory{}

// getFactory returns the global factory instance.
func getFactory() Factory {
	return globalFactory
}

// Factory defines the interface for creating runtime instances.
type Factory interface {
	Initialize(opts *options.InitializeOptions) (contract.RuntimeHandle, error)
}
