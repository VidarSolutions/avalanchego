// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/utils/set"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
	"github.com/VidarSolutions/avalanchego/vms/secp256k1fx"
)

var (
	_ UnsignedTx             = (*OperationTx)(nil)
	_ secp256k1fx.UnsignedTx = (*OperationTx)(nil)
)

// OperationTx is a transaction with no credentials.
type OperationTx struct {
	BaseTx `serialize:"true"`

	Ops []*Operation `serialize:"true" json:"operations"`
}

func (t *OperationTx) InitCtx(ctx *snow.Context) {
	for _, op := range t.Ops {
		op.Op.InitCtx(ctx)
	}
	t.BaseTx.InitCtx(ctx)
}

// Operations track which ops this transaction is performing. The returned array
// should not be modified.
func (t *OperationTx) Operations() []*Operation {
	return t.Ops
}

func (t *OperationTx) InputUTXOs() []*Vidar.UTXOID {
	utxos := t.BaseTx.InputUTXOs()
	for _, op := range t.Ops {
		utxos = append(utxos, op.UTXOIDs...)
	}
	return utxos
}

func (t *OperationTx) InputIDs() set.Set[ids.ID] {
	inputs := t.BaseTx.InputIDs()
	for _, op := range t.Ops {
		for _, utxo := range op.UTXOIDs {
			inputs.Add(utxo.InputID())
		}
	}
	return inputs
}

// ConsumedAssetIDs returns the IDs of the assets this transaction consumes
func (t *OperationTx) ConsumedAssetIDs() set.Set[ids.ID] {
	assets := t.AssetIDs()
	for _, op := range t.Ops {
		if len(op.UTXOIDs) > 0 {
			assets.Add(op.AssetID())
		}
	}
	return assets
}

// AssetIDs returns the IDs of the assets this transaction depends on
func (t *OperationTx) AssetIDs() set.Set[ids.ID] {
	assets := t.BaseTx.AssetIDs()
	for _, op := range t.Ops {
		assets.Add(op.AssetID())
	}
	return assets
}

// NumCredentials returns the number of expected credentials
func (t *OperationTx) NumCredentials() int {
	return t.BaseTx.NumCredentials() + len(t.Ops)
}

func (t *OperationTx) Visit(v Visitor) error {
	return v.OperationTx(t)
}
