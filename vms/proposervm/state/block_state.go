// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/VidarSolutions/avalanchego/cache"
	"github.com/VidarSolutions/avalanchego/cache/metercacher"
	"github.com/VidarSolutions/avalanchego/database"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/choices"
	"github.com/VidarSolutions/avalanchego/vms/proposervm/block"
)

const blockCacheSize = 8192

var (
	errBlockWrongVersion = errors.New("wrong version")

	_ BlockState = (*blockState)(nil)
)

type BlockState interface {
	GetBlock(blkID ids.ID) (block.Block, choices.Status, error)
	PutBlock(blk block.Block, status choices.Status) error
}

type blockState struct {
	// Caches BlockID -> Block. If the Block is nil, that means the block is not
	// in storage.
	blkCache cache.Cacher[ids.ID, *blockWrapper]

	db database.Database
}

type blockWrapper struct {
	Block  []byte         `serialize:"true"`
	Status choices.Status `serialize:"true"`

	block block.Block
}

func NewBlockState(db database.Database) BlockState {
	return &blockState{
		blkCache: &cache.LRU[ids.ID, *blockWrapper]{Size: blockCacheSize},
		db:       db,
	}
}

func NewMeteredBlockState(db database.Database, namespace string, metrics prometheus.Registerer) (BlockState, error) {
	blkCache, err := metercacher.New[ids.ID, *blockWrapper](
		fmt.Sprintf("%s_block_cache", namespace),
		metrics,
		&cache.LRU[ids.ID, *blockWrapper]{Size: blockCacheSize},
	)

	return &blockState{
		blkCache: blkCache,
		db:       db,
	}, err
}

func (s *blockState) GetBlock(blkID ids.ID) (block.Block, choices.Status, error) {
	if blk, found := s.blkCache.Get(blkID); found {
		if blk == nil {
			return nil, choices.Unknown, database.ErrNotFound
		}
		return blk.block, blk.Status, nil
	}

	blkWrapperBytes, err := s.db.Get(blkID[:])
	if err == database.ErrNotFound {
		s.blkCache.Put(blkID, nil)
		return nil, choices.Unknown, database.ErrNotFound
	}
	if err != nil {
		return nil, choices.Unknown, err
	}

	blkWrapper := blockWrapper{}
	parsedVersion, err := c.Unmarshal(blkWrapperBytes, &blkWrapper)
	if err != nil {
		return nil, choices.Unknown, err
	}
	if parsedVersion != version {
		return nil, choices.Unknown, errBlockWrongVersion
	}

	// The key was in the database
	blk, err := block.Parse(blkWrapper.Block)
	if err != nil {
		return nil, choices.Unknown, err
	}
	blkWrapper.block = blk

	s.blkCache.Put(blkID, &blkWrapper)
	return blk, blkWrapper.Status, nil
}

func (s *blockState) PutBlock(blk block.Block, status choices.Status) error {
	blkWrapper := blockWrapper{
		Block:  blk.Bytes(),
		Status: status,
		block:  blk,
	}

	bytes, err := c.Marshal(version, &blkWrapper)
	if err != nil {
		return err
	}

	blkID := blk.ID()
	s.blkCache.Put(blkID, &blkWrapper)
	return s.db.Put(blkID[:], bytes)
}
