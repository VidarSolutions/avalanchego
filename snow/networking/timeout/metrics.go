// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timeout

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"go.uber.org/zap"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/message"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/utils/metric"
	"github.com/VidarSolutions/avalanchego/utils/wrappers"
)

const (
	defaultRequestHelpMsg = "time (in ns) spent waiting for a response to this message"
	validatorIDLabel      = "validatorID"
)

type metrics struct {
	lock           sync.Mutex
	chainToMetrics map[ids.ID]*chainMetrics
}

func (m *metrics) RegisterChain(ctx *snow.ConsensusContext) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.chainToMetrics == nil {
		m.chainToMetrics = map[ids.ID]*chainMetrics{}
	}
	if _, exists := m.chainToMetrics[ctx.ChainID]; exists {
		return fmt.Errorf("chain %s has already been registered", ctx.ChainID)
	}
	cm, err := newChainMetrics(ctx, false)
	if err != nil {
		return fmt.Errorf("couldn't create metrics for chain %s: %w", ctx.ChainID, err)
	}
	m.chainToMetrics[ctx.ChainID] = cm
	return nil
}

// Record that a response of type [op] took [latency]
func (m *metrics) Observe(nodeID ids.NodeID, chainID ids.ID, op message.Op, latency time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cm, exists := m.chainToMetrics[chainID]
	if !exists {
		// TODO should this log an error?
		return
	}
	cm.observe(nodeID, op, latency)
}

// chainMetrics contains message response time metrics for a chain
type chainMetrics struct {
	ctx *snow.ConsensusContext

	messageLatencies map[message.Op]metric.Averager

	summaryEnabled   bool
	messageSummaries map[message.Op]*prometheus.SummaryVec
}

func newChainMetrics(ctx *snow.ConsensusContext, summaryEnabled bool) (*chainMetrics, error) {
	cm := &chainMetrics{
		ctx: ctx,

		messageLatencies: make(map[message.Op]metric.Averager, len(message.ConsensusResponseOps)),

		summaryEnabled:   summaryEnabled,
		messageSummaries: make(map[message.Op]*prometheus.SummaryVec, len(message.ConsensusResponseOps)),
	}

	errs := wrappers.Errs{}
	for _, op := range message.ConsensusResponseOps {
		cm.messageLatencies[op] = metric.NewAveragerWithErrs(
			"lat",
			op.String(),
			defaultRequestHelpMsg,
			ctx.Registerer,
			&errs,
		)

		if !summaryEnabled {
			continue
		}

		summaryName := fmt.Sprintf("%s_peer", op)
		summary := prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: "lat",
				Name:      summaryName,
				Help:      defaultRequestHelpMsg,
			},
			[]string{validatorIDLabel},
		)
		cm.messageSummaries[op] = summary

		if err := ctx.Registerer.Register(summary); err != nil {
			errs.Add(fmt.Errorf("failed to register %s statistics: %w", summaryName, err))
		}
	}
	return cm, errs.Err
}

func (cm *chainMetrics) observe(nodeID ids.NodeID, op message.Op, latency time.Duration) {
	lat := float64(latency)
	if msg, exists := cm.messageLatencies[op]; exists {
		msg.Observe(lat)
	}

	if !cm.summaryEnabled {
		return
	}

	labels := prometheus.Labels{
		validatorIDLabel: nodeID.String(),
	}

	msg, exists := cm.messageSummaries[op]
	if !exists {
		return
	}

	observer, err := msg.GetMetricWith(labels)
	if err != nil {
		cm.ctx.Log.Warn("failed to get observer with validatorID",
			zap.Error(err),
		)
		return
	}
	observer.Observe(lat)
}
