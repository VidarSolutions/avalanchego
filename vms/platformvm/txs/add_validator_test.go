// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/utils/crypto/secp256k1"
	"github.com/VidarSolutions/avalanchego/utils/timer/mockable"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/reward"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/stakeable"
	"github.com/VidarSolutions/avalanchego/vms/secp256k1fx"
)

func TestAddValidatorTxSyntacticVerify(t *testing.T) {
	require := require.New(t)
	clk := mockable.Clock{}
	ctx := snow.DefaultContextTest()
	ctx.VidarAssetID = ids.GenerateTestID()
	signers := [][]*secp256k1.PrivateKey{preFundedKeys}

	var (
		stx            *Tx
		addValidatorTx *AddValidatorTx
		err            error
	)

	// Case : signed tx is nil
	require.ErrorIs(stx.SyntacticVerify(ctx), ErrNilSignedTx)

	// Case : unsigned tx is nil
	require.ErrorIs(addValidatorTx.SyntacticVerify(ctx), ErrNilTx)

	validatorWeight := uint64(2022)
	rewardAddress := preFundedKeys[0].PublicKey().Address()
	inputs := []*Vidar.TransferableInput{{
		UTXOID: Vidar.UTXOID{
			TxID:        ids.ID{'t', 'x', 'I', 'D'},
			OutputIndex: 2,
		},
		Asset: Vidar.Asset{ID: ctx.VidarAssetID},
		In: &secp256k1fx.TransferInput{
			Amt:   uint64(5678),
			Input: secp256k1fx.Input{SigIndices: []uint32{0}},
		},
	}}
	outputs := []*Vidar.TransferableOutput{{
		Asset: Vidar.Asset{ID: ctx.VidarAssetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: uint64(1234),
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
			},
		},
	}}
	stakes := []*Vidar.TransferableOutput{{
		Asset: Vidar.Asset{ID: ctx.VidarAssetID},
		Out: &stakeable.LockOut{
			Locktime: uint64(clk.Time().Add(time.Second).Unix()),
			TransferableOut: &secp256k1fx.TransferOutput{
				Amt: validatorWeight,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
				},
			},
		},
	}}
	addValidatorTx = &AddValidatorTx{
		BaseTx: BaseTx{BaseTx: Vidar.BaseTx{
			NetworkID:    ctx.NetworkID,
			BlockchainID: ctx.ChainID,
			Ins:          inputs,
			Outs:         outputs,
		}},
		Validator: Validator{
			NodeID: ctx.NodeID,
			Start:  uint64(clk.Time().Unix()),
			End:    uint64(clk.Time().Add(time.Hour).Unix()),
			Wght:   validatorWeight,
		},
		StakeOuts: stakes,
		RewardsOwner: &secp256k1fx.OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs:     []ids.ShortID{rewardAddress},
		},
		DelegationShares: reward.PercentDenominator,
	}

	// Case: valid tx
	stx, err = NewSigned(addValidatorTx, Codec, signers)
	require.NoError(err)
	require.NoError(stx.SyntacticVerify(ctx))

	// Case: Wrong network ID
	addValidatorTx.SyntacticallyVerified = false
	addValidatorTx.NetworkID++
	stx, err = NewSigned(addValidatorTx, Codec, signers)
	require.NoError(err)
	err = stx.SyntacticVerify(ctx)
	require.Error(err)
	addValidatorTx.NetworkID--

	// Case: Stake owner has no addresses
	addValidatorTx.SyntacticallyVerified = false
	addValidatorTx.StakeOuts[0].
		Out.(*stakeable.LockOut).
		TransferableOut.(*secp256k1fx.TransferOutput).
		Addrs = nil
	stx, err = NewSigned(addValidatorTx, Codec, signers)
	require.NoError(err)
	err = stx.SyntacticVerify(ctx)
	require.Error(err)
	addValidatorTx.StakeOuts = stakes

	// Case: Rewards owner has no addresses
	addValidatorTx.SyntacticallyVerified = false
	addValidatorTx.RewardsOwner.(*secp256k1fx.OutputOwners).Addrs = nil
	stx, err = NewSigned(addValidatorTx, Codec, signers)
	require.NoError(err)
	err = stx.SyntacticVerify(ctx)
	require.Error(err)
	addValidatorTx.RewardsOwner.(*secp256k1fx.OutputOwners).Addrs = []ids.ShortID{rewardAddress}

	// Case: Too many shares
	addValidatorTx.SyntacticallyVerified = false
	addValidatorTx.DelegationShares++ // 1 more than max amount
	stx, err = NewSigned(addValidatorTx, Codec, signers)
	require.NoError(err)
	err = stx.SyntacticVerify(ctx)
	require.Error(err)
	addValidatorTx.DelegationShares--
}

func TestAddValidatorTxSyntacticVerifyNotVidar(t *testing.T) {
	require := require.New(t)
	clk := mockable.Clock{}
	ctx := snow.DefaultContextTest()
	ctx.VidarAssetID = ids.GenerateTestID()
	signers := [][]*secp256k1.PrivateKey{preFundedKeys}

	var (
		stx            *Tx
		addValidatorTx *AddValidatorTx
		err            error
	)

	assetID := ids.GenerateTestID()
	validatorWeight := uint64(2022)
	rewardAddress := preFundedKeys[0].PublicKey().Address()
	inputs := []*Vidar.TransferableInput{{
		UTXOID: Vidar.UTXOID{
			TxID:        ids.ID{'t', 'x', 'I', 'D'},
			OutputIndex: 2,
		},
		Asset: Vidar.Asset{ID: assetID},
		In: &secp256k1fx.TransferInput{
			Amt:   uint64(5678),
			Input: secp256k1fx.Input{SigIndices: []uint32{0}},
		},
	}}
	outputs := []*Vidar.TransferableOutput{{
		Asset: Vidar.Asset{ID: assetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: uint64(1234),
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
			},
		},
	}}
	stakes := []*Vidar.TransferableOutput{{
		Asset: Vidar.Asset{ID: assetID},
		Out: &stakeable.LockOut{
			Locktime: uint64(clk.Time().Add(time.Second).Unix()),
			TransferableOut: &secp256k1fx.TransferOutput{
				Amt: validatorWeight,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{preFundedKeys[0].PublicKey().Address()},
				},
			},
		},
	}}
	addValidatorTx = &AddValidatorTx{
		BaseTx: BaseTx{BaseTx: Vidar.BaseTx{
			NetworkID:    ctx.NetworkID,
			BlockchainID: ctx.ChainID,
			Ins:          inputs,
			Outs:         outputs,
		}},
		Validator: Validator{
			NodeID: ctx.NodeID,
			Start:  uint64(clk.Time().Unix()),
			End:    uint64(clk.Time().Add(time.Hour).Unix()),
			Wght:   validatorWeight,
		},
		StakeOuts: stakes,
		RewardsOwner: &secp256k1fx.OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs:     []ids.ShortID{rewardAddress},
		},
		DelegationShares: reward.PercentDenominator,
	}

	stx, err = NewSigned(addValidatorTx, Codec, signers)
	require.NoError(err)
	require.Error(stx.SyntacticVerify(ctx))
}

func TestAddValidatorTxNotDelegatorTx(t *testing.T) {
	txIntf := any((*AddValidatorTx)(nil))
	_, ok := txIntf.(DelegatorTx)
	require.False(t, ok)
}
