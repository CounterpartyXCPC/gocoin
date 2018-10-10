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

// File:        xy.go
// Description: Bictoin Cash Cash secp256k1 Package

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

package secp256k1

import (
	"fmt"
)

type XY struct {
	X, Y     Field
	Infinity bool
}

func (ge *XY) Print(lab string) {
	if ge.Infinity {
		fmt.Println(lab + " - Infinity")
		return
	}
	fmt.Println(lab+".X:", ge.X.String())
	fmt.Println(lab+".Y:", ge.Y.String())
}

func (elem *XY) ParsePubkey(pub []byte) bool {
	if len(pub) == 33 && (pub[0] == 0x02 || pub[0] == 0x03) {
		elem.X.SetB32(pub[1:33])
		elem.SetXO(&elem.X, pub[0] == 0x03)
	} else if len(pub) == 65 && (pub[0] == 0x04 || pub[0] == 0x06 || pub[0] == 0x07) {
		elem.X.SetB32(pub[1:33])
		elem.Y.SetB32(pub[33:65])
		if (pub[0] == 0x06 || pub[0] == 0x07) && elem.Y.IsOdd() != (pub[0] == 0x07) {
			return false
		}
	} else {
		return false
	}
	return true
}

// Returns serialized key in uncompressed format "<04> <X> <Y>"
// ... or in compressed format: "<02> <X>", eventually "<03> <X>"
func (pub *XY) Bytes(compressed bool) (raw []byte) {
	if compressed {
		raw = make([]byte, 33)
		if pub.Y.IsOdd() {
			raw[0] = 0x03
		} else {
			raw[0] = 0x02
		}
		pub.X.GetB32(raw[1:])
	} else {
		raw = make([]byte, 65)
		raw[0] = 0x04
		pub.X.GetB32(raw[1:33])
		pub.Y.GetB32(raw[33:65])
	}
	return
}

func (r *XY) SetXY(X, Y *Field) {
	r.Infinity = false
	r.X = *X
	r.Y = *Y
}

func (a *XY) IsValid() bool {
	if a.Infinity {
		return false
	}
	var y2, x3, c Field
	a.Y.Sqr(&y2)
	a.X.Sqr(&x3)
	x3.Mul(&x3, &a.X)
	c.SetInt(7)
	x3.SetAdd(&c)
	y2.Normalize()
	x3.Normalize()
	return y2.Equals(&x3)
}

func (r *XY) SetXYZ(a *XYZ) {
	var z2, z3 Field
	a.Z.InvVar(&a.Z)
	a.Z.Sqr(&z2)
	a.Z.Mul(&z3, &z2)
	a.X.Mul(&a.X, &z2)
	a.Y.Mul(&a.Y, &z3)
	a.Z.SetInt(1)
	r.Infinity = a.Infinity
	r.X = a.X
	r.Y = a.Y
}

func (a *XY) precomp(w int) (pre []XY) {
	pre = make([]XY, (1 << (uint(w) - 2)))
	pre[0] = *a
	var X, d, tmp XYZ
	X.SetXY(a)
	X.Double(&d)
	for i := 1; i < len(pre); i++ {
		d.AddXY(&tmp, &pre[i-1])
		pre[i].SetXYZ(&tmp)
	}
	return
}

func (a *XY) Neg(r *XY) {
	r.Infinity = a.Infinity
	r.X = a.X
	r.Y = a.Y
	r.Y.Normalize()
	r.Y.Negate(&r.Y, 1)
}

func (r *XY) SetXO(X *Field, odd bool) {
	var c, x2, x3 Field
	r.X = *X
	X.Sqr(&x2)
	X.Mul(&x3, &x2)
	r.Infinity = false
	c.SetInt(7)
	c.SetAdd(&x3)
	c.Sqrt(&r.Y)
	if r.Y.IsOdd() != odd {
		r.Y.Negate(&r.Y, 1)
	}
	r.Y.Normalize()
}

func (pk *XY) AddXY(a *XY) {
	var xyz XYZ
	xyz.SetXY(pk)
	xyz.AddXY(&xyz, a)
	pk.SetXYZ(&xyz)
}

func (pk *XY) GetPublicKey(out []byte) {
	pk.X.Normalize() // See GitHub issue #15
	pk.X.GetB32(out[1:33])
	if len(out) == 65 {
		out[0] = 0x04
		pk.Y.Normalize()
		pk.Y.GetB32(out[33:65])
	} else {
		if pk.Y.IsOdd() {
			out[0] = 0x03
		} else {
			out[0] = 0x02
		}
	}
}
