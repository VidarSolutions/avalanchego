// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
)

func TestBaseTxMarshalJSON(t *testing.T) {
	blockchainID := ids.ID{1}
	utxoTxID := ids.ID{2}
	assetID := ids.ID{3}
	fxID := ids.ID{4}
	tx := &BaseTx{BaseTx: Vidar.BaseTx{
		BlockchainID: blockchainID,
		NetworkID:    4,
		Ins: []*Vidar.TransferableInput{
			{
				FxID:   fxID,
				UTXOID: Vidar.UTXOID{TxID: utxoTxID, OutputIndex: 5},
				Asset:  Vidar.Asset{ID: assetID},
				In:     &Vidar.TestTransferable{Val: 100},
			},
		},
		Outs: []*Vidar.TransferableOutput{
			{
				FxID:  fxID,
				Asset: Vidar.Asset{ID: assetID},
				Out:   &Vidar.TestTransferable{Val: 100},
			},
		},
		Memo: []byte{1, 2, 3},
	}}

	txBytes, err := json.Marshal(tx)
	require.NoError(t, err)

	asString := string(txBytes)

	require.Contains(t, asString, `"networkID":4`)
	require.Contains(t, asString, `"blockchainID":"SYXsAycDPUu4z2ZksJD5fh5nTDcH3vCFHnpcVye5XuJ2jArg"`)
	require.Contains(t, asString, `"inputs":[{"txID":"t64jLxDRmxo8y48WjbRALPAZuSDZ6qPVaaeDzxHA4oSojhLt","outputIndex":5,"assetID":"2KdbbWvpeAShCx5hGbtdF15FMMepq9kajsNTqVvvEbhiCRSxU","fxID":"2mB8TguRrYvbGw7G2UBqKfmL8osS7CfmzAAHSzuZK8bwpRKdY","input":{"Err":null,"Val":100}}]`)
	require.Contains(t, asString, `"outputs":[{"assetID":"2KdbbWvpeAShCx5hGbtdF15FMMepq9kajsNTqVvvEbhiCRSxU","fxID":"2mB8TguRrYvbGw7G2UBqKfmL8osS7CfmzAAHSzuZK8bwpRKdY","output":{"Err":null,"Val":100}}]`)
}
