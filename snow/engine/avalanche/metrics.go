// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/VidarSolutions/avalanchego/utils/wrappers"
)

type metrics struct {
	bootstrapFinished,
	numVtxRequests, numPendingVts,
	numMissingTxs, pendingTxs,
	blockerVtxs, blockerTxs prometheus.Gauge

	whitelistVtxIssueSuccess, whitelistVtxIssueFailure,
	numUselessPutBytes, numUselessPushQueryBytes prometheus.Counter
}

func (m *metrics) Initialize(namespace string, reg prometheus.Registerer) error {
	errs := wrappers.Errs{}
	m.bootstrapFinished = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "bootstrap_finished",
		Help:      "Whether or not bootstrap process has completed. 1 is success, 0 is fail or ongoing.",
	})
	m.numVtxRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "vtx_requests",
		Help:      "Number of outstanding vertex requests",
	})
	m.numPendingVts = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "pending_vts",
		Help:      "Number of pending vertices",
	})
	m.numMissingTxs = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "missing_txs",
		Help:      "Number of missing transactions",
	})
	m.pendingTxs = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "pending_txs",
		Help:      "Number of transactions from the VM waiting to be issued",
	})
	m.blockerVtxs = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "blocker_vtxs",
		Help:      "Number of vertices that are blocking other vertices from being issued because they haven't been issued",
	})
	m.blockerTxs = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "blocker_txs",
		Help:      "Number of transactions that are blocking other transactions from being issued because they haven't been issued",
	})
	m.whitelistVtxIssueSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "whitelist_vtx_issue_success",
		Help:      "Number of DAG linearization request issued (pending, not necessarily accepted)",
	})
	m.whitelistVtxIssueFailure = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "whitelist_vtx_issue_failure",
		Help:      "Number of DAG linearization request issue failed (verification failure)",
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

	errs.Add(
		reg.Register(m.bootstrapFinished),
		reg.Register(m.numVtxRequests),
		reg.Register(m.numPendingVts),
		reg.Register(m.numMissingTxs),
		reg.Register(m.pendingTxs),
		reg.Register(m.blockerVtxs),
		reg.Register(m.blockerTxs),
		reg.Register(m.whitelistVtxIssueSuccess),
		reg.Register(m.whitelistVtxIssueFailure),
		reg.Register(m.numUselessPutBytes),
		reg.Register(m.numUselessPushQueryBytes),
	)
	return errs.Err
}
