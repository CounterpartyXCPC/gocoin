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

// File:		target_test.go
// Description:	Bictoin Cash Target Package Testing

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

package bch

import (
	//	"fmt"
	"math"
	"math/big"
	"testing"
)

type onevec struct {
	b uint32
	e string
	d float64
}

var testvecs = []onevec{
	{b: 0x1b0404cb, e: "00000000000404CB000000000000000000000000000000000000000000000000"},
	{b: 0x1d00ffff, e: "00000000FFFF0000000000000000000000000000000000000000000000000000"},
	{b: 436330132, d: 8974296.01488785},
	{b: 436543292, d: 3275464.59},
	{b: 436591499, d: 2864140.51},
	{b: 436841986, d: 1733207.51},
	{b: 437155514, d: 1159929.50},
	{b: 436789733, d: 1888786.71},
	{b: 453031340, d: 92347.59},
	{b: 453281356, d: 14484.16},
	{b: 470771548, d: 16.62},
	{b: 486604799, d: 1.00},
}

func TestTarget(t *testing.T) {
	for i := range testvecs {
		x := SetCompact(testvecs[i].b)
		d := GetDifficulty(testvecs[i].b)

		c := GetCompact(x)
		//fmt.Printf("%d. %d/%d -> %.8f / %.8f\n", i, testvecs[i].b, c, d, testvecs[i].d)
		if testvecs[i].b != c {
			t.Error("Set/GetCompact mismatch at alement", i)
		}

		if testvecs[i].e != "" {
			y, _ := new(big.Int).SetString(testvecs[i].e, 16)
			if x.Cmp(y) != 0 {
				t.Error("Target mismatch at alement", i)
			}
		}

		if testvecs[i].d != 0 && math.Abs(d-testvecs[i].d) > 0.1 {
			t.Error("Difficulty mismatch at alement", i)
		}
	}
}
