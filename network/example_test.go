// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/VidarSolutions/avalanchego/genesis"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/message"
	"github.com/VidarSolutions/avalanchego/snow/networking/router"
	"github.com/VidarSolutions/avalanchego/snow/validators"
	"github.com/VidarSolutions/avalanchego/utils/constants"
	"github.com/VidarSolutions/avalanchego/utils/ips"
	"github.com/VidarSolutions/avalanchego/utils/logging"
	"github.com/VidarSolutions/avalanchego/utils/set"
	"github.com/VidarSolutions/avalanchego/version"
)

var _ router.ExternalHandler = (*testExternalHandler)(nil)

// Note: all of the external handler's methods are called on peer goroutines. It
// is possible for multiple concurrent calls to happen with different NodeIDs.
// However, a given NodeID will only be performing one call at a time.
type testExternalHandler struct {
	log logging.Logger
}

// Note: HandleInbound will be called with raw P2P messages, the networking
// implementation does not implicitly register timeouts, so this handler is only
// called by messages explicitly sent by the peer. If timeouts are required,
// that must be handled by the user of this utility.
func (t *testExternalHandler) HandleInbound(_ context.Context, message message.InboundMessage) {
	t.log.Info(
		"receiving message",
		zap.Stringer("op", message.Op()),
	)
}

func (t *testExternalHandler) Connected(nodeID ids.NodeID, version *version.Application, subnetID ids.ID) {
	t.log.Info(
		"connected",
		zap.Stringer("nodeID", nodeID),
		zap.Stringer("version", version),
		zap.Stringer("subnetID", subnetID),
	)
}

func (t *testExternalHandler) Disconnected(nodeID ids.NodeID) {
	t.log.Info(
		"disconnected",
		zap.Stringer("nodeID", nodeID),
	)
}

type testAggressiveValidatorSet struct {
	validators.Set
}

func (*testAggressiveValidatorSet) Contains(ids.NodeID) bool {
	return true
}

func ExampleNewTestNetwork() {
	log := logging.NewLogger(
		"networking",
		logging.NewWrappedCore(
			logging.Info,
			os.Stdout,
			logging.Colors.ConsoleEncoder(),
		),
	)

	// Needs to be periodically updated by the caller to have the latest
	// validator set
	validators := &testAggressiveValidatorSet{
		Set: validators.NewSet(),
	}

	// If we want to be able to communicate with non-primary network subnets, we
	// should register them here.
	trackedSubnets := set.Set[ids.ID]{}

	// Messages and connections are handled by the external handler.
	handler := &testExternalHandler{
		log: log,
	}

	network, err := NewTestNetwork(
		log,
		constants.FujiID,
		validators,
		trackedSubnets,
		handler,
	)
	if err != nil {
		log.Fatal(
			"failed to create test network",
			zap.Error(err),
		)
		return
	}

	// We need to initially connect to some nodes in the network before peer
	// gossip will enable connecting to all the remaining nodes in the network.
	beaconIPs, beaconIDs := genesis.SampleBeacons(constants.FujiID, 5)
	for i, beaconIDStr := range beaconIDs {
		beaconID, err := ids.NodeIDFromString(beaconIDStr)
		if err != nil {
			log.Fatal(
				"failed to parse beaconID",
				zap.String("beaconID", beaconIDStr),
				zap.Error(err),
			)
			return
		}

		beaconIPStr := beaconIPs[i]
		ipPort, err := ips.ToIPPort(beaconIPStr)
		if err != nil {
			log.Fatal(
				"failed to parse beaconIP",
				zap.String("beaconIP", beaconIPStr),
				zap.Error(err),
			)
			return
		}

		network.ManuallyTrack(beaconID, ipPort)
	}

	// Typically network.StartClose() should be called based on receiving a
	// SIGINT or SIGTERM. For the example, we close the network after 15s.
	go log.RecoverAndPanic(func() {
		time.Sleep(15 * time.Second)
		network.StartClose()
	})

	// network.Send(...) and network.Gossip(...) can be used here to send
	// messages to peers.

	// Calling network.Dispatch() will block until a fatal error occurs or
	// network.StartClose() is called.
	err = network.Dispatch()
	log.Info(
		"network exited",
		zap.Error(err),
	)
}
