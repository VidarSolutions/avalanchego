// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"context"
	"time"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/engine/common"
	"github.com/VidarSolutions/avalanchego/utils"
)

var _ Atomic = (*atomic)(nil)

type Atomic interface {
	common.AppHandler

	Set(common.AppHandler)
}

type atomic struct {
	handler utils.Atomic[common.AppHandler]
}

func NewAtomic(h common.AppHandler) Atomic {
	a := &atomic{}
	a.handler.Set(h)
	return a
}

func (a *atomic) CrossChainAppRequest(
	ctx context.Context,
	chainID ids.ID,
	requestID uint32,
	deadline time.Time,
	msg []byte,
) error {
	h := a.handler.Get()
	return h.CrossChainAppRequest(
		ctx,
		chainID,
		requestID,
		deadline,
		msg,
	)
}

func (a *atomic) CrossChainAppRequestFailed(
	ctx context.Context,
	chainID ids.ID,
	requestID uint32,
) error {
	h := a.handler.Get()
	return h.CrossChainAppRequestFailed(
		ctx,
		chainID,
		requestID,
	)
}

func (a *atomic) CrossChainAppResponse(
	ctx context.Context,
	chainID ids.ID,
	requestID uint32,
	msg []byte,
) error {
	h := a.handler.Get()
	return h.CrossChainAppResponse(
		ctx,
		chainID,
		requestID,
		msg,
	)
}

func (a *atomic) AppRequest(
	ctx context.Context,
	nodeID ids.NodeID,
	requestID uint32,
	deadline time.Time,
	msg []byte,
) error {
	h := a.handler.Get()
	return h.AppRequest(
		ctx,
		nodeID,
		requestID,
		deadline,
		msg,
	)
}

func (a *atomic) AppRequestFailed(
	ctx context.Context,
	nodeID ids.NodeID,
	requestID uint32,
) error {
	h := a.handler.Get()
	return h.AppRequestFailed(
		ctx,
		nodeID,
		requestID,
	)
}

func (a *atomic) AppResponse(
	ctx context.Context,
	nodeID ids.NodeID,
	requestID uint32,
	msg []byte,
) error {
	h := a.handler.Get()
	return h.AppResponse(
		ctx,
		nodeID,
		requestID,
		msg,
	)
}

func (a *atomic) AppGossip(
	ctx context.Context,
	nodeID ids.NodeID,
	msg []byte,
) error {
	h := a.handler.Get()
	return h.AppGossip(
		ctx,
		nodeID,
		msg,
	)
}

func (a *atomic) Set(h common.AppHandler) {
	a.handler.Set(h)
}
