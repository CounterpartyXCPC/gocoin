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

// File:        ec.go
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
//"encoding/hex"
)

func ecdsa_verify(pubkey, sig, msg []byte) int {
	var m Number
	var s Signature
	m.SetBytes(msg)

	var q XY
	if !q.ParsePubkey(pubkey) {
		return -1
	}

	if s.ParseBytes(sig) < 0 {
		return -2
	}

	if !s.Verify(&q, &m) {
		return 0
	}
	return 1
}

func Verify(k, s, m []byte) bool {
	return ecdsa_verify(k, s, m) == 1
}

func DecompressPoint(X []byte, off bool, Y []byte) {
	var rx, ry, c, x2, x3 Field
	rx.SetB32(X)
	rx.Sqr(&x2)
	rx.Mul(&x3, &x2)
	c.SetInt(7)
	c.SetAdd(&x3)
	c.Sqrt(&ry)
	ry.Normalize()
	if ry.IsOdd() != off {
		ry.Negate(&ry, 1)
	}
	ry.Normalize()
	ry.GetB32(Y)
	return
}

func RecoverPublicKey(r, s, h []byte, recid int, pubkey *XY) bool {
	var sig Signature
	var msg Number
	sig.R.SetBytes(r)
	if sig.R.Sign() <= 0 || sig.R.Cmp(&TheCurve.Order.Int) >= 0 {
		return false
	}
	sig.S.SetBytes(s)
	if sig.S.Sign() <= 0 || sig.S.Cmp(&TheCurve.Order.Int) >= 0 {
		return false
	}
	msg.SetBytes(h)
	if !sig.recover(pubkey, &msg, recid) {
		return false
	}
	return true
}

// Standard EC multiplacation k(xy)
// xy - is the standarized public key format (33 or 65 bytes long)
// out - should be the buffer for 33 bytes (1st byte will be set to either 02 or 03)
func Multiply(xy, k, out []byte) bool {
	var pk XY
	var xyz XYZ
	var na, nzero Number
	if !pk.ParsePubkey(xy) {
		return false
	}
	xyz.SetXY(&pk)
	na.SetBytes(k)
	xyz.ECmult(&xyz, &na, &nzero)
	pk.SetXYZ(&xyz)
	pk.GetPublicKey(out)
	return true
}

// Multiply k by G
// out - should be the buffer for 33 bytes (1st byte will be set to either 02 or 03)
func BaseMultiply(k, out []byte) bool {
	var r XYZ
	var n Number
	var pk XY
	n.SetBytes(k)
	ECmultGen(&r, &n)
	pk.SetXYZ(&r)
	pk.GetPublicKey(out)
	return true
}

// out = G*k + xy
func BaseMultiplyAdd(xy, k, out []byte) bool {
	var r XYZ
	var n Number
	var pk XY
	if !pk.ParsePubkey(xy) {
		return false
	}
	n.SetBytes(k)
	ECmultGen(&r, &n)
	r.AddXY(&r, &pk)
	pk.SetXYZ(&r)
	pk.GetPublicKey(out)
	return true
}
