// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/VidarSolutions/avalanchego/utils/metric"
	"github.com/VidarSolutions/avalanchego/utils/wrappers"
)

type metrics struct {
	bootstrapFinished, numRequests, numBlocked, numBlockers, numNonVerifieds prometheus.Gauge
	numBuilt, numBuildsFailed, numUselessPutBytes, numUselessPushQueryBytes  prometheus.Counter
	getAncestorsBlks                                                         metric.Averager
}

// Initialize the metrics
func (m *metrics) Initialize(namespace string, reg prometheus.Registerer) error {
	errs := wrappers.Errs{}
	m.bootstrapFinished = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "bootstrap_finished",
		Help:      "Whether or not bootstrap process has completed. 1 is success, 0 is fail or ongoing.",
	})
	m.numRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "requests",
		Help:      "Number of outstanding block requests",
	})
	m.numBlocked = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "blocked",
		Help:      "Number of blocks that are pending issuance",
	})
	m.numBlockers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "blockers",
		Help:      "Number of blocks that are blocking other blocks from being issued because they haven't been issued",
	})
	m.numBuilt = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "blks_built",
		Help:      "Number of blocks that have been built locally",
	})
	m.numBuildsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "blk_builds_failed",
		Help:      "Number of BuildBlock calls that have failed",
	})
	m.numUselessPutBytes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "num_useless_put_bytes",
		Help:      "Amount of useless bytes received in Put messages",
	})
	m.numUselessPushQueryBytes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "num_useless_push_query_bytes",
		Help:      "Amount of useless bytes received in PushQuery messages",
	})
	m.getAncestorsBlks = metric.NewAveragerWithErrs(
		namespace,
		"get_ancestors_blks",
		"blocks fetched in a call to GetAncestors",
		reg,
		&errs,
	)
	m.numNonVerifieds = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "non_verified_blks",
		Help:      "Number of non-verified blocks in the memory",
	})

	errs.Add(
		reg.Register(m.bootstrapFinished),
		reg.Register(m.numRequests),
		reg.Register(m.numBlocked),
		reg.Register(m.numBlockers),
		reg.Register(m.numNonVerifieds),
		reg.Register(m.numBuilt),
		reg.Register(m.numBuildsFailed),
		reg.Register(m.numUselessPutBytes),
		reg.Register(m.numUselessPushQueryBytes),
	)
	return errs.Err
}
