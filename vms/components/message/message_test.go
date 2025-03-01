// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/utils"
	"github.com/VidarSolutions/avalanchego/utils/units"
)

func TestParseGibberish(t *testing.T) {
	randomBytes := utils.RandomBytes(256 * units.KiB)
	_, err := Parse(randomBytes)
	require.Error(t, err)
}
