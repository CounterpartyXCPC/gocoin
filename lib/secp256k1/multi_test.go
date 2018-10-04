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

// File:        multi_test.go
// Description: Bictoin Cash Cash secp256k1 Package

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

package secp256k1

import (
	"encoding/hex"
	"math/rand"
	"testing"
)

/*
Test strings, that will cause failure
*/

//problem seckeys
var _test_seckey []string = []string{
	"08efb79385c9a8b0d1c6f5f6511be0c6f6c2902963d874a3a4bacc18802528d3",
	"78298d9ecdc0640c9ae6883201a53f4518055442642024d23c45858f45d0c3e6",
	"04e04fe65bfa6ded50a12769a3bd83d7351b2dbff08c9bac14662b23a3294b9e",
	"2f5141f1b75747996c5de77c911dae062d16ae48799052c04ead20ccd5afa113",
}

func RandBytes(n int) []byte {
	b := make([]byte, n, n)

	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(256))
	}
	return b
}

//tests some keys that should work
func Test_Abnormal_Keys1(t *testing.T) {

	for i := 0; i < len(_test_seckey); i++ {

		seckey1, _ := hex.DecodeString(_test_seckey[i])

		pubkey1 := make([]byte, 33)

		ret := BaseMultiply(seckey1, pubkey1)

		if ret == false {
			t.Errorf("base multiplication fail")
		}
		//func BaseMultiply(k, out []byte) bool {

		var pubkey2 XY
		ret = pubkey2.ParsePubkey(pubkey1)
		if ret == false {
			t.Errorf("pubkey parse fail")
		}

		if pubkey2.IsValid() == false {
			t.Errorf("pubkey is not valid")
		}

	}
}

//tests random keys
func Test_Abnormal_Keys2(t *testing.T) {
	for i := 0; i < 64*1024; i++ {

		seckey1 := RandBytes(32)

		pubkey1 := make([]byte, 33)

		ret := BaseMultiply(seckey1, pubkey1)

		if ret == false {
			t.Error("base multiplication fail")
		}
		//func BaseMultiply(k, out []byte) bool {

		var pubkey2 XY
		ret = pubkey2.ParsePubkey(pubkey1)
		if ret == false {
			t.Error("pubkey parse fail")
		}

		if pubkey2.IsValid() == false {
			t.Error("pubkey is not valid for", hex.EncodeToString(seckey1))
		}
	}
}
