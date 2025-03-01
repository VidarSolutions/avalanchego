// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"time"

	_ "embed"

	"github.com/VidarSolutions/avalanchego/utils/cb58"
	"github.com/VidarSolutions/avalanchego/utils/crypto/secp256k1"
	"github.com/VidarSolutions/avalanchego/utils/units"
	"github.com/VidarSolutions/avalanchego/utils/wrappers"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/reward"
)

// PrivateKey-vmRQiZeXEXYMyJhEiqdC2z5JhuDbxL8ix9UVvjgMu2Er1NepE => P-local1g65uqn6t77p656w64023nh8nd9updzmxyymev2
// PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN => X-local18jma8ppw3nhx5r4ap8clazz0dps7rv5u00z96u
// 56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027 => 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC

const (
	VMRQKeyStr          = "vmRQiZeXEXYMyJhEiqdC2z5JhuDbxL8ix9UVvjgMu2Er1NepE"
	VMRQKeyFormattedStr = secp256k1.PrivateKeyPrefix + VMRQKeyStr

	EWOQKeyStr          = "ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	EWOQKeyFormattedStr = secp256k1.PrivateKeyPrefix + EWOQKeyStr
)

var (
	VMRQKey *secp256k1.PrivateKey
	EWOQKey *secp256k1.PrivateKey

	//go:embed genesis_local.json
	localGenesisConfigJSON []byte

	// LocalParams are the params used for local networks
	LocalParams = Params{
		TxFeeConfig: TxFeeConfig{
			TxFee:                         units.MilliVidar,
			CreateAssetTxFee:              units.MilliVidar,
			CreateSubnetTxFee:             100 * units.MilliVidar,
			TransformSubnetTxFee:          100 * units.MilliVidar,
			CreateBlockchainTxFee:         100 * units.MilliVidar,
			AddPrimaryNetworkValidatorFee: 0,
			AddPrimaryNetworkDelegatorFee: 0,
			AddSubnetValidatorFee:         units.MilliVidar,
			AddSubnetDelegatorFee:         units.MilliVidar,
		},
		StakingConfig: StakingConfig{
			UptimeRequirement: .8, // 80%
			MinValidatorStake: 2 * units.KiloVidar,
			MaxValidatorStake: 3 * units.MegaVidar,
			MinDelegatorStake: 25 * units.Vidar,
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

func init() {
	errs := wrappers.Errs{}
	vmrqBytes, err := cb58.Decode(VMRQKeyStr)
	errs.Add(err)
	ewoqBytes, err := cb58.Decode(EWOQKeyStr)
	errs.Add(err)

	factory := secp256k1.Factory{}
	VMRQKey, err = factory.ToPrivateKey(vmrqBytes)
	errs.Add(err)
	EWOQKey, err = factory.ToPrivateKey(ewoqBytes)
	errs.Add(err)

	if errs.Err != nil {
		panic(errs.Err)
	}
}
