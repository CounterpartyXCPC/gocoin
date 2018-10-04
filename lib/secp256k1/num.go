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

// File:        num.go
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
	"fmt"
	"math/big"
)

var (
	BigInt1 *big.Int = new(big.Int).SetInt64(1)
)

type Number struct {
	big.Int
}

func (a *Number) Print(label string) {
	fmt.Println(label, hex.EncodeToString(a.Bytes()))
}

func (r *Number) mod_mul(a, b, m *Number) {
	r.Mul(&a.Int, &b.Int)
	r.Mod(&r.Int, &m.Int)
	return
}

func (r *Number) mod_inv(a, b *Number) {
	r.ModInverse(&a.Int, &b.Int)
	return
}

func (r *Number) mod(a *Number) {
	r.Mod(&r.Int, &a.Int)
	return
}

func (a *Number) SetHex(s string) {
	a.SetString(s, 16)
}

func (num *Number) mask_bits(bits uint) {
	mask := new(big.Int).Lsh(BigInt1, bits)
	mask.Sub(mask, BigInt1)
	num.Int.And(&num.Int, mask)
}

func (a *Number) split_exp(r1, r2 *Number) {
	var bnc1, bnc2, bnn2, bnt1, bnt2 Number

	bnn2.Int.Rsh(&TheCurve.Order.Int, 1)

	bnc1.Mul(&a.Int, &TheCurve.a1b2.Int)
	bnc1.Add(&bnc1.Int, &bnn2.Int)
	bnc1.Div(&bnc1.Int, &TheCurve.Order.Int)

	bnc2.Mul(&a.Int, &TheCurve.b1.Int)
	bnc2.Add(&bnc2.Int, &bnn2.Int)
	bnc2.Div(&bnc2.Int, &TheCurve.Order.Int)

	bnt1.Mul(&bnc1.Int, &TheCurve.a1b2.Int)
	bnt2.Mul(&bnc2.Int, &TheCurve.a2.Int)
	bnt1.Add(&bnt1.Int, &bnt2.Int)
	r1.Sub(&a.Int, &bnt1.Int)

	bnt1.Mul(&bnc1.Int, &TheCurve.b1.Int)
	bnt2.Mul(&bnc2.Int, &TheCurve.a1b2.Int)
	r2.Sub(&bnt1.Int, &bnt2.Int)
}

func (a *Number) split(rl, rh *Number, bits uint) {
	rl.Int.Set(&a.Int)
	rh.Int.Rsh(&rl.Int, bits)
	rl.mask_bits(bits)
}

func (num *Number) rsh(bits uint) {
	num.Rsh(&num.Int, bits)
}

func (num *Number) inc() {
	num.Add(&num.Int, BigInt1)
}

func (num *Number) rsh_x(bits uint) (res int) {
	res = int(new(big.Int).And(&num.Int, new(big.Int).SetUint64((1<<bits)-1)).Uint64())
	num.Rsh(&num.Int, bits)
	return
}

func (num *Number) IsOdd() bool {
	return num.Bit(0) != 0
}

func (num *Number) get_bin(le int) []byte {
	bts := num.Bytes()
	if len(bts) > le {
		panic("buffer too small")
	}
	if len(bts) == le {
		return bts
	}
	return append(make([]byte, le-len(bts)), bts...)
}
