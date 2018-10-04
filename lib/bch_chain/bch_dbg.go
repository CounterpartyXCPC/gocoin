// ======================================================================

//      cccccccccc          pppppppppp
//    cccccccccccccc      pppppppppppppp
//  ccccccccccccccc    ppppppppppppppppppp
// cccccc       cc    ppppppp        pppppp
// cccccc          pppppppp          pppppp
// cccccc        ccccpppp            pppppp
// cccccccc    cccccccc    pppp    ppppppp
//  ccccccccccccccccc     ppppppppppppppp
//     cccccccccccc      pppppppppppppp
//       cccccccc        pppppppppppp
//                       pppppp
//                       pppppp

// ======================================================================
// Copyright © 2018. Counterparty Cash Association (CCA) Zug, CH.
// All Rights Reserved. All work owned by CCA is herby released
// under Creative Commons Zero (0) License.

// Some rights of 3rd party, derivative and included works remain the
// property of thier respective owners. All marks, brands and logos of
// member groups remain the exclusive property of their owners and no
// right or endorsement is conferred by reference to thier organization
// or brand(s) by CCA.

// File:		bch_dbg.go
// Description:	Bictoin Cash bch_chain Package

// Credits:

// Julian Smith, Direction, Development
// Arsen Yeremin, Development
// Sumanth Kumar, Development
// Clayton Wong, Development
// Liming Jiang, Development
// Piotr Narewski, Gocoin Founder

// Includes reference work of Shuai Qi "qshuai" (https://github.com/qshuai)

// Includes reference work of btsuite:

// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2018 The bcext developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// + Other contributors

// =====================================================================

package bch_chain

const (
	DBG_WASTED  = 1 << 0
	DBG_UNSPENT = 1 << 1
	DBG_BLOCKS  = 1 << 2
	DBG_ORPHAS  = 1 << 3
	DBG_TX      = 1 << 4
	DBG_SCRIPT  = 1 << 5
	DBG_VERIFY  = 1 << 6
	DBG_SCRERR  = 1 << 7
)

var dbgmask uint32 = 0

func don(b uint32) bool {
	return (dbgmask & b) != 0
}

func DbgSwitch(b uint32, on bool) {
	if on {
		dbgmask |= b
	} else {
		dbgmask ^= (b & dbgmask)
	}
}
