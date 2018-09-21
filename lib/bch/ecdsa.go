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

// File:		ecdsa.go
// Description:	Bictoin Cash ECDSA Package

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
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
	"sync/atomic"

	"github.com/counterpartyxcpc/gocoin-cash/lib/secp256k1"
)

var (
	ecdsaVerifyCnt uint64
	EC_Verify      func(k, s, h []byte) bool
)

func EcdsaVerifyCnt() uint64 {
	return atomic.LoadUint64(&ecdsaVerifyCnt)
}

func EcdsaVerify(kd []byte, sd []byte, hash []byte) bool {
	atomic.AddUint64(&ecdsaVerifyCnt, 1)
	if len(kd) == 0 || len(sd) == 0 {
		return false
	}
	if EC_Verify != nil {
		return EC_Verify(kd, sd, hash)
	}
	return secp256k1.Verify(kd, sd, hash)
}

func EcdsaSign(priv, hash []byte) (r, s *big.Int, err error) {
	var sig secp256k1.Signature
	var sec, msg, nonce secp256k1.Number

	sec.SetBytes(priv)
	msg.SetBytes(hash)

	sha := sha256.New()
	sha.Write(priv)
	sha.Write(hash)
	for {
		var buf [32]byte
		rand.Read(buf[:])
		sha.Write(buf[:])
		nonce.SetBytes(sha.Sum(nil))
		if nonce.Sign() > 0 && nonce.Cmp(&secp256k1.TheCurve.Order.Int) < 0 {
			break
		}
	}

	if sig.Sign(&sec, &msg, &nonce, nil) != 1 {
		err = errors.New("ESCDS Sign error()")
	}
	return &sig.R.Int, &sig.S.Int, nil
}
