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

// File:		membind.go
// Description:	Bictoin Cash utxo Package

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

package utxo

import (
	"sync/atomic"
)

var (
	malloc func(le uint32) []byte = func(le uint32) []byte {
		return make([]byte, int(le))
	}

	free func([]byte) = func(v []byte) {
	}

	malloc_and_copy func(v []byte) []byte = func(v []byte) []byte {
		return v
	}

	MembindInit func() = func() {}
)

var (
	extraMemoryConsumed int64 // if we are using the glibc memory manager
	extraMemoryAllocCnt int64 // if we are using the glibc memory manager
)

func ExtraMemoryConsumed() int64 {
	return atomic.LoadInt64(&extraMemoryConsumed)
}

func ExtraMemoryAllocCnt() int64 {
	return atomic.LoadInt64(&extraMemoryAllocCnt)
}
