// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import (
	"context"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/choices"
	"github.com/VidarSolutions/avalanchego/snow/consensus/snowstorm"
	"github.com/VidarSolutions/avalanchego/utils/set"
)

var _ Vertex = (*TestVertex)(nil)

// TestVertex is a useful test vertex
type TestVertex struct {
	choices.TestDecidable

	VerifyErrV    error
	ParentsV      []Vertex
	ParentsErrV   error
	HasWhitelistV bool
	WhitelistV    set.Set[ids.ID]
	WhitelistErrV error
	HeightV       uint64
	HeightErrV    error
	TxsV          []snowstorm.Tx
	TxsErrV       error
	BytesV        []byte
}

func (v *TestVertex) Verify(context.Context) error {
	return v.VerifyErrV
}

func (v *TestVertex) Parents() ([]Vertex, error) {
	return v.ParentsV, v.ParentsErrV
}

func (v *TestVertex) HasWhitelist() bool {
	return v.HasWhitelistV
}

func (v *TestVertex) Whitelist(context.Context) (set.Set[ids.ID], error) {
	return v.WhitelistV, v.WhitelistErrV
}

func (v *TestVertex) Height() (uint64, error) {
	return v.HeightV, v.HeightErrV
}

func (v *TestVertex) Txs(context.Context) ([]snowstorm.Tx, error) {
	return v.TxsV, v.TxsErrV
}

func (v *TestVertex) Bytes() []byte {
	return v.BytesV
}
