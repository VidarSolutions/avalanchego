// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowstorm

import (
	"context"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow/events"
	"github.com/VidarSolutions/avalanchego/utils/set"
	"github.com/VidarSolutions/avalanchego/utils/wrappers"
)

var _ events.Blockable = (*rejector)(nil)

type rejector struct {
	g        *Directed
	errs     *wrappers.Errs
	deps     set.Set[ids.ID]
	rejected bool // true if the tx has been rejected
	txID     ids.ID
}

func (r *rejector) Dependencies() set.Set[ids.ID] {
	return r.deps
}

func (r *rejector) Fulfill(ctx context.Context, _ ids.ID) {
	if r.rejected || r.errs.Errored() {
		return
	}
	r.rejected = true
	asSet := set.NewSet[ids.ID](1)
	asSet.Add(r.txID)
	r.errs.Add(r.g.reject(ctx, asSet))
}

func (*rejector) Abandon(context.Context, ids.ID) {}

func (*rejector) Update(context.Context) {}
