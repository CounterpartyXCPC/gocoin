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

// File:		mkmulti.go
// Description:	Bictoin Cash Cash main Package

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
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

/*
{
"address" : "2NAHUDSC1EmbTBwQQp4VQ2FNzWDqHtmk1i6",
"redeemScript" : "512102cdc4fff0ad031ea5f2d0d4337e2bf976b84334f8f80b08fe3f69886d58bc5a8a2102ebf54926d3edaae51bde71f2976948559a8d43fce52f5e7ed9ed85dbaa449d7f52ae"
}
*/
func main() {
	var testnet bool
	if len(os.Args) < 3 {
		fmt.Println("Specify one integer and at least one public key.")
		fmt.Println("For Testent, make the integer negative.")
		return
	}
	cnt, er := strconv.ParseInt(os.Args[1], 10, 32)
	if er != nil {
		println("Count value:", er.Error())
		return
	}
	if cnt < 0 {
		testnet = true
		cnt = -cnt
	}
	if cnt < 1 || cnt > 16 {
		println("The integer (required number of keys) must be between 1 and 16")
		return
	}
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(0x50 + cnt))
	fmt.Println("Trying to prepare multisig address for", cnt, "out of", len(os.Args)-2, "public keys ...")
	var pkeys byte
	var ads string
	for i := 2; i < len(os.Args); i++ {
		if pkeys == 16 {
			println("Oh, give me a break. You don't need more than 16 public keys - stopping here!")
			break
		}
		d, er := hex.DecodeString(os.Args[i])
		if er != nil {
			println("pubkey", i, er.Error())
		}
		_, er = bch.NewPublicKey(d)
		if er != nil {
			println("pubkey", i, er.Error())
			return
		}
		pkeys++
		buf.WriteByte(byte(len(d)))
		buf.Write(d)
		if ads != "" {
			ads += ", "
		}
		ads += "\"" + bch.NewAddrFromPubkey(d, bch.AddrVerPubkey(testnet)).String() + "\""
	}
	buf.WriteByte(0x50 + pkeys)
	buf.WriteByte(0xae)

	p2sh := buf.Bytes()
	addr := bch.NewAddrFromPubkey(p2sh, bch.AddrVerScript(testnet))

	rec := "{\n"
	rec += fmt.Sprintf("\t\"multiAddress\" : \"%s\",\n", addr.String())
	rec += fmt.Sprintf("\t\"scriptPubKey\" : \"a914%s87\",\n", hex.EncodeToString(addr.Hash160[:]))
	rec += fmt.Sprintf("\t\"keysRequired\" : %d,\n", cnt)
	rec += fmt.Sprintf("\t\"keysProvided\" : %d,\n", pkeys)
	rec += fmt.Sprintf("\t\"redeemScript\" : \"%s\",\n", hex.EncodeToString(p2sh))
	rec += fmt.Sprintf("\t\"listOfAddres\" : [%s]\n", ads)
	rec += "}\n"
	fname := addr.String() + ".json"
	ioutil.WriteFile(fname, []byte(rec), 0666)
	fmt.Println("The address record stored in", fname)
}
