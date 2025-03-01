// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package states

import (
	"errors"
	"fmt"
	"time"

	"github.com/VidarSolutions/avalanchego/database"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/vms/avm/blocks"
	"github.com/VidarSolutions/avalanchego/vms/avm/txs"
	"github.com/VidarSolutions/avalanchego/vms/components/Vidar"
)

var (
	_ Diff = (*diff)(nil)

	ErrMissingParentState = errors.New("missing parent state")
)

type Diff interface {
	Chain

	Apply(Chain)
}

type diff struct {
	parentID      ids.ID
	stateVersions Versions

	// map of modified UTXOID -> *UTXO if the UTXO is nil, it has been removed
	modifiedUTXOs map[ids.ID]*Vidar.UTXO
	addedTxs      map[ids.ID]*txs.Tx      // map of txID -> tx
	addedBlockIDs map[uint64]ids.ID       // map of height -> blockID
	addedBlocks   map[ids.ID]blocks.Block // map of blockID -> block

	lastAccepted ids.ID
	timestamp    time.Time
}

func NewDiff(
	parentID ids.ID,
	stateVersions Versions,
) (Diff, error) {
	parentState, ok := stateVersions.GetState(parentID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingParentState, parentID)
	}
	return &diff{
		parentID:      parentID,
		stateVersions: stateVersions,
		modifiedUTXOs: make(map[ids.ID]*Vidar.UTXO),
		addedTxs:      make(map[ids.ID]*txs.Tx),
		addedBlockIDs: make(map[uint64]ids.ID),
		addedBlocks:   make(map[ids.ID]blocks.Block),
		lastAccepted:  parentState.GetLastAccepted(),
		timestamp:     parentState.GetTimestamp(),
	}, nil
}

func (d *diff) GetUTXO(utxoID ids.ID) (*Vidar.UTXO, error) {
	if utxo, modified := d.modifiedUTXOs[utxoID]; modified {
		if utxo == nil {
			return nil, database.ErrNotFound
		}
		return utxo, nil
	}

	parentState, ok := d.stateVersions.GetState(d.parentID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingParentState, d.parentID)
	}
	return parentState.GetUTXO(utxoID)
}

func (d *diff) GetUTXOFromID(utxoID *Vidar.UTXOID) (*Vidar.UTXO, error) {
	return d.GetUTXO(utxoID.InputID())
}

func (d *diff) AddUTXO(utxo *Vidar.UTXO) {
	d.modifiedUTXOs[utxo.InputID()] = utxo
}

func (d *diff) DeleteUTXO(utxoID ids.ID) {
	d.modifiedUTXOs[utxoID] = nil
}

func (d *diff) GetTx(txID ids.ID) (*txs.Tx, error) {
	if tx, exists := d.addedTxs[txID]; exists {
		return tx, nil
	}

	parentState, ok := d.stateVersions.GetState(d.parentID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingParentState, d.parentID)
	}
	return parentState.GetTx(txID)
}

func (d *diff) AddTx(tx *txs.Tx) {
	d.addedTxs[tx.ID()] = tx
}

func (d *diff) GetBlockID(height uint64) (ids.ID, error) {
	if blkID, exists := d.addedBlockIDs[height]; exists {
		return blkID, nil
	}

	parentState, ok := d.stateVersions.GetState(d.parentID)
	if !ok {
		return ids.Empty, fmt.Errorf("%w: %s", ErrMissingParentState, d.parentID)
	}
	return parentState.GetBlockID(height)
}

func (d *diff) GetBlock(blkID ids.ID) (blocks.Block, error) {
	if blk, exists := d.addedBlocks[blkID]; exists {
		return blk, nil
	}

	parentState, ok := d.stateVersions.GetState(d.parentID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingParentState, d.parentID)
	}
	return parentState.GetBlock(blkID)
}

func (d *diff) AddBlock(blk blocks.Block) {
	blkID := blk.ID()
	d.addedBlockIDs[blk.Height()] = blkID
	d.addedBlocks[blkID] = blk
}

func (d *diff) GetLastAccepted() ids.ID {
	return d.lastAccepted
}

func (d *diff) SetLastAccepted(lastAccepted ids.ID) {
	d.lastAccepted = lastAccepted
}

func (d *diff) GetTimestamp() time.Time {
	return d.timestamp
}

func (d *diff) SetTimestamp(t time.Time) {
	d.timestamp = t
}

func (d *diff) Apply(state Chain) {
	for utxoID, utxo := range d.modifiedUTXOs {
		if utxo != nil {
			state.AddUTXO(utxo)
		} else {
			state.DeleteUTXO(utxoID)
		}
	}

	for _, tx := range d.addedTxs {
		state.AddTx(tx)
	}

	for _, blk := range d.addedBlocks {
		state.AddBlock(blk)
	}

	state.SetLastAccepted(d.lastAccepted)
	state.SetTimestamp(d.timestamp)
}
