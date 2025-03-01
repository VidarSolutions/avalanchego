// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chains

import (
	"context"

	"github.com/VidarSolutions/avalanchego/api/metrics"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/snow/engine/avalanche/vertex"
	"github.com/VidarSolutions/avalanchego/snow/engine/common"
	"github.com/VidarSolutions/avalanchego/snow/engine/snowman/block"

	dbManager "github.com/VidarSolutions/avalanchego/database/manager"
)

var (
	_ vertex.LinearizableVM = (*initializeOnLinearizeVM)(nil)
	_ block.ChainVM         = (*linearizeOnInitializeVM)(nil)
)

// initializeOnLinearizeVM transforms the consensus engine's call to Linearize
// into a call to Initialize. This enables the proposervm to be initialized by
// the call to Linearize. This also provides the stopVertexID to the
// linearizeOnInitializeVM.
type initializeOnLinearizeVM struct {
	vertex.DAGVM
	vmToInitialize common.VM
	vmToLinearize  *linearizeOnInitializeVM

	registerer   metrics.OptionalGatherer
	ctx          *snow.Context
	dbManager    dbManager.Manager
	genesisBytes []byte
	upgradeBytes []byte
	configBytes  []byte
	toEngine     chan<- common.Message
	fxs          []*common.Fx
	appSender    common.AppSender
}

func (vm *initializeOnLinearizeVM) Linearize(ctx context.Context, stopVertexID ids.ID) error {
	vm.vmToLinearize.stopVertexID = stopVertexID
	vm.ctx.Metrics = vm.registerer
	return vm.vmToInitialize.Initialize(
		ctx,
		vm.ctx,
		vm.dbManager,
		vm.genesisBytes,
		vm.upgradeBytes,
		vm.configBytes,
		vm.toEngine,
		vm.fxs,
		vm.appSender,
	)
}

// linearizeOnInitializeVM transforms the proposervm's call to Initialize into a
// call to Linearize. This enables the proposervm to provide its toEngine
// channel to the VM that is being linearized.
type linearizeOnInitializeVM struct {
	vertex.LinearizableVMWithEngine
	stopVertexID ids.ID
}

func (vm *linearizeOnInitializeVM) Initialize(
	ctx context.Context,
	_ *snow.Context,
	_ dbManager.Manager,
	_ []byte,
	_ []byte,
	_ []byte,
	toEngine chan<- common.Message,
	_ []*common.Fx,
	_ common.AppSender,
) error {
	return vm.Linearize(ctx, vm.stopVertexID, toEngine)
}
