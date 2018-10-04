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

// File:        field_test.go
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
	"crypto/rand"
	"testing"
)

func TestFeInv(t *testing.T) {
	var in, out, exp Field
	in.SetHex("813925AF112AAB8243F8CCBADE4CC7F63DF387263028DE6E679232A73A7F3C31")
	exp.SetHex("7F586430EA30F914965770F6098E492699C62EE1DF6CAFFA77681C179FDF3117")
	in.Inv(&out)
	if !out.Equals(&exp) {
		t.Error("fe.Inv() failed")
	}
}

func BenchmarkFieldSqrt(b *testing.B) {
	var dat [32]byte
	var f, tmp Field
	rand.Read(dat[:])
	f.SetB32(dat[:])
	for i := 0; i < b.N; i++ {
		f.Sqrt(&tmp)
	}
}

func BenchmarkFieldInv(b *testing.B) {
	var dat [32]byte
	var f, tmp Field
	rand.Read(dat[:])
	f.SetB32(dat[:])
	for i := 0; i < b.N; i++ {
		f.Inv(&tmp)
	}
}
