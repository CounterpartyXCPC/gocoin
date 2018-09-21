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

// File:		key.go
// Description:	Bictoin Cash Key Package

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
	"encoding/hex"
	"errors"

	"github.com/counterpartyxcpc/gocoin-cash/lib/secp256k1"
)

type PublicKey struct {
	secp256k1.XY
}

type Signature struct {
	secp256k1.Signature
	HashType byte
}

func NewPublicKey(buf []byte) (res *PublicKey, e error) {
	res = new(PublicKey)
	if !res.XY.ParsePubkey(buf) {
		e = errors.New("NewPublicKey: Unknown format: " + hex.EncodeToString(buf[:]))
		res = nil
	}
	return
}

func NewSignature(buf []byte) (*Signature, error) {
	sig := new(Signature)
	le := sig.ParseBytes(buf)
	if le < 0 {
		return nil, errors.New("NewSignature: ParseBytes error")
	}
	if le < len(buf) {
		sig.HashType = buf[len(buf)-1]
	}
	return sig, nil
}

// Recoved public key form a signature
func (sig *Signature) RecoverPublicKey(msg []byte, recid int) (key *PublicKey) {
	key = new(PublicKey)
	if !secp256k1.RecoverPublicKey(sig.R.Bytes(), sig.S.Bytes(), msg, recid, &key.XY) {
		key = nil
	}
	return
}

func (sig *Signature) IsLowS() bool {
	return sig.S.Cmp(&secp256k1.TheCurve.HalfOrder.Int) < 1
}

// Returns serialized canoncal signature followed by a hash type
func (sig *Signature) Bytes() []byte {
	return append(sig.Signature.Bytes(), sig.HashType)
}
