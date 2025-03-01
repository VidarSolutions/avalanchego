// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ledger

import (
	"fmt"

	ledger "github.com/ava-labs/ledger-avalanche/go"

	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/utils/crypto/keychain"
	"github.com/VidarSolutions/avalanchego/utils/hashing"
	"github.com/VidarSolutions/avalanchego/version"
)

const (
	rootPath          = "m/44'/9000'/0'"
	ledgerBufferLimit = 8192
	ledgerPathSize    = 9
)

var _ keychain.Ledger = (*Ledger)(nil)

// Ledger is a wrapper around the low-level Ledger Device interface that
// provides Avalanche-specific access.
type Ledger struct {
	device *ledger.LedgerAvalanche
}

func New() (keychain.Ledger, error) {
	device, err := ledger.FindLedgerAvalancheApp()
	return &Ledger{
		device: device,
	}, err
}

func addressPath(index uint32) string {
	return fmt.Sprintf("%s/0/%d", rootPath, index)
}

func (l *Ledger) Address(hrp string, addressIndex uint32) (ids.ShortID, error) {
	_, hash, err := l.device.GetPubKey(addressPath(addressIndex), true, hrp, "")
	if err != nil {
		return ids.ShortEmpty, err
	}
	return ids.ToShortID(hash)
}

func (l *Ledger) Addresses(addressIndices []uint32) ([]ids.ShortID, error) {
	addresses := make([]ids.ShortID, len(addressIndices))
	for i, v := range addressIndices {
		_, hash, err := l.device.GetPubKey(addressPath(v), false, "", "")
		if err != nil {
			return nil, err
		}
		copy(addresses[i][:], hash)
	}
	return addresses, nil
}

func convertToSigningPaths(input []uint32) []string {
	output := make([]string, len(input))
	for i, v := range input {
		output[i] = fmt.Sprintf("0/%d", v)
	}
	return output
}

func (l *Ledger) SignHash(hash []byte, addressIndices []uint32) ([][]byte, error) {
	strIndices := convertToSigningPaths(addressIndices)
	response, err := l.device.SignHash(rootPath, strIndices, hash)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to sign hash", err)
	}
	responses := make([][]byte, len(addressIndices))
	for i, index := range strIndices {
		sig, ok := response.Signature[index]
		if !ok {
			return nil, fmt.Errorf("missing signature %s", index)
		}
		responses[i] = sig
	}
	return responses, nil
}

func (l *Ledger) Sign(txBytes []byte, addressIndices []uint32) ([][]byte, error) {
	// will pass to the ledger addressIndices both as signing paths and change paths
	numSigningPaths := len(addressIndices)
	numChangePaths := len(addressIndices)
	if len(txBytes)+(numSigningPaths+numChangePaths)*ledgerPathSize > ledgerBufferLimit {
		// There is a limit on the tx length that can be parsed by the ledger
		// app. When the tx that is being signed is too large, we sign with hash
		// instead.
		//
		// Ref: https://github.com/ava-labs/avalanche-wallet-sdk/blob/9a71f05e424e06b94eaccf21fd32d7983ed1b040/src/Wallet/Ledger/provider/ZondaxProvider.ts#L68
		unsignedHash := hashing.ComputeHash256(txBytes)
		return l.SignHash(unsignedHash, addressIndices)
	}
	strIndices := convertToSigningPaths(addressIndices)
	response, err := l.device.Sign(rootPath, strIndices, txBytes, strIndices)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to sign transaction", err)
	}
	responses := make([][]byte, len(strIndices))
	for i, index := range strIndices {
		sig, ok := response.Signature[index]
		if !ok {
			return nil, fmt.Errorf("missing signature %s", index)
		}
		responses[i] = sig
	}
	return responses, nil
}

func (l *Ledger) Version() (*version.Semantic, error) {
	resp, err := l.device.GetVersion()
	if err != nil {
		return nil, err
	}
	return &version.Semantic{
		Major: int(resp.Major),
		Minor: int(resp.Minor),
		Patch: int(resp.Patch),
	}, nil
}

func (l *Ledger) Disconnect() error {
	return l.device.Close()
}
