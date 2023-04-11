// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package units

// Denominations of value
const (
	NanoVidar  uint64 = 1
	MicroVidar uint64 = 1000 * NanoVidar
	Schmeckle uint64 = 49*MicroVidar + 463*NanoVidar
	MilliVidar uint64 = 1000 * MicroVidar
	Vidar      uint64 = 1000 * MilliVidar
	KiloVidar  uint64 = 1000 * Vidar
	MegaVidar  uint64 = 1000 * KiloVidar
)
