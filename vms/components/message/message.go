// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message

import (
	"errors"

	"github.com/VidarSolutions/avalanchego/ids"
)

var (
	_ Message = (*Tx)(nil)

	ErrUnexpectedCodecVersion = errors.New("unexpected codec version")
)

type Message interface {
	// Handle this message with the correct message handler
	Handle(handler Handler, nodeID ids.NodeID, requestID uint32) error

	// initialize should be called whenever a message is built or parsed
	initialize([]byte)

	// Bytes returns the binary representation of this message
	//
	// Bytes should only be called after being initialized
	Bytes() []byte
}

type message []byte

func (m *message) initialize(bytes []byte) {
	*m = bytes
}

func (m *message) Bytes() []byte {
	return *m
}

func Parse(bytes []byte) (Message, error) {
	var msg Message
	version, err := c.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	if version != codecVersion {
		return nil, ErrUnexpectedCodecVersion
	}
	msg.initialize(bytes)
	return msg, nil
}

func Build(msg Message) ([]byte, error) {
	bytes, err := c.Marshal(codecVersion, &msg)
	msg.initialize(bytes)
	return bytes, err
}
