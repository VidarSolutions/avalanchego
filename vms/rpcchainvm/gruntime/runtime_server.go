// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gruntime

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/VidarSolutions/avalanchego/vms/rpcchainvm/runtime"

	pb "github.com/VidarSolutions/avalanchego/proto/pb/vm/runtime"
)

var _ pb.RuntimeServer = &Server{}

// Server is a VM runtime initializer controlled by RPC.
type Server struct {
	pb.UnsafeRuntimeServer
	runtime runtime.Initializer
}

func NewServer(runtime runtime.Initializer) *Server {
	return &Server{
		runtime: runtime,
	}
}

func (s *Server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.runtime.Initialize(ctx, uint(req.ProtocolVersion), req.Addr)
}
