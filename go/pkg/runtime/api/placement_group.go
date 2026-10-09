// Copyright 2026 The Ray Authors.
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

package api

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/ray-project/ray/go/pkg/errors"
	"github.com/ray-project/ray/go/pkg/ids"
	"github.com/ray-project/ray/go/pkg/log"
	"github.com/ray-project/ray/go/pkg/runtime/submitter"
	"github.com/ray-project/ray/go/proto"
)

// PlacementStrategy mirrors the Ray core placement strategy. The values match
// rpc.PlacementStrategy (0=PACK, 1=SPREAD, 2=STRICT_PACK, 3=STRICT_SPREAD).
type PlacementStrategy int32

const (
	// PlacementStrategyPack packs bundles onto as few nodes as possible.
	PlacementStrategyPack PlacementStrategy = PlacementStrategy(proto.PlacementStrategy_PACK)
	// PlacementStrategySpread spreads bundles across nodes.
	PlacementStrategySpread PlacementStrategy = PlacementStrategy(proto.PlacementStrategy_SPREAD)
	// PlacementStrategyStrictPack packs all bundles onto a single node.
	PlacementStrategyStrictPack PlacementStrategy = PlacementStrategy(proto.PlacementStrategy_STRICT_PACK)
	// PlacementStrategyStrictSpread places each bundle on a distinct node.
	PlacementStrategyStrictSpread PlacementStrategy = PlacementStrategy(proto.PlacementStrategy_STRICT_SPREAD)
)

// PlacementGroupState mirrors the GCS placement group state. The values match
// PlacementGroupTableData.PlacementGroupState: PENDING=0, PREPARED=1,
// CREATED=2, REMOVED=3, RESCHEDULING=4.
type PlacementGroupState int32

const (
	// PlacementGroupStatePending means the placement group is pending or
	// being scheduled.
	PlacementGroupStatePending PlacementGroupState = PlacementGroupState(proto.PlacementGroupTableData_PENDING)
	// PlacementGroupStatePrepared means the placement group is scheduled and
	// all nodes have prepared the resources.
	PlacementGroupStatePrepared PlacementGroupState = PlacementGroupState(proto.PlacementGroupTableData_PREPARED)
	// PlacementGroupStateCreated means the placement group has been created.
	PlacementGroupStateCreated PlacementGroupState = PlacementGroupState(proto.PlacementGroupTableData_CREATED)
	// PlacementGroupStateRemoved means the placement group has been removed
	// and will not be rescheduled.
	PlacementGroupStateRemoved PlacementGroupState = PlacementGroupState(proto.PlacementGroupTableData_REMOVED)
	// PlacementGroupStateRescheduling means the placement group is being
	// rescheduled because the node it was placed on is dead.
	PlacementGroupStateRescheduling PlacementGroupState = PlacementGroupState(proto.PlacementGroupTableData_RESCHEDULING)
)

// PlacementGroupCreationOptions holds the user-facing placement group creation
// parameters. Bundles is the list of resource bundles that make up the
// placement group; each bundle is a map of resource name to quantity (e.g.
// {"CPU": 1.0}).
type PlacementGroupCreationOptions struct {
	Name     string
	Bundles  []map[string]float64
	Strategy PlacementStrategy
}

// NewPlacementGroupCreationOptionsBuilder returns a new builder with defaults.
func NewPlacementGroupCreationOptionsBuilder() *PlacementGroupCreationOptionsBuilder {
	return &PlacementGroupCreationOptionsBuilder{}
}

// PlacementGroupCreationOptionsBuilder builds PlacementGroupCreationOptions.
type PlacementGroupCreationOptionsBuilder struct {
	opts PlacementGroupCreationOptions
}

// WithName sets the unique name of the placement group.
func (b *PlacementGroupCreationOptionsBuilder) WithName(name string) *PlacementGroupCreationOptionsBuilder {
	b.opts.Name = name
	return b
}

// WithBundles sets the resource bundles that make up the placement group.
func (b *PlacementGroupCreationOptionsBuilder) WithBundles(bundles []map[string]float64) *PlacementGroupCreationOptionsBuilder {
	b.opts.Bundles = bundles
	return b
}

// WithStrategy sets the placement strategy.
func (b *PlacementGroupCreationOptionsBuilder) WithStrategy(s PlacementStrategy) *PlacementGroupCreationOptionsBuilder {
	b.opts.Strategy = s
	return b
}

// Build validates the options and returns them.
func (b *PlacementGroupCreationOptionsBuilder) Build() (*PlacementGroupCreationOptions, error) {
	if b.opts.Name == "" {
		return nil, fmt.Errorf("placement group name is required")
	}
	if len(b.opts.Bundles) == 0 {
		return nil, fmt.Errorf("at least one bundle is required")
	}
	return &b.opts, nil
}

// toSubmitterOptions converts the public options into the submitter-layer
// options, mapping the placement strategy to its int32 representation.
func (o *PlacementGroupCreationOptions) toSubmitterOptions() *submitter.PlacementGroupCreationOptions {
	return &submitter.PlacementGroupCreationOptions{
		Name:     o.Name,
		Bundles:  o.Bundles,
		Strategy: int32(o.Strategy),
	}
}

// placementGroupsFacade is the package-level placement group entry point.
type placementGroupsFacade struct{}

// PlacementGroups exposes placement group operations.
var PlacementGroups = placementGroupsFacade{}

// CreatePlacementGroup creates a placement group and returns it.
func CreatePlacementGroup(ctx context.Context, opts *PlacementGroupCreationOptions) (*PlacementGroup, error) {
	return PlacementGroups.CreatePlacementGroup(ctx, opts)
}

// RemovePlacementGroup removes an existing placement group by id.
func RemovePlacementGroup(ctx context.Context, id ids.PlacementGroupID) error {
	return PlacementGroups.RemovePlacementGroup(ctx, id)
}

// GetPlacementGroup queries a placement group by id.
func GetPlacementGroup(ctx context.Context, id ids.PlacementGroupID) (*PlacementGroup, error) {
	return PlacementGroups.GetPlacementGroup(ctx, id)
}

// GetPlacementGroupByName queries a placement group by name within the given
// namespace. An empty namespace falls back to the current namespace.
func GetPlacementGroupByName(ctx context.Context, name, namespace string) (*PlacementGroup, error) {
	return PlacementGroups.GetPlacementGroupByName(ctx, name, namespace)
}

// GetAllPlacementGroups lists all placement groups.
func GetAllPlacementGroups(ctx context.Context) ([]*PlacementGroup, error) {
	return PlacementGroups.GetAllPlacementGroups(ctx)
}

// CreatePlacementGroup creates a placement group via the task submitter and
// returns a PlacementGroup value object carrying the generated id together
// with the creation options.
func (p placementGroupsFacade) CreatePlacementGroup(ctx context.Context, opts *PlacementGroupCreationOptions) (*PlacementGroup, error) {
	if opts == nil {
		return nil, fmt.Errorf("placement group creation options are required")
	}
	// submitter.PlacementGroupCreationOptions.Validate is the single source of
	// truth for option validation (name, bundles, strategy), so the builder
	// does not duplicate the strategy check.
	so := opts.toSubmitterOptions()
	if err := so.Validate(); err != nil {
		return nil, err
	}
	s := getTaskSubmitter()
	if s == nil {
		return nil, errors.NewRuntimeError("create_placement_group", submitterNotAvailable)
	}
	id, err := s.CreatePlacementGroup(ctx, so)
	if err != nil {
		return nil, err
	}
	return newPlacementGroupFromOptions(opts, id), nil
}

// RemovePlacementGroup removes a placement group via the task submitter.
func (p placementGroupsFacade) RemovePlacementGroup(ctx context.Context, id ids.PlacementGroupID) error {
	s := getTaskSubmitter()
	if s == nil {
		return errors.NewRuntimeError("remove_placement_group", submitterNotAvailable)
	}
	return s.RemovePlacementGroup(ctx, id)
}

// WaitPlacementGroupReady blocks until the placement group is ready via the
// task submitter.
func (p placementGroupsFacade) WaitPlacementGroupReady(ctx context.Context, id ids.PlacementGroupID, timeout time.Duration) error {
	s := getTaskSubmitter()
	if s == nil {
		return errors.NewRuntimeError("wait_placement_group_ready", submitterNotAvailable)
	}
	return s.WaitPlacementGroupReady(ctx, id, timeout)
}

// GetPlacementGroup queries a placement group by id via the GCS client.
func (p placementGroupsFacade) GetPlacementGroup(ctx context.Context, id ids.PlacementGroupID) (*PlacementGroup, error) {
	c, err := GetGCSClient()
	if err != nil {
		return nil, err
	}
	data, err := c.GetPlacementGroupInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return newPlacementGroupFromProto(data), nil
}

// GetPlacementGroupByName queries a placement group by name via the GCS client.
// An empty namespace falls back to the current namespace, matching Java's
// AbstractRayRuntime.getPlacementGroup(name, namespace).
func (p placementGroupsFacade) GetPlacementGroupByName(ctx context.Context, name, namespace string) (*PlacementGroup, error) {
	if namespace == "" {
		namespace = currentNamespace()
	}
	c, err := GetGCSClient()
	if err != nil {
		return nil, err
	}
	data, err := c.GetPlacementGroupInfoByName(ctx, name, namespace)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return newPlacementGroupFromProto(data), nil
}

// currentNamespace returns the current namespace from the runtime context.
// It returns an empty string when no runtime context is available, in which
// case callers fall back to the underlying GCS "default" namespace.
func currentNamespace() string {
	rc, err := GetRuntimeContext()
	if err != nil || rc == nil {
		return ""
	}
	return rc.Namespace()
}

// GetAllPlacementGroups lists all placement groups via the GCS client.
func (p placementGroupsFacade) GetAllPlacementGroups(ctx context.Context) ([]*PlacementGroup, error) {
	c, err := GetGCSClient()
	if err != nil {
		return nil, err
	}
	list, err := c.GetAllPlacementGroupInfo(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*PlacementGroup, 0, len(list))
	for _, d := range list {
		if d == nil {
			continue
		}
		out = append(out, newPlacementGroupFromProto(d))
	}
	return out, nil
}

// PlacementGroup is the user-facing placement group value object. Its fields
// are immutable and exposed through accessor methods, matching Java's
// PlacementGroup interface.
type PlacementGroup struct {
	id       ids.PlacementGroupID
	name     string
	bundles  []map[string]float64
	strategy PlacementStrategy
	state    PlacementGroupState
}

// ID returns the placement group id.
func (p *PlacementGroup) ID() ids.PlacementGroupID { return p.id }

// Name returns the placement group name.
func (p *PlacementGroup) Name() string { return p.name }

// Bundles returns the resource bundles that make up the placement group.
func (p *PlacementGroup) Bundles() []map[string]float64 { return p.bundles }

// Strategy returns the placement strategy.
func (p *PlacementGroup) Strategy() PlacementStrategy { return p.strategy }

// State returns the placement group state.
func (p *PlacementGroup) State() PlacementGroupState { return p.state }

// Wait blocks until the placement group is ready or the timeout expires.
// It returns (true, nil) when the group is ready and (false, nil) when the
// timeout expires before the group becomes ready; a real failure returns
// (false, error).
func (p *PlacementGroup) Wait(timeoutSeconds int) (bool, error) {
	if timeoutSeconds <= 0 {
		return false, fmt.Errorf("wait placement group: timeout must be positive, got %d", timeoutSeconds)
	}
	s := getTaskSubmitter()
	if s == nil {
		return false, errors.NewRuntimeError("wait_placement_group", submitterNotAvailable)
	}
	err := s.WaitPlacementGroupReady(context.Background(), p.id, time.Duration(timeoutSeconds)*time.Second)
	if err == nil {
		return true, nil
	}
	if stderrors.Is(err, submitter.ErrPlacementGroupNotReady) {
		return false, nil
	}
	return false, err
}

// newPlacementGroupFromOptions builds a PlacementGroup value object from the
// creation options and the id returned by the submitter.
func newPlacementGroupFromOptions(opts *PlacementGroupCreationOptions, id ids.PlacementGroupID) *PlacementGroup {
	return &PlacementGroup{
		id:       id,
		name:     opts.Name,
		bundles:  opts.Bundles,
		strategy: PlacementStrategy(opts.Strategy),
	}
}

// newPlacementGroupFromProto converts GCS placement group table data into the
// public PlacementGroup value object. Invalid/absent ids are left as the nil
// id rather than failing the whole query.
func newPlacementGroupFromProto(data *proto.PlacementGroupTableData) *PlacementGroup {
	if data == nil {
		return nil
	}
	pg := &PlacementGroup{
		name:     data.Name,
		bundles:  bundlesFromProto(data.Bundles),
		strategy: PlacementStrategy(data.Strategy),
		state:    PlacementGroupState(data.State),
	}
	if id, err := ids.PlacementGroupIDFromBinary(data.PlacementGroupId); err == nil {
		pg.id = id
	} else {
		log.Log.Error(err, "failed to parse placement group id from GCS data")
	}
	return pg
}

// bundlesFromProto converts the GCS bundle specs into the resource-map form
// used by the public API, mirroring PlacementGroupCreationOptions.Bundles.
func bundlesFromProto(bundles []*proto.Bundle) []map[string]float64 {
	if len(bundles) == 0 {
		return nil
	}
	res := make([]map[string]float64, 0, len(bundles))
	for _, b := range bundles {
		if b == nil {
			continue
		}
		res = append(res, b.UnitResources)
	}
	return res
}
