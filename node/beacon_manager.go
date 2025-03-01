// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"sync/atomic"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/networking/router"
	"github.com/VidarSolutions/avalanchego/snow/validators"
	"github.com/VidarSolutions/avalanchego/utils/constants"
	"github.com/VidarSolutions/avalanchego/utils/timer"
	"github.com/VidarSolutions/avalanchego/version"
)

var _ router.Router = (*beaconManager)(nil)

type beaconManager struct {
	router.Router
	timer         *timer.Timer
	beacons       validators.Set
	requiredConns int64
	numConns      int64
}

func (b *beaconManager) Connected(nodeID ids.NodeID, nodeVersion *version.Application, subnetID ids.ID) {
	if constants.PrimaryNetworkID == subnetID &&
		b.beacons.Contains(nodeID) &&
		atomic.AddInt64(&b.numConns, 1) >= b.requiredConns {
		b.timer.Cancel()
	}
	b.Router.Connected(nodeID, nodeVersion, subnetID)
}

func (b *beaconManager) Disconnected(nodeID ids.NodeID) {
	if b.beacons.Contains(nodeID) {
		atomic.AddInt64(&b.numConns, -1)
	}
	b.Router.Disconnected(nodeID)
}
