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

// File:		funcs_test.go
// Description:	Bictoin Cash Function Package Testing

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

package bch

import (
	"testing"
)

func TestParseAmount(t *testing.T) {
	var tv = []struct {
		af string
		ai uint64
	}{
		{"84.3449", 8434490000},
		{"84.3448", 8434480000},
		{"84.3447", 8434470000},
		{"84.3446", 8434460000},
		{"84.3445", 8434450000},
		{"84.3444", 8434440000},
		{"84.3443", 8434430000},
		{"84.3442", 8434420000},
		{"84.3441", 8434410000},
		{"84.3440", 8434400000},
		{"84.3439", 8434390000},
		{"0.99999990", 99999990},
		{"0.99999991", 99999991},
		{"0.99999992", 99999992},
		{"0.99999993", 99999993},
		{"0.99999994", 99999994},
		{"0.99999995", 99999995},
		{"0.99999996", 99999996},
		{"0.99999997", 99999997},
		{"0.99999998", 99999998},
		{"0.99999999", 99999999},
		{"1.00000001", 100000001},
		{"1.00000002", 100000002},
		{"1.00000003", 100000003},
		{"1.00000004", 100000004},
		{"1.00000005", 100000005},
		{"1.00000006", 100000006},
		{"1.00000007", 100000007},
		{"1000000.0", 100000000000000},
		{"100000.0", 10000000000000},
		{"10000.0", 1000000000000},
		{"1000.0", 100000000000},
		{"100.0", 10000000000},
		{"10.0", 1000000000},
		{"1.0", 100000000},
		{"0.1", 10000000},
		{"0.01", 1000000},
		{"0.001", 100000},
		{"0.00001", 1000},
		{"0.000001", 100},
		{"0.0000001", 10},
		{"0.00000001", 1},
	}
	for i := range tv {
		res, _ := StringToSatoshis(tv[i].af)
		if res != tv[i].ai {
			t.Error("Mismatch at index", i, tv[i].af, res, tv[i].ai)
		}
	}
}
