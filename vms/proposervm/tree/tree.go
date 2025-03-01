// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package tree

import (
	"context"

	"golang.org/x/exp/maps"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/consensus/snowman"
)

type Tree interface {
	// Add places the block in the tree
	Add(snowman.Block)

	// Get returns the block that was added to this tree whose parent and ID
	// match the provided block. If non-exists, then false will be returned.
	Get(snowman.Block) (snowman.Block, bool)

	// Accept marks the provided block as accepted and rejects every conflicting
	// block.
	Accept(context.Context, snowman.Block) error
}

type tree struct {
	// parentID -> childID -> childBlock
	nodes map[ids.ID]map[ids.ID]snowman.Block
}

func New() Tree {
	return &tree{
		nodes: make(map[ids.ID]map[ids.ID]snowman.Block),
	}
}

func (t *tree) Add(blk snowman.Block) {
	parentID := blk.Parent()
	children, exists := t.nodes[parentID]
	if !exists {
		children = make(map[ids.ID]snowman.Block)
		t.nodes[parentID] = children
	}
	blkID := blk.ID()
	children[blkID] = blk
}

func (t *tree) Get(blk snowman.Block) (snowman.Block, bool) {
	parentID := blk.Parent()
	children := t.nodes[parentID]
	blkID := blk.ID()
	originalBlk, exists := children[blkID]
	return originalBlk, exists
}

func (t *tree) Accept(ctx context.Context, blk snowman.Block) error {
	// accept the provided block
	if err := blk.Accept(ctx); err != nil {
		return err
	}

	// get the siblings of the block
	parentID := blk.Parent()
	children := t.nodes[parentID]
	blkID := blk.ID()
	delete(children, blkID)
	delete(t.nodes, parentID)

	// mark the siblings of the accepted block as rejectable
	childrenToReject := maps.Values(children)

	// reject all the rejectable blocks
	for len(childrenToReject) > 0 {
		i := len(childrenToReject) - 1
		child := childrenToReject[i]
		childrenToReject = childrenToReject[:i]

		// reject the block
		if err := child.Reject(ctx); err != nil {
			return err
		}

		// mark the progeny of this block as being rejectable
		blkID := child.ID()
		children := t.nodes[blkID]
		childrenToReject = append(childrenToReject, maps.Values(children)...)
		delete(t.nodes, blkID)
	}
	return nil
}
