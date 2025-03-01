// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"log"
	"time"

	"github.com/VidarSolutions/avalanchego/genesis"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/utils/units"
	"github.com/VidarSolutions/avalanchego/vms/components/verify"
	"github.com/VidarSolutions/avalanchego/vms/secp256k1fx"
	"github.com/VidarSolutions/avalanchego/wallet/subnet/primary"
)

func main() {
	key := genesis.EWOQKey
	uri := primary.LocalAPIURI
	kc := secp256k1fx.NewKeychain(key)
	subnetOwner := key.Address()

	ctx := context.Background()

	// NewWalletFromURI fetches the available UTXOs owned by [kc] on the network
	// that [uri] is hosting.
	walletSyncStartTime := time.Now()
	wallet, err := primary.NewWalletFromURI(ctx, uri, kc)
	if err != nil {
		log.Fatalf("failed to initialize wallet: %s\n", err)
	}
	log.Printf("synced wallet in %s\n", time.Since(walletSyncStartTime))

	// Get the X-chain wallet
	xWallet := wallet.X()

	// Pull out useful constants to use when issuing transactions.
	owner := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs: []ids.ShortID{
			subnetOwner,
		},
	}

	createAssetStartTime := time.Now()
	createAssetTxID, err := xWallet.IssueCreateAssetTx(
		"HI",
		"HI",
		1,
		map[uint32][]verify.State{
			0: {
				&secp256k1fx.TransferOutput{
					Amt:          units.Schmeckle,
					OutputOwners: *owner,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("failed to issue create asset transaction: %s\n", err)
	}
	log.Printf("created new asset %s in %s\n", createAssetTxID, time.Since(createAssetStartTime))
}
