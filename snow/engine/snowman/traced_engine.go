// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/consensus/snowman"
	"github.com/VidarSolutions/avalanchego/snow/engine/common"
	"github.com/VidarSolutions/avalanchego/trace"
)

var _ Engine = (*tracedEngine)(nil)

type tracedEngine struct {
	common.Engine
	engine Engine
	tracer trace.Tracer
}

func TraceEngine(engine Engine, tracer trace.Tracer) Engine {
	return &tracedEngine{
		Engine: common.TraceEngine(engine, tracer),
		engine: engine,
		tracer: tracer,
	}
}

func (e *tracedEngine) GetBlock(ctx context.Context, blkID ids.ID) (snowman.Block, error) {
	ctx, span := e.tracer.Start(ctx, "tracedEngine.GetBlock", oteltrace.WithAttributes(
		attribute.Stringer("blkID", blkID),
	))
	defer span.End()

	return e.engine.GetBlock(ctx, blkID)
}
