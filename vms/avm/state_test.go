// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/choices"
	"github.com/VidarSolutions/avalanchego/snow/engine/common"
	"github.com/VidarSolutions/avalanchego/utils/constants"
	"github.com/VidarSolutions/avalanchego/utils/crypto/secp256k1"
	"github.com/VidarSolutions/avalanchego/utils/units"
	"github.com/VidarSolutions/avalanchego/vms/avm/txs"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
	"github.com/VidarSolutions/avalanchego/vms/secp256k1fx"
)

func TestSetsAndGets(t *testing.T) {
	_, _, vm, _ := GenesisVMWithArgs(
		t,
		[]*common.Fx{{
			ID: ids.GenerateTestID(),
			Fx: &FxTest{
				InitializeF: func(vmIntf interface{}) error {
					vm := vmIntf.(secp256k1fx.VM)
					return vm.CodecRegistry().RegisterType(&Vidar.TestVerifiable{})
				},
			},
		}},
		nil,
	)
	ctx := vm.ctx
	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx.Lock.Unlock()
	}()

	state := vm.state

	utxo := &Vidar.UTXO{
		UTXOID: Vidar.UTXOID{
			TxID:        ids.Empty,
			OutputIndex: 1,
		},
		Asset: Vidar.Asset{ID: ids.Empty},
		Out:   &Vidar.TestVerifiable{},
	}
	utxoID := utxo.InputID()

	tx := &txs.Tx{Unsigned: &txs.BaseTx{BaseTx: Vidar.BaseTx{
		NetworkID:    constants.UnitTestID,
		BlockchainID: chainID,
		Ins: []*Vidar.TransferableInput{{
			UTXOID: Vidar.UTXOID{
				TxID:        ids.Empty,
				OutputIndex: 0,
			},
			Asset: Vidar.Asset{ID: assetID},
			In: &secp256k1fx.TransferInput{
				Amt: 20 * units.KiloVidar,
				Input: secp256k1fx.Input{
					SigIndices: []uint32{
						0,
					},
				},
			},
		}},
	}}}
	if err := tx.SignSECP256K1Fx(vm.parser.Codec(), [][]*secp256k1.PrivateKey{{keys[0]}}); err != nil {
		t.Fatal(err)
	}

	txID := tx.ID()

	state.AddUTXO(utxo)
	state.AddTx(tx)
	state.AddStatus(txID, choices.Accepted)

	resultUTXO, err := state.GetUTXO(utxoID)
	if err != nil {
		t.Fatal(err)
	}
	resultTx, err := state.GetTx(txID)
	if err != nil {
		t.Fatal(err)
	}
	resultStatus, err := state.GetStatus(txID)
	if err != nil {
		t.Fatal(err)
	}

	if resultUTXO.OutputIndex != 1 {
		t.Fatalf("Wrong UTXO returned")
	}
	if resultTx.ID() != tx.ID() {
		t.Fatalf("Wrong Tx returned")
	}
	if resultStatus != choices.Accepted {
		t.Fatalf("Wrong Status returned")
	}
}

func TestFundingNoAddresses(t *testing.T) {
	_, _, vm, _ := GenesisVMWithArgs(
		t,
		[]*common.Fx{{
			ID: ids.GenerateTestID(),
			Fx: &FxTest{
				InitializeF: func(vmIntf interface{}) error {
					vm := vmIntf.(secp256k1fx.VM)
					return vm.CodecRegistry().RegisterType(&Vidar.TestVerifiable{})
				},
			},
		}},
		nil,
	)
	ctx := vm.ctx
	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx.Lock.Unlock()
	}()

	state := vm.state

	utxo := &Vidar.UTXO{
		UTXOID: Vidar.UTXOID{
			TxID:        ids.Empty,
			OutputIndex: 1,
		},
		Asset: Vidar.Asset{ID: ids.Empty},
		Out:   &Vidar.TestVerifiable{},
	}

	state.AddUTXO(utxo)
	state.DeleteUTXO(utxo.InputID())
}

func TestFundingAddresses(t *testing.T) {
	_, _, vm, _ := GenesisVMWithArgs(
		t,
		[]*common.Fx{{
			ID: ids.GenerateTestID(),
			Fx: &FxTest{
				InitializeF: func(vmIntf interface{}) error {
					vm := vmIntf.(secp256k1fx.VM)
					return vm.CodecRegistry().RegisterType(&Vidar.TestAddressable{})
				},
			},
		}},
		nil,
	)
	ctx := vm.ctx
	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx.Lock.Unlock()
	}()

	state := vm.state

	utxo := &Vidar.UTXO{
		UTXOID: Vidar.UTXOID{
			TxID:        ids.Empty,
			OutputIndex: 1,
		},
		Asset: Vidar.Asset{ID: ids.Empty},
		Out: &Vidar.TestAddressable{
			Addrs: [][]byte{{0}},
		},
	}

	state.AddUTXO(utxo)
	require.NoError(t, state.Commit())

	utxos, err := state.UTXOIDs([]byte{0}, ids.Empty, math.MaxInt32)
	require.NoError(t, err)
	require.Len(t, utxos, 1)
	require.Equal(t, utxo.InputID(), utxos[0])

	state.DeleteUTXO(utxo.InputID())
	require.NoError(t, state.Commit())

	utxos, err = state.UTXOIDs([]byte{0}, ids.Empty, math.MaxInt32)
	require.NoError(t, err)
	require.Empty(t, utxos)
}
