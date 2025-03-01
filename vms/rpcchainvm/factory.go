// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpcchainvm

import (
	"context"
	"fmt"

	"github.com/VidarSolutions/avalanchego/utils/logging"
	"github.com/VidarSolutions/avalanchego/utils/resource"
	"github.com/VidarSolutions/avalanchego/vms"
	"github.com/VidarSolutions/avalanchego/vms/rpcchainvm/grpcutils"
	"github.com/VidarSolutions/avalanchego/vms/rpcchainvm/runtime"
	"github.com/VidarSolutions/avalanchego/vms/rpcchainvm/runtime/subprocess"

	vmpb "github.com/VidarSolutions/avalanchego/proto/pb/vm"
)

var _ vms.Factory = (*factory)(nil)

type factory struct {
	path           string
	processTracker resource.ProcessTracker
	runtimeTracker runtime.Tracker
}

func NewFactory(path string, processTracker resource.ProcessTracker, runtimeTracker runtime.Tracker) vms.Factory {
	return &factory{
		path:           path,
		processTracker: processTracker,
		runtimeTracker: runtimeTracker,
	}
}

func (f *factory) New(log logging.Logger) (interface{}, error) {
	config := &subprocess.Config{
		Stderr:           log,
		Stdout:           log,
		HandshakeTimeout: runtime.DefaultHandshakeTimeout,
		Log:              log,
	}

	listener, err := grpcutils.NewListener()
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	status, stopper, err := subprocess.Bootstrap(
		context.TODO(),
		listener,
		subprocess.NewCmd(f.path),
		config,
	)
	if err != nil {
		return nil, err
	}

	clientConn, err := grpcutils.Dial(status.Addr)
	if err != nil {
		return nil, err
	}

	vm := NewClient(vmpb.NewVMClient(clientConn))
	vm.SetProcess(stopper, status.Pid, f.processTracker)

	f.runtimeTracker.TrackRuntime(stopper)

	return vm, nil
}
