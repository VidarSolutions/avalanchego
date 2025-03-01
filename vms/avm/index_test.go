// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/codec"
	"github.com/VidarSolutions/avalanchego/database"
	"github.com/VidarSolutions/avalanchego/database/manager"
	"github.com/VidarSolutions/avalanchego/database/memdb"
	"github.com/VidarSolutions/avalanchego/database/prefixdb"
	"github.com/VidarSolutions/avalanchego/database/versiondb"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/snow/engine/common"
	"github.com/VidarSolutions/avalanchego/utils/constants"
	"github.com/VidarSolutions/avalanchego/utils/crypto/secp256k1"
	"github.com/VidarSolutions/avalanchego/utils/wrappers"
	"github.com/VidarSolutions/avalanchego/version"
	"github.com/VidarSolutions/avalanchego/vms/avm/txs"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
	"github.com/VidarSolutions/avalanchego/vms/components/index"
	"github.com/VidarSolutions/avalanchego/vms/secp256k1fx"
)

var indexEnabledAvmConfig = Config{
	IndexTransactions: true,
}

func TestIndexTransaction_Ordered(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)
	issuer := make(chan common.Message, 1)
	baseDBManager := manager.NewMemDB(version.Semantic1_0_0)
	ctx := NewContext(t)
	genesisTx := GetVidarTxFromGenesisTest(genesisBytes, t)
	VidarID := genesisTx.ID()
	vm := setupTestVM(t, ctx, baseDBManager, genesisBytes, issuer, indexEnabledAvmConfig)
	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx.Lock.Unlock()
	}()

	key := keys[0]
	addr := key.PublicKey().Address()

	var uniqueTxs []*UniqueTx
	txAssetID := Vidar.Asset{ID: VidarID}

	ctx.Lock.Lock()
	for i := 0; i < 5; i++ {
		// create utxoID and assetIDs
		utxoID := Vidar.UTXOID{
			TxID: ids.GenerateTestID(),
		}

		// build the transaction
		tx := buildTX(utxoID, txAssetID, addr)

		// sign the transaction
		if err := signTX(vm.parser.Codec(), tx, key); err != nil {
			t.Fatal(err)
		}

		// Provide the platform UTXO
		utxo := buildPlatformUTXO(utxoID, txAssetID, addr)

		// save utxo to state
		vm.state.AddUTXO(utxo)

		// issue transaction
		if _, err := vm.IssueTx(tx.Bytes()); err != nil {
			t.Fatalf("should have issued the transaction correctly but erred: %s", err)
		}

		ctx.Lock.Unlock()

		msg := <-issuer
		if msg != common.PendingTxs {
			t.Fatalf("Wrong message")
		}

		ctx.Lock.Lock()

		// get pending transactions
		txs := vm.PendingTxs(context.Background())
		if len(txs) != 1 {
			t.Fatalf("Should have returned %d tx(s)", 1)
		}

		parsedTx := txs[0]
		uniqueParsedTX := parsedTx.(*UniqueTx)
		uniqueTxs = append(uniqueTxs, uniqueParsedTX)

		var inputUTXOs []*Vidar.UTXO
		for _, utxoID := range uniqueParsedTX.InputUTXOs() {
			utxo, err := vm.dagState.GetUTXOFromID(utxoID)
			if err != nil {
				t.Fatal(err)
			}

			inputUTXOs = append(inputUTXOs, utxo)
		}

		// index the transaction
		err := vm.addressTxsIndexer.Accept(uniqueParsedTX.ID(), inputUTXOs, uniqueParsedTX.UTXOs())
		require.NoError(t, err)
	}

	// ensure length is 5
	require.Len(t, uniqueTxs, 5)
	// for each *UniqueTx check its indexed at right index
	for i, tx := range uniqueTxs {
		assertIndexedTX(t, vm.db, uint64(i), addr, txAssetID.ID, tx.ID())
	}

	assertLatestIdx(t, vm.db, addr, txAssetID.ID, 5)
}

func TestIndexTransaction_MultipleTransactions(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)
	issuer := make(chan common.Message, 1)
	baseDBManager := manager.NewMemDB(version.Semantic1_0_0)
	ctx := NewContext(t)
	genesisTx := GetVidarTxFromGenesisTest(genesisBytes, t)

	VidarID := genesisTx.ID()
	vm := setupTestVM(t, ctx, baseDBManager, genesisBytes, issuer, indexEnabledAvmConfig)
	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx.Lock.Unlock()
	}()

	addressTxMap := map[ids.ShortID]*UniqueTx{}
	txAssetID := Vidar.Asset{ID: VidarID}

	ctx.Lock.Lock()
	for _, key := range keys {
		addr := key.PublicKey().Address()
		// create utxoID and assetIDs
		utxoID := Vidar.UTXOID{
			TxID: ids.GenerateTestID(),
		}

		// build the transaction
		tx := buildTX(utxoID, txAssetID, addr)

		// sign the transaction
		if err := signTX(vm.parser.Codec(), tx, key); err != nil {
			t.Fatal(err)
		}

		// Provide the platform UTXO
		utxo := buildPlatformUTXO(utxoID, txAssetID, addr)

		// save utxo to state
		vm.state.AddUTXO(utxo)

		// issue transaction
		if _, err := vm.IssueTx(tx.Bytes()); err != nil {
			t.Fatalf("should have issued the transaction correctly but erred: %s", err)
		}

		ctx.Lock.Unlock()

		msg := <-issuer
		if msg != common.PendingTxs {
			t.Fatalf("Wrong message")
		}

		ctx.Lock.Lock()

		// get pending transactions
		txs := vm.PendingTxs(context.Background())
		if len(txs) != 1 {
			t.Fatalf("Should have returned %d tx(s)", 1)
		}

		parsedTx := txs[0]
		uniqueParsedTX := parsedTx.(*UniqueTx)
		addressTxMap[addr] = uniqueParsedTX

		var inputUTXOs []*Vidar.UTXO
		for _, utxoID := range uniqueParsedTX.InputUTXOs() {
			utxo, err := vm.dagState.GetUTXOFromID(utxoID)
			if err != nil {
				t.Fatal(err)
			}

			inputUTXOs = append(inputUTXOs, utxo)
		}

		// index the transaction
		err := vm.addressTxsIndexer.Accept(uniqueParsedTX.ID(), inputUTXOs, uniqueParsedTX.UTXOs())
		require.NoError(t, err)
	}

	// ensure length is same as keys length
	require.Len(t, addressTxMap, len(keys))

	// for each *UniqueTx check its indexed at right index for the right address
	for key, tx := range addressTxMap {
		assertIndexedTX(t, vm.db, uint64(0), key, txAssetID.ID, tx.ID())
		assertLatestIdx(t, vm.db, key, txAssetID.ID, 1)
	}
}

func TestIndexTransaction_MultipleAddresses(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)
	issuer := make(chan common.Message, 1)
	baseDBManager := manager.NewMemDB(version.Semantic1_0_0)
	ctx := NewContext(t)
	genesisTx := GetVidarTxFromGenesisTest(genesisBytes, t)

	VidarID := genesisTx.ID()
	vm := setupTestVM(t, ctx, baseDBManager, genesisBytes, issuer, indexEnabledAvmConfig)
	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx.Lock.Unlock()
	}()

	txAssetID := Vidar.Asset{ID: VidarID}
	addrs := make([]ids.ShortID, len(keys))
	for _, key := range keys {
		addrs = append(addrs, key.PublicKey().Address())
	}

	ctx.Lock.Lock()

	key := keys[0]
	addr := key.PublicKey().Address()
	// create utxoID and assetIDs
	utxoID := Vidar.UTXOID{
		TxID: ids.GenerateTestID(),
	}

	// build the transaction
	tx := buildTX(utxoID, txAssetID, addrs...)

	// sign the transaction
	if err := signTX(vm.parser.Codec(), tx, key); err != nil {
		t.Fatal(err)
	}

	// Provide the platform UTXO
	utxo := buildPlatformUTXO(utxoID, txAssetID, addr)

	// save utxo to state
	vm.state.AddUTXO(utxo)

	var inputUTXOs []*Vidar.UTXO //nolint:prealloc
	for _, utxoID := range tx.Unsigned.InputUTXOs() {
		utxo, err := vm.dagState.GetUTXOFromID(utxoID)
		if err != nil {
			t.Fatal(err)
		}

		inputUTXOs = append(inputUTXOs, utxo)
	}

	// index the transaction
	err := vm.addressTxsIndexer.Accept(tx.ID(), inputUTXOs, tx.UTXOs())
	require.NoError(t, err)
	require.NoError(t, err)

	assertIndexedTX(t, vm.db, uint64(0), addr, txAssetID.ID, tx.ID())
	assertLatestIdx(t, vm.db, addr, txAssetID.ID, 1)
}

func TestIndexTransaction_UnorderedWrites(t *testing.T) {
	genesisBytes := BuildGenesisTest(t)
	issuer := make(chan common.Message, 1)
	baseDBManager := manager.NewMemDB(version.Semantic1_0_0)
	ctx := NewContext(t)
	genesisTx := GetVidarTxFromGenesisTest(genesisBytes, t)
	VidarID := genesisTx.ID()
	vm := setupTestVM(t, ctx, baseDBManager, genesisBytes, issuer, indexEnabledAvmConfig)
	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		ctx.Lock.Unlock()
	}()

	addressTxMap := map[ids.ShortID]*UniqueTx{}
	txAssetID := Vidar.Asset{ID: VidarID}

	ctx.Lock.Lock()
	for _, key := range keys {
		addr := key.PublicKey().Address()
		// create utxoID and assetIDs
		utxoID := Vidar.UTXOID{
			TxID: ids.GenerateTestID(),
		}

		// build the transaction
		tx := buildTX(utxoID, txAssetID, addr)

		// sign the transaction
		if err := signTX(vm.parser.Codec(), tx, key); err != nil {
			t.Fatal(err)
		}

		// Provide the platform UTXO
		utxo := buildPlatformUTXO(utxoID, txAssetID, addr)

		// save utxo to state
		vm.state.AddUTXO(utxo)

		// issue transaction
		if _, err := vm.IssueTx(tx.Bytes()); err != nil {
			t.Fatalf("should have issued the transaction correctly but erred: %s", err)
		}

		ctx.Lock.Unlock()

		msg := <-issuer
		if msg != common.PendingTxs {
			t.Fatalf("Wrong message")
		}

		ctx.Lock.Lock()

		// get pending transactions
		txs := vm.PendingTxs(context.Background())
		if len(txs) != 1 {
			t.Fatalf("Should have returned %d tx(s)", 1)
		}

		parsedTx := txs[0]
		uniqueParsedTX := parsedTx.(*UniqueTx)
		addressTxMap[addr] = uniqueParsedTX

		var inputUTXOs []*Vidar.UTXO
		for _, utxoID := range uniqueParsedTX.InputUTXOs() {
			utxo, err := vm.dagState.GetUTXOFromID(utxoID)
			if err != nil {
				t.Fatal(err)
			}

			inputUTXOs = append(inputUTXOs, utxo)
		}

		// index the transaction, NOT calling Accept(ids.ID) method
		err := vm.addressTxsIndexer.Accept(uniqueParsedTX.ID(), inputUTXOs, uniqueParsedTX.UTXOs())
		require.NoError(t, err)
	}

	// ensure length is same as keys length
	require.Len(t, addressTxMap, len(keys))

	// for each *UniqueTx check its indexed at right index for the right address
	for key, tx := range addressTxMap {
		assertIndexedTX(t, vm.db, uint64(0), key, txAssetID.ID, tx.ID())
		assertLatestIdx(t, vm.db, key, txAssetID.ID, 1)
	}
}

func TestIndexer_Read(t *testing.T) {
	// setup vm, db etc
	_, vm, _, _, _ := setup(t, true)

	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
		vm.ctx.Lock.Unlock()
	}()

	// generate test address and asset IDs
	assetID := ids.GenerateTestID()
	addr := ids.GenerateTestShortID()

	// setup some fake txs under the above generated address and asset IDs
	testTxCount := 25
	testTxs := setupTestTxsInDB(t, vm.db, addr, assetID, testTxCount)
	require.Len(t, testTxs, 25)

	// read the pages, 5 items at a time
	var cursor uint64
	var pageSize uint64 = 5
	for cursor < 25 {
		txIDs, err := vm.addressTxsIndexer.Read(addr[:], assetID, cursor, pageSize)
		require.NoError(t, err)
		require.Len(t, txIDs, 5)
		require.Equal(t, txIDs, testTxs[cursor:cursor+pageSize])
		cursor += pageSize
	}
}

func TestIndexingNewInitWithIndexingEnabled(t *testing.T) {
	baseDBManager := manager.NewMemDB(version.Semantic1_0_0)
	ctx := NewContext(t)

	db := baseDBManager.NewPrefixDBManager([]byte{1}).Current().Database

	// start with indexing enabled
	_, err := index.NewIndexer(db, ctx.Log, "", prometheus.NewRegistry(), true)
	require.NoError(t, err)

	// now disable indexing with allow-incomplete set to false
	_, err = index.NewNoIndexer(db, false)
	require.Error(t, err)

	// now disable indexing with allow-incomplete set to true
	_, err = index.NewNoIndexer(db, true)
	require.NoError(t, err)
}

func TestIndexingNewInitWithIndexingDisabled(t *testing.T) {
	ctx := NewContext(t)
	db := memdb.New()

	// disable indexing with allow-incomplete set to false
	_, err := index.NewNoIndexer(db, false)
	require.NoError(t, err)

	// It's not OK to have an incomplete index when allowIncompleteIndices is false
	_, err = index.NewIndexer(db, ctx.Log, "", prometheus.NewRegistry(), false)
	require.Error(t, err)

	// It's OK to have an incomplete index when allowIncompleteIndices is true
	_, err = index.NewIndexer(db, ctx.Log, "", prometheus.NewRegistry(), true)
	require.NoError(t, err)

	// It's OK to have an incomplete index when indexing currently disabled
	_, err = index.NewNoIndexer(db, false)
	require.NoError(t, err)

	// It's OK to have an incomplete index when allowIncompleteIndices is true
	_, err = index.NewNoIndexer(db, true)
	require.NoError(t, err)
}

func TestIndexingAllowIncomplete(t *testing.T) {
	baseDBManager := manager.NewMemDB(version.Semantic1_0_0)
	ctx := NewContext(t)

	prefixDB := baseDBManager.NewPrefixDBManager([]byte{1}).Current().Database
	db := versiondb.New(prefixDB)
	// disabled indexer will persist idxEnabled as false
	_, err := index.NewNoIndexer(db, false)
	require.NoError(t, err)

	// we initialise with indexing enabled now and allow incomplete indexing as false
	_, err = index.NewIndexer(db, ctx.Log, "", prometheus.NewRegistry(), false)
	// we should get error because:
	// - indexing was disabled previously
	// - node now is asked to enable indexing with allow incomplete set to false
	require.Error(t, err)
}

func buildPlatformUTXO(utxoID Vidar.UTXOID, txAssetID Vidar.Asset, addr ids.ShortID) *Vidar.UTXO {
	return &Vidar.UTXO{
		UTXOID: utxoID,
		Asset:  txAssetID,
		Out: &secp256k1fx.TransferOutput{
			Amt: 1000,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr},
			},
		},
	}
}

func signTX(codec codec.Manager, tx *txs.Tx, key *secp256k1.PrivateKey) error {
	return tx.SignSECP256K1Fx(codec, [][]*secp256k1.PrivateKey{{key}})
}

func buildTX(utxoID Vidar.UTXOID, txAssetID Vidar.Asset, address ...ids.ShortID) *txs.Tx {
	return &txs.Tx{Unsigned: &txs.BaseTx{
		BaseTx: Vidar.BaseTx{
			NetworkID:    constants.UnitTestID,
			BlockchainID: chainID,
			Ins: []*Vidar.TransferableInput{{
				UTXOID: utxoID,
				Asset:  txAssetID,
				In: &secp256k1fx.TransferInput{
					Amt:   1000,
					Input: secp256k1fx.Input{SigIndices: []uint32{0}},
				},
			}},
			Outs: []*Vidar.TransferableOutput{{
				Asset: txAssetID,
				Out: &secp256k1fx.TransferOutput{
					Amt: 1000,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     address,
					},
				},
			}},
		},
	}}
}

func setupTestVM(t *testing.T, ctx *snow.Context, baseDBManager manager.Manager, genesisBytes []byte, issuer chan common.Message, config Config) *VM {
	vm := &VM{}
	avmConfigBytes, err := json.Marshal(config)
	require.NoError(t, err)
	appSender := &common.SenderTest{T: t}

	err = vm.Initialize(
		context.Background(),
		ctx,
		baseDBManager.NewPrefixDBManager([]byte{1}),
		genesisBytes,
		nil,
		avmConfigBytes,
		issuer,
		[]*common.Fx{{
			ID: ids.Empty,
			Fx: &secp256k1fx.Fx{},
		}},
		appSender,
	)
	if err != nil {
		t.Fatal(err)
	}

	vm.batchTimeout = 0

	if err := vm.SetState(context.Background(), snow.Bootstrapping); err != nil {
		t.Fatal(err)
	}

	if err := vm.SetState(context.Background(), snow.NormalOp); err != nil {
		t.Fatal(err)
	}
	return vm
}

func assertLatestIdx(t *testing.T, db database.Database, sourceAddress ids.ShortID, assetID ids.ID, expectedIdx uint64) {
	addressDB := prefixdb.New(sourceAddress[:], db)
	assetDB := prefixdb.New(assetID[:], addressDB)

	expectedIdxBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(expectedIdxBytes, expectedIdx)

	idxBytes, err := assetDB.Get([]byte("idx"))
	require.NoError(t, err)

	require.EqualValues(t, expectedIdxBytes, idxBytes)
}

func checkIndexedTX(db database.Database, index uint64, sourceAddress ids.ShortID, assetID ids.ID, transactionID ids.ID) error {
	addressDB := prefixdb.New(sourceAddress[:], db)
	assetDB := prefixdb.New(assetID[:], addressDB)

	idxBytes := make([]byte, wrappers.LongLen)
	binary.BigEndian.PutUint64(idxBytes, index)
	tx1Bytes, err := assetDB.Get(idxBytes)
	if err != nil {
		return err
	}

	var txID ids.ID
	copy(txID[:], tx1Bytes)

	if txID != transactionID {
		return fmt.Errorf("txID %s not same as %s", txID, transactionID)
	}
	return nil
}

func assertIndexedTX(t *testing.T, db database.Database, index uint64, sourceAddress ids.ShortID, assetID ids.ID, transactionID ids.ID) {
	if err := checkIndexedTX(db, index, sourceAddress, assetID, transactionID); err != nil {
		t.Fatal(err)
	}
}

// Sets up test tx IDs in DB in the following structure for the indexer to pick
// them up:
//
//	[address] prefix DB
//	  [assetID] prefix DB
//	    - "idx": 2
//	    - 0: txID1
//	    - 1: txID1
func setupTestTxsInDB(t *testing.T, db *versiondb.Database, address ids.ShortID, assetID ids.ID, txCount int) []ids.ID {
	var testTxs []ids.ID
	for i := 0; i < txCount; i++ {
		testTxs = append(testTxs, ids.GenerateTestID())
	}

	addressPrefixDB := prefixdb.New(address[:], db)
	assetPrefixDB := prefixdb.New(assetID[:], addressPrefixDB)
	var idx uint64
	idxBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idxBytes, idx)
	for _, txID := range testTxs {
		txID := txID
		err := assetPrefixDB.Put(idxBytes, txID[:])
		require.NoError(t, err)
		idx++
		binary.BigEndian.PutUint64(idxBytes, idx)
	}
	_, err := db.CommitBatch()
	require.NoError(t, err)

	err = assetPrefixDB.Put([]byte("idx"), idxBytes)
	require.NoError(t, err)
	err = db.Commit()
	require.NoError(t, err)
	return testTxs
}
