// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utxo

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/utils/crypto/secp256k1"
	"github.com/VidarSolutions/avalanchego/utils/timer/mockable"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
	"github.com/VidarSolutions/avalanchego/vms/components/verify"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/stakeable"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/txs"
	"github.com/VidarSolutions/avalanchego/vms/secp256k1fx"
)

var _ txs.UnsignedTx = (*dummyUnsignedTx)(nil)

type dummyUnsignedTx struct {
	txs.BaseTx
}

func (*dummyUnsignedTx) Visit(txs.Visitor) error {
	return nil
}

func TestVerifySpendUTXOs(t *testing.T) {
	fx := &secp256k1fx.Fx{}

	require.NoError(t, fx.InitializeVM(&secp256k1fx.TestVM{}))
	require.NoError(t, fx.Bootstrapped())

	h := &handler{
		ctx: snow.DefaultContextTest(),
		clk: &mockable.Clock{},
		fx:  fx,
	}

	// The handler time during a test, unless [chainTimestamp] is set
	now := time.Unix(1607133207, 0)

	unsignedTx := dummyUnsignedTx{
		BaseTx: txs.BaseTx{},
	}
	unsignedTx.SetBytes([]byte{0})

	customAssetID := ids.GenerateTestID()

	// Note that setting [chainTimestamp] also set's the handler's clock.
	// Adjust input/output locktimes accordingly.
	tests := []struct {
		description     string
		utxos           []*Vidar.UTXO
		ins             []*Vidar.TransferableInput
		outs            []*Vidar.TransferableOutput
		creds           []verify.Verifiable
		producedAmounts map[ids.ID]uint64
		shouldErr       bool
	}{
		{
			description:     "no inputs, no outputs, no fee",
			utxos:           []*Vidar.UTXO{},
			ins:             []*Vidar.TransferableInput{},
			outs:            []*Vidar.TransferableOutput{},
			creds:           []verify.Verifiable{},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       false,
		},
		{
			description: "no inputs, no outputs, positive fee",
			utxos:       []*Vidar.UTXO{},
			ins:         []*Vidar.TransferableInput{},
			outs:        []*Vidar.TransferableOutput{},
			creds:       []verify.Verifiable{},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "wrong utxo assetID, one input, no outputs, no fee",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: customAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "one wrong assetID input, no outputs, no fee",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: customAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "one input, one wrong assetID output, no fee",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "attempt to consume locked output as unlocked",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Add(time.Second).Unix()),
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "attempt to modify locktime",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Add(time.Second).Unix()),
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &stakeable.LockIn{
					Locktime: uint64(now.Unix()),
					TransferableIn: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "one input, no outputs, positive fee",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: false,
		},
		{
			description: "wrong number of credentials",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs:  []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "wrong number of UTXOs",
			utxos:       []*Vidar.UTXO{},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "invalid credential",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				(*secp256k1fx.Credential)(nil),
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "invalid signature",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs: []ids.ShortID{
							ids.GenerateTestShortID(),
						},
					},
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
					Input: secp256k1fx.Input{
						SigIndices: []uint32{0},
					},
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{
					Sigs: [][secp256k1.SignatureLen]byte{
						{},
					},
				},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "one input, no outputs, positive fee",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: false,
		},
		{
			description: "locked one input, no outputs, no fee",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Unix()) + 1,
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &stakeable.LockIn{
					Locktime: uint64(now.Unix()) + 1,
					TransferableIn: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       false,
		},
		{
			description: "locked one input, no outputs, positive fee",
			utxos: []*Vidar.UTXO{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Unix()) + 1,
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*Vidar.TransferableInput{{
				Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
				In: &stakeable.LockIn{
					Locktime: uint64(now.Unix()) + 1,
					TransferableIn: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			}},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "one locked and one unlocked input, one locked output, positive fee",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &stakeable.LockIn{
						Locktime: uint64(now.Unix()) + 1,
						TransferableIn: &secp256k1fx.TransferInput{
							Amt: 1,
						},
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: false,
		},
		{
			description: "one locked and one unlocked input, one locked output, positive fee, partially locked",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &stakeable.LockIn{
						Locktime: uint64(now.Unix()) + 1,
						TransferableIn: &secp256k1fx.TransferInput{
							Amt: 1,
						},
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 2,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: false,
		},
		{
			description: "one unlocked input, one locked output, zero fee",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) - 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       false,
		},
		{
			description: "attempted overflow",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2,
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: math.MaxUint64,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "attempted mint",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "attempted mint through locking",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: math.MaxUint64,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "attempted mint through mixed locking (low then high)",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2,
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: math.MaxUint64,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "attempted mint through mixed locking (high then low)",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: math.MaxUint64,
					},
				},
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "transfer non-Vidar asset",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       false,
		},
		{
			description: "lock non-Vidar asset",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Add(time.Second).Unix()),
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       false,
		},
		{
			description: "attempted asset conversion",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			shouldErr:       true,
		},
		{
			description: "attempted asset conversion with burn",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "two inputs, one output with custom asset, with fee",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: false,
		},
		{
			description: "one input, fee, custom asset",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "one input, custom fee",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				customAssetID: 1,
			},
			shouldErr: false,
		},
		{
			description: "one input, custom fee, wrong burn",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				customAssetID: 1,
			},
			shouldErr: true,
		},
		{
			description: "two inputs, multiple fee",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: h.ctx.VidarAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.VidarAssetID: 1,
				customAssetID:     1,
			},
			shouldErr: false,
		},
		{
			description: "one unlock input, one locked output, zero fee, unlocked, custom asset",
			utxos: []*Vidar.UTXO{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) - 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			ins: []*Vidar.TransferableInput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*Vidar.TransferableOutput{
				{
					Asset: Vidar.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: make(map[ids.ID]uint64),
			shouldErr:       false,
		},
	}

	for _, test := range tests {
		h.clk.Set(now)

		t.Run(test.description, func(t *testing.T) {
			require := require.New(t)
			err := h.VerifySpendUTXOs(
				&unsignedTx,
				test.utxos,
				test.ins,
				test.outs,
				test.creds,
				test.producedAmounts,
			)

			if test.shouldErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}
