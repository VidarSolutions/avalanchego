// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/utils/set"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
	"github.com/VidarSolutions/avalanchego/vms/secp256k1fx"
)

var (
	_ UnsignedTx             = (*ImportTx)(nil)
	_ secp256k1fx.UnsignedTx = (*ImportTx)(nil)
)

// ImportTx is a transaction that imports an asset from another blockchain.
type ImportTx struct {
	BaseTx `serialize:"true"`

	// Which chain to consume the funds from
	SourceChain ids.ID `serialize:"true" json:"sourceChain"`

	// The inputs to this transaction
	ImportedIns []*Vidar.TransferableInput `serialize:"true" json:"importedInputs"`
}

// InputUTXOs track which UTXOs this transaction is consuming.
func (t *ImportTx) InputUTXOs() []*Vidar.UTXOID {
	utxos := t.BaseTx.InputUTXOs()
	for _, in := range t.ImportedIns {
		in.Symbol = true
		utxos = append(utxos, &in.UTXOID)
	}
	return utxos
}

func (t *ImportTx) InputIDs() set.Set[ids.ID] {
	inputs := t.BaseTx.InputIDs()
	for _, in := range t.ImportedIns {
		inputs.Add(in.InputID())
	}
	return inputs
}

// ConsumedAssetIDs returns the IDs of the assets this transaction consumes
func (t *ImportTx) ConsumedAssetIDs() set.Set[ids.ID] {
	assets := t.BaseTx.AssetIDs()
	for _, in := range t.ImportedIns {
		assets.Add(in.AssetID())
	}
	return assets
}

// AssetIDs returns the IDs of the assets this transaction depends on
func (t *ImportTx) AssetIDs() set.Set[ids.ID] {
	assets := t.BaseTx.AssetIDs()
	for _, in := range t.ImportedIns {
		assets.Add(in.AssetID())
	}
	return assets
}

// NumCredentials returns the number of expected credentials
func (t *ImportTx) NumCredentials() int {
	return t.BaseTx.NumCredentials() + len(t.ImportedIns)
}

func (t *ImportTx) Visit(v Visitor) error {
	return v.ImportTx(t)
}
