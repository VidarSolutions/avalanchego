// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package propertyfx

import (
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/utils/logging"
	"github.com/VidarSolutions/avalanchego/vms"
)

var (
	_ vms.Factory = (*Factory)(nil)

	// ID that this Fx uses when labeled
	ID = ids.ID{'p', 'r', 'o', 'p', 'e', 'r', 't', 'y', 'f', 'x'}
)

type Factory struct{}

func (*Factory) New(logging.Logger) (interface{}, error) {
	return &Fx{}, nil
}
