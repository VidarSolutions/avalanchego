// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/snow/uptime"
	"github.com/VidarSolutions/avalanchego/utils"
	"github.com/VidarSolutions/avalanchego/utils/timer/mockable"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/config"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/fx"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/reward"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/utxo"
)

type Backend struct {
	Config       *config.Config
	Ctx          *snow.Context
	Clk          *mockable.Clock
	Fx           fx.Fx
	FlowChecker  utxo.Verifier
	Uptimes      uptime.Manager
	Rewards      reward.Calculator
	Bootstrapped *utils.Atomic[bool]
}
