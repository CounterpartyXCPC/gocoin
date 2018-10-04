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

// File:		bech32_test.go
// Description:	Bictoin Cash bech32 Package

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

package bech32

import (
	"strings"
	"testing"
)

var (
	valid_checksum = []string{
		"A12UEL5L",
		"an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs",
		"abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
		"11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
		"split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w"}

	invalid_checksum = []string{
		" 1nwldj5",
		"\x7f1axkwrx",
		"an84characterslonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1569pvx",
		"pzry9x0s0muk",
		"1pzry9x0s0muk",
		"x1b4n0q5v",
		"li1dgmt3",
		"de1lg7wt\xff"}
)

func TestValidChecksum(t *testing.T) {
	for _, s := range valid_checksum {
		hrp, data := Decode(s)
		if data == nil || hrp == "" {
			t.Error("Decode fails: ", s)
		} else {
			rebuild := Encode(hrp, data)
			if rebuild == "" {
				t.Error("Encode fails: ", s)
			} else {
				if !strings.EqualFold(s, rebuild) {
					t.Error("Encode produces incorrect result: ", s)
				}
			}
		}
	}
}

func TestInvalidChecksum(t *testing.T) {
	for _, s := range invalid_checksum {
		hrp, data := Decode(s)
		if data != nil || hrp != "" {
			t.Error("Decode succeeds on invalid string: ", s)
		}
	}
}
