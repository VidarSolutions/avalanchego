// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sender

import (
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/message"
	"github.com/VidarSolutions/avalanchego/subnets"
	"github.com/VidarSolutions/avalanchego/utils/set"
)

// ExternalSender sends consensus messages to other validators
// Right now this is implemented in the networking package
type ExternalSender interface {
	// Send a message to a specific set of nodes
	Send(
		msg message.OutboundMessage,
		nodeIDs set.Set[ids.NodeID],
		subnetID ids.ID,
		allower subnets.Allower,
	) set.Set[ids.NodeID]

	// Send a message to a random group of nodes in a subnet.
	// Nodes are sampled based on their validator status.
	Gossip(
		msg message.OutboundMessage,
		subnetID ids.ID,
		numValidatorsToSend int,
		numNonValidatorsToSend int,
		numPeersToSend int,
		allower subnets.Allower,
	) set.Set[ids.NodeID]
}
