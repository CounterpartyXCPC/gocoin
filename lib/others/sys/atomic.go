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

// File:        atomic.go
// Description: Bictoin Cash Cash sys Package

// Credits:

// Piotr Narewski, Gocoin Founder

// Julian Smith, Direction + Development
// Arsen Yeremin, Development
// Sumanth Kumar, Development
// Clayton Wong, Development
// Liming Jiang, Development

// Includes reference work of btsuite:

// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2018 The bcext developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Credits:

// Piotr Narewski, Gocoin Founder

// Julian Smith, Direction + Development
// Arsen Yeremin, Development
// Sumanth Kumar, Development
// Clayton Wong, Development
// Liming Jiang, Development

// Includes reference work of btsuite:

// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2018 The bcext developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Includes reference work of Bitcoin Core (https://github.com/bitcoin/bitcoin)
// Includes reference work of Bitcoin-ABC (https://github.com/Bitcoin-ABC/bitcoin-abc)
// Includes reference work of Bitcoin Unlimited (https://github.com/BitcoinUnlimited/BitcoinUnlimited/tree/BitcoinCash)
// Includes reference work of gcash by Shuai Qi "qshuai" (https://github.com/bcext/gcash)
// Includes reference work of gcash (https://github.com/gcash/bchd)

// + Other contributors

// =====================================================================

package sys

import (
	"fmt"
	"sync/atomic"
)

type SyncBool struct {
	val int32
}

func (b *SyncBool) Get() bool {
	return atomic.LoadInt32(&b.val) != 0
}

func (b *SyncBool) Set() {
	atomic.StoreInt32(&b.val, 1)
}

func (b *SyncBool) Clr() {
	atomic.StoreInt32(&b.val, 0)
}

func (b *SyncBool) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprint(b.Get())), nil
}

func (b *SyncBool) Store(val bool) {
	if val {
		b.Set()
	} else {
		b.Clr()
	}
}

type SyncInt struct {
	val int64
}

func (b *SyncInt) Get() int {
	return int(atomic.LoadInt64(&b.val))
}

func (b *SyncInt) Store(val int) {
	atomic.StoreInt64(&b.val, int64(val))
}

func (b *SyncInt) Add(val int) {
	atomic.AddInt64(&b.val, int64(val))
}

func (b *SyncInt) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprint(b.Get())), nil
}
