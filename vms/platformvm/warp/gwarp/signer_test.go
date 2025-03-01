// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gwarp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/utils/crypto/bls"
	"github.com/VidarSolutions/avalanchego/vms/platformvm/warp"
	"github.com/VidarSolutions/avalanchego/vms/rpcchainvm/grpcutils"

	pb "github.com/VidarSolutions/avalanchego/proto/pb/warp"
)

type testSigner struct {
	client  *Client
	server  warp.Signer
	sk      *bls.SecretKey
	chainID ids.ID
	closeFn func()
}

func setupSigner(t testing.TB) *testSigner {
	require := require.New(t)

	sk, err := bls.NewSecretKey()
	require.NoError(err)

	chainID := ids.GenerateTestID()

	s := &testSigner{
		server:  warp.NewSigner(sk, chainID),
		sk:      sk,
		chainID: chainID,
	}

	listener, err := grpcutils.NewListener()
	if err != nil {
		t.Fatalf("Failed to create listener: %s", err)
	}
	serverCloser := grpcutils.ServerCloser{}

	server := grpcutils.NewServer()
	pb.RegisterSignerServer(server, NewServer(s.server))
	serverCloser.Add(server)

	go grpcutils.Serve(listener, server)

	conn, err := grpcutils.Dial(listener.Addr().String())
	require.NoError(err)

	s.client = NewClient(pb.NewSignerClient(conn))
	s.closeFn = func() {
		serverCloser.Stop()
		_ = conn.Close()
		_ = listener.Close()
	}
	return s
}

func TestInterface(t *testing.T) {
	for _, test := range warp.SignerTests {
		s := setupSigner(t)
		test(t, s.client, s.sk, s.chainID)
		s.closeFn()
	}
}
