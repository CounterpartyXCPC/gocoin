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

// File:        signmsg.go
// Description: Bictoin Cash Cash main Package

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

package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/ltc"
)

// this function signs either a message or a raw hash
func sign_message() {
	var hash []byte
	var signkey *bch.PrivateAddr

	signkey = address_to_key(*signaddr)
	if signkey == nil {
		println("You do not have a private key for", *signaddr)
		return
	}

	if *signhash != "" {
		hash, er := hex.DecodeString(*signhash)
		if er != nil {
			println("Incorrect content of -hash parameter")
			println(er.Error())
			return
		} else if len(hash) > 0 {
			txsig := new(bch.Signature)
			txsig.HashType = 0x01
			r, s, e := bch.EcdsaSign(signkey.Key, hash)
			if e != nil {
				println(e.Error())
				return
			}
			txsig.R.Set(r)
			txsig.S.Set(s)
			fmt.Println("PublicKey:", hex.EncodeToString(signkey.BtcAddr.Pubkey))
			fmt.Println(hex.EncodeToString(txsig.Bytes()))
			return
		}
	}

	var msg []byte
	if *message == "" {
		msg, _ = ioutil.ReadAll(os.Stdin)
	} else {
		msg = []byte(*message)
	}

	hash = make([]byte, 32)
	if litecoin {
		ltc.HashFromMessage(msg, hash)
	} else {
		bch.HashFromMessage(msg, hash)
	}

	btcsig := new(bch.Signature)
	var sb [65]byte
	sb[0] = 27
	if signkey.IsCompressed() {
		sb[0] += 4
	}

	r, s, e := bch.EcdsaSign(signkey.Key, hash)
	if e != nil {
		println(e.Error())
		return
	}
	btcsig.R.Set(r)
	btcsig.S.Set(s)

	rd := btcsig.R.Bytes()
	sd := btcsig.S.Bytes()
	copy(sb[1+32-len(rd):], rd)
	copy(sb[1+64-len(sd):], sd)

	rpk := btcsig.RecoverPublicKey(hash[:], 0)
	sa := bch.NewAddrFromPubkey(rpk.Bytes(signkey.IsCompressed()), signkey.BtcAddr.Version)
	if sa.Hash160 == signkey.BtcAddr.Hash160 {
		fmt.Println(base64.StdEncoding.EncodeToString(sb[:]))
		return
	}

	rpk = btcsig.RecoverPublicKey(hash[:], 1)
	sa = bch.NewAddrFromPubkey(rpk.Bytes(signkey.IsCompressed()), signkey.BtcAddr.Version)
	if sa.Hash160 == signkey.BtcAddr.Hash160 {
		sb[0]++
		fmt.Println(base64.StdEncoding.EncodeToString(sb[:]))
		return
	}
	println("Something went wrong. The message has not been signed.")
}
