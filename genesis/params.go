// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"time"

	"github.com/VidarSolutions/avalanchego/utils/constants"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/reward"
)

type StakingConfig struct {
	// Staking uptime requirements
	UptimeRequirement float64 `json:"uptimeRequirement"`
	// Minimum stake, in nVidar, required to validate the primary network
	MinValidatorStake uint64 `json:"minValidatorStake"`
	// Maximum stake, in nVidar, allowed to be placed on a single validator in
	// the primary network
	MaxValidatorStake uint64 `json:"maxValidatorStake"`
	// Minimum stake, in nVidar, that can be delegated on the primary network
	MinDelegatorStake uint64 `json:"minDelegatorStake"`
	// Minimum delegation fee, in the range [0, 1000000], that can be charged
	// for delegation on the primary network.
	MinDelegationFee uint32 `json:"minDelegationFee"`
	// MinStakeDuration is the minimum amount of time a validator can validate
	// for in a single period.
	MinStakeDuration time.Duration `json:"minStakeDuration"`
	// MaxStakeDuration is the maximum amount of time a validator can validate
	// for in a single period.
	MaxStakeDuration time.Duration `json:"maxStakeDuration"`
	// RewardConfig is the config for the reward function.
	RewardConfig reward.Config `json:"rewardConfig"`
}

type TxFeeConfig struct {
	// Transaction fee
	TxFee uint64 `json:"txFee"`
	// Transaction fee for create asset transactions
	CreateAssetTxFee uint64 `json:"createAssetTxFee"`
	// Transaction fee for create subnet transactions
	CreateSubnetTxFee uint64 `json:"createSubnetTxFee"`
	// Transaction fee for transform subnet transactions
	TransformSubnetTxFee uint64 `json:"transformSubnetTxFee"`
	// Transaction fee for create blockchain transactions
	CreateBlockchainTxFee uint64 `json:"createBlockchainTxFee"`
	// Transaction fee for adding a primary network validator
	AddPrimaryNetworkValidatorFee uint64 `json:"addPrimaryNetworkValidatorFee"`
	// Transaction fee for adding a primary network delegator
	AddPrimaryNetworkDelegatorFee uint64 `json:"addPrimaryNetworkDelegatorFee"`
	// Transaction fee for adding a subnet validator
	AddSubnetValidatorFee uint64 `json:"addSubnetValidatorFee"`
	// Transaction fee for adding a subnet delegator
	AddSubnetDelegatorFee uint64 `json:"addSubnetDelegatorFee"`
}

type Params struct {
	StakingConfig
	TxFeeConfig
}

func GetTxFeeConfig(networkID uint32) TxFeeConfig {
	switch networkID {
	case constants.MainnetID:
		return MainnetParams.TxFeeConfig
	case constants.FujiID:
		return FujiParams.TxFeeConfig
	case constants.LocalID:
		return LocalParams.TxFeeConfig
	default:
		return LocalParams.TxFeeConfig
	}
}

func GetStakingConfig(networkID uint32) StakingConfig {
	switch networkID {
	case constants.MainnetID:
		return MainnetParams.StakingConfig
	case constants.FujiID:
		return FujiParams.StakingConfig
	case constants.LocalID:
		return LocalParams.StakingConfig
	default:
		return LocalParams.StakingConfig
	}
}
