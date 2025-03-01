// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package proposervm

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/snow"
	"github.com/VidarSolutions/avalanchego/snow/consensus/snowman"
	"github.com/VidarSolutions/avalanchego/snow/engine/snowman/block"
	"github.com/VidarSolutions/avalanchego/snow/engine/snowman/block/mocks"
	"github.com/VidarSolutions/avalanchego/snow/validators"
	"github.com/VidarSolutions/avalanchego/utils/logging"
	"github.com/VidarSolutions/avalanchego/vms/proposervm/proposer"
)

// Assert that when the underlying VM implements ChainVMWithBuildBlockContext
// and the proposervm is activated, we call the VM's BuildBlockWithContext
// method to build a block rather than BuildBlockWithContext. If the proposervm
// isn't activated, we should call BuildBlock rather than BuildBlockWithContext.
func TestPostForkCommonComponents_buildChild(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pChainHeight := uint64(1337)
	parentID := ids.GenerateTestID()
	parentTimestamp := time.Now()
	blkID := ids.GenerateTestID()
	innerBlk := snowman.NewMockBlock(ctrl)
	innerBlk.EXPECT().ID().Return(blkID).AnyTimes()
	innerBlk.EXPECT().Height().Return(pChainHeight - 1).AnyTimes()
	builtBlk := snowman.NewMockBlock(ctrl)
	builtBlk.EXPECT().Bytes().Return([]byte{1, 2, 3}).AnyTimes()
	builtBlk.EXPECT().ID().Return(ids.GenerateTestID()).AnyTimes()
	builtBlk.EXPECT().Height().Return(pChainHeight).AnyTimes()
	innerVM := mocks.NewMockChainVM(ctrl)
	innerBlockBuilderVM := mocks.NewMockBuildBlockWithContextChainVM(ctrl)
	innerBlockBuilderVM.EXPECT().BuildBlockWithContext(gomock.Any(), &block.Context{
		PChainHeight: pChainHeight - 1,
	}).Return(builtBlk, nil).AnyTimes()
	vdrState := validators.NewMockState(ctrl)
	vdrState.EXPECT().GetMinimumHeight(context.Background()).Return(pChainHeight, nil).AnyTimes()
	windower := proposer.NewMockWindower(ctrl)
	windower.EXPECT().Delay(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(time.Duration(0), nil).AnyTimes()

	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(err)
	vm := &VM{
		ChainVM:        innerVM,
		blockBuilderVM: innerBlockBuilderVM,
		ctx: &snow.Context{
			ValidatorState: vdrState,
			Log:            logging.NoLog{},
		},
		Windower:          windower,
		stakingCertLeaf:   &x509.Certificate{},
		stakingLeafSigner: pk,
	}

	blk := &postForkCommonComponents{
		innerBlk: innerBlk,
		vm:       vm,
	}

	// Should call BuildBlockWithContext since proposervm is activated
	gotChild, err := blk.buildChild(
		context.Background(),
		parentID,
		parentTimestamp,
		pChainHeight-1,
	)
	require.NoError(err)
	require.Equal(builtBlk, gotChild.(*postForkBlock).innerBlk)
}
