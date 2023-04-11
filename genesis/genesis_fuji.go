// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"time"

	_ "embed"

	"github.com/VidarSolutions/avalanchego/utils/units"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/reward"
)

var (
	//go:embed genesis_fuji.json
	fujiGenesisConfigJSON []byte

	// FujiParams are the params used for the fuji testnet
	FujiParams = Params{
		TxFeeConfig: TxFeeConfig{
			TxFee:                         units.MilliVidar,
			CreateAssetTxFee:              10 * units.MilliVidar,
			CreateSubnetTxFee:             100 * units.MilliVidar,
			TransformSubnetTxFee:          1 * units.Vidar,
			CreateBlockchainTxFee:         100 * units.MilliVidar,
			AddPrimaryNetworkValidatorFee: 0,
			AddPrimaryNetworkDelegatorFee: 0,
			AddSubnetValidatorFee:         units.MilliVidar,
			AddSubnetDelegatorFee:         units.MilliVidar,
		},
		StakingConfig: StakingConfig{
			UptimeRequirement: .8, // 80%
			MinValidatorStake: 1 * units.Vidar,
			MaxValidatorStake: 3 * units.MegaVidar,
			MinDelegatorStake: 1 * units.Vidar,
			MinDelegationFee:  20000, // 2%
			MinStakeDuration:  24 * time.Hour,
			MaxStakeDuration:  365 * 24 * time.Hour,
			RewardConfig: reward.Config{
				MaxConsumptionRate: .12 * reward.PercentDenominator,
				MinConsumptionRate: .10 * reward.PercentDenominator,
				MintingPeriod:      365 * 24 * time.Hour,
				SupplyCap:          720 * units.MegaVidar,
			},
		},
	}
)
