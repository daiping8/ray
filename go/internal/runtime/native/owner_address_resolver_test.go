// Copyright 2026 The Ray Authors.
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

package native

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/proto"
	"github.com/stretchr/testify/assert"
)

// mustActorID builds an ActorID from a hex string, failing the test if the hex
// is malformed.
func mustActorID(t *testing.T, hexStr string) ids.ActorID {
	t.Helper()
	id, err := ids.ActorIDFromHex(hexStr)
	if err != nil {
		t.Fatalf("ActorIDFromHex(%q) failed: %v", hexStr, err)
	}
	return id
}

// TestOwnerAddressResolverCachesPerActor verifies the resolver queries
// getActorInfo once per actor: a repeat resolution of the same actor is served
// from the cache, while a different actor is resolved independently.
func TestOwnerAddressResolverCachesPerActor(t *testing.T) {
	var mu sync.Mutex
	counts := make(map[string]int)
	getActorInfo := func(ctx context.Context, actorID ids.ActorID) (*proto.ActorTableData, error) {
		mu.Lock()
		counts[actorID.Hex()]++
		mu.Unlock()
		return &proto.ActorTableData{Address: &proto.Address{IpAddress: "10.0.0.1", Port: 10001}}, nil
	}

	nr := &NativeRuntime{}
	resolve := nr.newOwnerAddressResolver(getActorInfo)

	actorA := mustActorID(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	actorB := mustActorID(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	first, err := resolve(context.Background(), actorA)
	assert.NoError(t, err)
	assert.NotNil(t, first)
	second, err := resolve(context.Background(), actorA)
	assert.NoError(t, err)
	assert.Same(t, first, second, "a repeated resolution must return the cached address")

	_, err = resolve(context.Background(), actorB)
	assert.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, counts[actorA.Hex()], "same actor must be resolved through getActorInfo exactly once")
	assert.Equal(t, 1, counts[actorB.Hex()], "each distinct actor must be resolved exactly once")
}

// TestOwnerAddressResolverSingleFlight verifies that N concurrent first-time
// resolutions of the same actor share a single getActorInfo call: the per-actor
// lock lets one caller resolve while the rest wait, then hit the populated
// cache.
func TestOwnerAddressResolverSingleFlight(t *testing.T) {
	var calls int32
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	getActorInfo := func(ctx context.Context, actorID ids.ActorID) (*proto.ActorTableData, error) {
		atomic.AddInt32(&calls, 1)
		entered <- struct{}{}
		<-release
		return &proto.ActorTableData{Address: &proto.Address{IpAddress: "10.0.0.1", Port: 10001}}, nil
	}

	nr := &NativeRuntime{}
	resolve := nr.newOwnerAddressResolver(getActorInfo)
	actorID := mustActorID(t, "11111111111111111111111111111111")

	const n = 32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := resolve(context.Background(), actorID); err != nil {
				t.Errorf("resolve failed: %v", err)
			}
		}()
	}
	close(start)

	// The winner of the per-actor lock is parked inside getActorInfo; the rest
	// are waiting on that lock. Releasing now lets them all finish.
	<-entered
	close(release)
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&calls),
		"N concurrent first-time resolutions of one actor must share a single getActorInfo call")
}

// TestOwnerAddressResolverRejectsMissingAddress verifies that a lookup with no
// usable address fails and does not populate the cache, so a later call retries
// the lookup instead of caching the failure.
func TestOwnerAddressResolverRejectsMissingAddress(t *testing.T) {
	cases := []struct {
		name      string
		actorInfo *proto.ActorTableData
	}{
		{"nil actor info", nil},
		{"empty address", &proto.ActorTableData{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls int32
			getActorInfo := func(ctx context.Context, actorID ids.ActorID) (*proto.ActorTableData, error) {
				atomic.AddInt32(&calls, 1)
				return tc.actorInfo, nil
			}

			nr := &NativeRuntime{}
			resolve := nr.newOwnerAddressResolver(getActorInfo)
			actorID := mustActorID(t, "cccccccccccccccccccccccccccccccc")

			_, err := resolve(context.Background(), actorID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "has no address")

			// A failed resolution must not be cached: the second call must reach
			// getActorInfo again.
			_, err = resolve(context.Background(), actorID)
			assert.Error(t, err)
			assert.Equal(t, int32(2), atomic.LoadInt32(&calls),
				"a resolution without an address must not populate the cache")
		})
	}
}

// TestOwnerAddressResolverPropagatesLookupError verifies a getActorInfo failure
// surfaces as a resolver error that wraps the original cause.
func TestOwnerAddressResolverPropagatesLookupError(t *testing.T) {
	sentinel := errors.New("gcs unavailable")
	nr := &NativeRuntime{}
	resolve := nr.newOwnerAddressResolver(func(ctx context.Context, actorID ids.ActorID) (*proto.ActorTableData, error) {
		return nil, sentinel
	})
	actorID := mustActorID(t, "dddddddddddddddddddddddddddddddd")

	_, err := resolve(context.Background(), actorID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	assert.Contains(t, err.Error(), "failed to get actor info for owner")
}
