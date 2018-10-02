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

// File:		unspent_test.go
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
	"encoding/hex"
	"testing"
)

const (
	UtxoRecord = "B26B877AF9D16E5F634C4997A8393C9496BAA14C34D73829767723D96D4AE368FE19AC0700060100166A146F6D6E69000000000000001F0000008B3B93DC0002FD22021976A914A25DEC4D0011064EF106A983C39C7A540699F22088AC"
	//UtxoRecord = "875207AE844E25A60BB57C7E68FDEA8C3BD04FBF678866EF3E7E9FDD408B9E98FEF07A06000401FD60EA17A914379238E99325F2BD2D1F773B8D95CFB9EA92C31887"
)

func BenchmarkFullUtxoRec(b *testing.B) {
	raw, _ := hex.DecodeString(UtxoRecord)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if FullUtxoRec(raw) == nil {
			b.Fatal("Nil pointer returned")
		}
	}
}

func BenchmarkNewUtxoRec(b *testing.B) {
	raw, _ := hex.DecodeString(UtxoRecord)
	var key UtxoKeyType
	copy(key[:], raw[:])
	dat := raw[UtxoIdxLen:]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if NewUtxoRec(key, dat) == nil {
			b.Fatal("Nil pointer returned")
		}
	}
}

func BenchmarkNewUtxoRecStatic(b *testing.B) {
	raw, _ := hex.DecodeString(UtxoRecord)
	var key UtxoKeyType
	copy(key[:], raw[:])
	dat := raw[UtxoIdxLen:]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if NewUtxoRec(key, dat) == nil {
			b.Fatal("Nil pointer returned")
		}
	}
}

func TestMembinds(t *testing.T) {
	MembindInit()
	ptr := malloc(0x100000)
	for i := range ptr {
		ptr[i] = byte(i)
	}
	free(ptr)
}
