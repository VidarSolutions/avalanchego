// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package secp256k1fx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/codec/linearcodec"
	"github.com/VidarSolutions/avalanchego/ids"
	"github.com/VidarSolutions/avalanchego/utils/cb58"
	"github.com/VidarSolutions/avalanchego/utils/crypto/secp256k1"
	"github.com/VidarSolutions/avalanchego/utils/logging"
)

var (
	txBytes  = []byte{0, 1, 2, 3, 4, 5}
	sigBytes = [secp256k1.SignatureLen]byte{ // signature of addr on txBytes
		0x0e, 0x33, 0x4e, 0xbc, 0x67, 0xa7, 0x3f, 0xe8,
		0x24, 0x33, 0xac, 0xa3, 0x47, 0x88, 0xa6, 0x3d,
		0x58, 0xe5, 0x8e, 0xf0, 0x3a, 0xd5, 0x84, 0xf1,
		0xbc, 0xa3, 0xb2, 0xd2, 0x5d, 0x51, 0xd6, 0x9b,
		0x0f, 0x28, 0x5d, 0xcd, 0x3f, 0x71, 0x17, 0x0a,
		0xf9, 0xbf, 0x2d, 0xb1, 0x10, 0x26, 0x5c, 0xe9,
		0xdc, 0xc3, 0x9d, 0x7a, 0x01, 0x50, 0x9d, 0xe8,
		0x35, 0xbd, 0xcb, 0x29, 0x3a, 0xd1, 0x49, 0x32,
		0x00,
	}
	addr = ids.ShortID{
		0x01, 0x5c, 0xce, 0x6c, 0x55, 0xd6, 0xb5, 0x09,
		0x84, 0x5c, 0x8c, 0x4e, 0x30, 0xbe, 0xd9, 0x8d,
		0x39, 0x1a, 0xe7, 0xf0,
	}
	addr2     ids.ShortID
	sig2Bytes [secp256k1.SignatureLen]byte // signature of addr2 on txBytes
)

func init() {
	b, err := cb58.Decode("31SoC6ehdWUWFcuzkXci7ymFEQ8HGTJgw")
	if err != nil {
		panic(err)
	}
	copy(addr2[:], b)
	b, err = cb58.Decode("c7doHa86hWYyfXTVnNsdP1CG1gxhXVpZ9Q5CiHi2oFRdnaxh2YR2Mvu2cUNMgyQy4BNQaXAxWWPt36BJ5pDWX1Xeos4h9L")
	if err != nil {
		panic(err)
	}
	copy(sig2Bytes[:], b)
}

func TestFxInitialize(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
}

func TestFxInitializeInvalid(t *testing.T) {
	require := require.New(t)
	fx := Fx{}
	require.ErrorIs(fx.Initialize(nil), ErrWrongVMType)
}

func TestFxVerifyTransfer(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	require.NoError(fx.Bootstrapping())
	require.NoError(fx.Bootstrapped())
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.NoError(fx.VerifyTransfer(tx, in, cred, out))
}

func TestFxVerifyTransferNilTx(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.ErrorIs(fx.VerifyTransfer(nil, in, cred, out), ErrWrongTxType)
}

func TestFxVerifyTransferNilOutput(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, in, cred, nil), ErrWrongUTXOType)
}

func TestFxVerifyTransferNilInput(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, nil, cred, out), ErrWrongInputType)
}

func TestFxVerifyTransferNilCredential(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, in, nil, out), ErrWrongCredentialType)
}

func TestFxVerifyTransferInvalidOutput(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 0,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, in, cred, out), errOutputUnoptimized)
}

func TestFxVerifyTransferWrongAmounts(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 2,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.Error(fx.VerifyTransfer(tx, in, cred, out))
}

func TestFxVerifyTransferTimelocked(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  uint64(date.Add(time.Second).Unix()),
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, in, cred, out), ErrTimelocked)
}

func TestFxVerifyTransferTooManySigners(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0, 1},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
			{},
		},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, in, cred, out), ErrTooManySigners)
}

func TestFxVerifyTransferTooFewSigners(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, in, cred, out), ErrTooFewSigners)
}

func TestFxVerifyTransferMismatchedSigners(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
			{},
		},
	}

	require.ErrorIs(fx.VerifyTransfer(tx, in, cred, out), ErrInputCredentialSignersMismatch)
}

func TestFxVerifyTransferInvalidSignature(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	require.NoError(fx.Bootstrapping())
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			{},
		},
	}

	require.NoError(fx.VerifyTransfer(tx, in, cred, out))
	require.NoError(fx.Bootstrapped())
	require.Error(fx.VerifyTransfer(tx, in, cred, out), errAddrsNotSortedUnique)
}

func TestFxVerifyTransferWrongSigner(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	require.NoError(fx.Bootstrapping())
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				ids.ShortEmpty,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{0},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.NoError(fx.VerifyTransfer(tx, in, cred, out))
	require.NoError(fx.Bootstrapped())
	require.Error(fx.VerifyTransfer(tx, in, cred, out))
}

func TestFxVerifyTransferSigIndexOOB(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	require.NoError(fx.Bootstrapping())
	tx := &TestTx{UnsignedBytes: txBytes}
	out := &TransferOutput{
		Amt: 1,
		OutputOwners: OutputOwners{
			Locktime:  0,
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	in := &TransferInput{
		Amt: 1,
		Input: Input{
			SigIndices: []uint32{1}, // There is no address at index 1
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	require.NoError(fx.VerifyTransfer(tx, in, cred, out))
	require.NoError(fx.Bootstrapped())
	require.ErrorIs(fx.VerifyTransfer(tx, in, cred, out), ErrInputOutputIndexOutOfBounds)
}

func TestFxVerifyOperation(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	utxo := &MintOutput{
		OutputOwners: OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
		TransferOutput: TransferOutput{
			Amt: 1,
			OutputOwners: OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	utxos := []interface{}{utxo}
	require.NoError(fx.VerifyOperation(tx, op, cred, utxos))
}

func TestFxVerifyOperationUnknownTx(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	utxo := &MintOutput{
		OutputOwners: OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
		TransferOutput: TransferOutput{
			Amt: 1,
			OutputOwners: OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	utxos := []interface{}{utxo}
	require.ErrorIs(fx.VerifyOperation(nil, op, cred, utxos), ErrWrongTxType)
}

func TestFxVerifyOperationUnknownOperation(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	utxo := &MintOutput{
		OutputOwners: OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	utxos := []interface{}{utxo}
	require.ErrorIs(fx.VerifyOperation(tx, nil, cred, utxos), ErrWrongOpType)
}

func TestFxVerifyOperationUnknownCredential(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	utxo := &MintOutput{
		OutputOwners: OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
		TransferOutput: TransferOutput{
			Amt: 1,
			OutputOwners: OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}

	utxos := []interface{}{utxo}
	require.ErrorIs(fx.VerifyOperation(tx, op, nil, utxos), ErrWrongCredentialType)
}

func TestFxVerifyOperationWrongNumberOfUTXOs(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	utxo := &MintOutput{
		OutputOwners: OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
		TransferOutput: TransferOutput{
			Amt: 1,
			OutputOwners: OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	utxos := []interface{}{utxo, utxo}
	require.ErrorIs(fx.VerifyOperation(tx, op, cred, utxos), ErrWrongNumberOfUTXOs)
}

func TestFxVerifyOperationUnknownUTXOType(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
		TransferOutput: TransferOutput{
			Amt: 1,
			OutputOwners: OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	utxos := []interface{}{nil}
	require.ErrorIs(fx.VerifyOperation(tx, op, cred, utxos), ErrWrongUTXOType)
}

func TestFxVerifyOperationInvalidOperationVerify(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	utxo := &MintOutput{
		OutputOwners: OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
		TransferOutput: TransferOutput{
			Amt: 1,
			OutputOwners: OutputOwners{
				Locktime:  0,
				Threshold: 1,
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	utxos := []interface{}{utxo}
	require.ErrorIs(fx.VerifyOperation(tx, op, cred, utxos), errOutputUnspendable)
}

func TestFxVerifyOperationMismatchedMintOutputs(t *testing.T) {
	require := require.New(t)
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	date := time.Date(2019, time.January, 19, 16, 25, 17, 3, time.UTC)
	vm.Clk.Set(date)
	fx := Fx{}
	require.NoError(fx.Initialize(&vm))
	tx := &TestTx{UnsignedBytes: txBytes}
	utxo := &MintOutput{
		OutputOwners: OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				addr,
			},
		},
	}
	op := &MintOperation{
		MintInput: Input{
			SigIndices: []uint32{0},
		},
		MintOutput: MintOutput{
			OutputOwners: OutputOwners{},
		},
		TransferOutput: TransferOutput{
			Amt: 1,
			OutputOwners: OutputOwners{
				Locktime:  0,
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}
	cred := &Credential{
		Sigs: [][secp256k1.SignatureLen]byte{
			sigBytes,
		},
	}

	utxos := []interface{}{utxo}
	require.ErrorIs(fx.VerifyOperation(tx, op, cred, utxos), ErrWrongMintCreated)
}

func TestVerifyPermission(t *testing.T) {
	vm := TestVM{
		Codec: linearcodec.NewDefault(),
		Log:   logging.NoLog{},
	}
	fx := Fx{}
	require.NoError(t, fx.Initialize(&vm))
	require.NoError(t, fx.Bootstrapping())
	require.NoError(t, fx.Bootstrapped())

	now := time.Now()
	fx.VM.Clock().Set(now)

	type test struct {
		description string
		tx          UnsignedTx
		in          *Input
		cred        *Credential
		cg          *OutputOwners
		expectedErr error
	}
	tests := []test{
		{
			"threshold 0, no sigs, has addrs",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{}},
			&OutputOwners{
				Threshold: 0,
				Addrs:     []ids.ShortID{addr},
			},
			errOutputUnoptimized,
		},
		{
			"threshold 0, no sigs, no addrs",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{}},
			&OutputOwners{
				Threshold: 0,
				Addrs:     []ids.ShortID{},
			},
			nil,
		},
		{
			"threshold 1, 1 sig",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes}},
			&OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr},
			},
			nil,
		},
		{
			"threshold 0, 1 sig (too many sigs)",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes}},
			&OutputOwners{
				Threshold: 0,
				Addrs:     []ids.ShortID{addr},
			},
			errOutputUnoptimized,
		},
		{
			"threshold 1, 0 sigs (too few sigs)",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{}},
			&OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr},
			},
			ErrTooFewSigners,
		},
		{
			"threshold 1, 1 incorrect sig",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes}},
			&OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{ids.GenerateTestShortID()},
			},
			ErrWrongSig,
		},
		{
			"repeated sig",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0, 0}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes, sigBytes}},
			&OutputOwners{
				Threshold: 2,
				Addrs:     []ids.ShortID{addr, addr2},
			},
			errNotSortedUnique,
		},
		{
			"threshold 2, repeated address and repeated sig",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0, 1}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes, sigBytes}},
			&OutputOwners{
				Threshold: 2,
				Addrs:     []ids.ShortID{addr, addr},
			},
			errAddrsNotSortedUnique,
		},
		{
			"threshold 2, 2 sigs",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0, 1}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes, sig2Bytes}},
			&OutputOwners{
				Threshold: 2,
				Addrs:     []ids.ShortID{addr, addr2},
			},
			nil,
		},
		{
			"threshold 2, 2 sigs reversed (should be sorted)",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{1, 0}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sig2Bytes, sigBytes}},
			&OutputOwners{
				Threshold: 2,
				Addrs:     []ids.ShortID{addr, addr2},
			},
			errNotSortedUnique,
		},
		{
			"threshold 1, 1 sig, index out of bounds",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{1}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes}},
			&OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr},
			},
			ErrInputOutputIndexOutOfBounds,
		},
		{
			"too many signers",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0, 1}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes, sig2Bytes}},
			&OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr, addr2},
			},
			ErrTooManySigners,
		},
		{
			"number of signatures doesn't match",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes, sig2Bytes}},
			&OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr, addr2},
			},
			ErrInputCredentialSignersMismatch,
		},
		{
			"output is locked",
			&TestTx{UnsignedBytes: txBytes},
			&Input{SigIndices: []uint32{0}},
			&Credential{Sigs: [][secp256k1.SignatureLen]byte{sigBytes, sig2Bytes}},
			&OutputOwners{
				Threshold: 1,
				Locktime:  uint64(now.Add(time.Second).Unix()),
				Addrs:     []ids.ShortID{addr, addr2},
			},
			ErrTimelocked,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := fx.VerifyPermission(test.tx, test.in, test.cred, test.cg)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}
