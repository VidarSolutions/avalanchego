// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"reflect"

	"github.com/VidarSolutions/avalanchego/codec"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/vms/avm/config"
	"github.com/VidarSolutions/avalanchego/vms/avm/fxs"
)

type Backend struct {
	Ctx           *snow.Context
	Config        *config.Config
	Fxs           []*fxs.ParsedFx
	TypeToFxIndex map[reflect.Type]int
	Codec         codec.Manager
	// Note: FeeAssetID may be different than ctx.VidarAssetID if this AVM is
	// running in a subnet.
	FeeAssetID   ids.ID
	Bootstrapped bool
}
