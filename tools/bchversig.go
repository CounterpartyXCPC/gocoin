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

// File:		cashaddrr.go
// Description:	Bictoin Cash Cash Adress Package

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

// Package main manages Counterparty Cash (XCPC) nodes. As XCPC transactions are executed
// or queried, the state is maintain in the local LevelDB databstore. Signed RAW transactions
// are parsed to gocoin-cash for transmission to the Bitcoin Cash blockchain.

// This tool is able to verify whether a message was signed with the given bitcoin cash address
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/ltc"
)

// @todo - update this below to reflect BCH
var (
	addr     = flag.String("a", "", "base58 encoded bitcoin address that supposedly signed the message (required)")
	sign     = flag.String("s", "", "base64 encoded signature of the message (required)")
	mess     = flag.String("m", "", "the message (optional)")
	mfil     = flag.String("f", "", "the filename containing a signed message (optional)")
	unix     = flag.Bool("u", false, "remove all \\r characters from the message (optional)")
	help     = flag.Bool("h", false, "print this help")
	verb     = flag.Bool("v", false, "verbose mode")
	litecoin = flag.Bool("ltc", false, "litecoin mode")
)

func main() {
	var msg []byte

	flag.Parse()

	if *help || *addr == "" || *sign == "" {
		flag.PrintDefaults()
		return
	}

	ad, er := btc.NewAddrFromString(*addr)
	if !*litecoin && ad != nil && ad.Version == ltc.AddrVerPubkey(false) {
		*litecoin = true
	}
	if er != nil {
		println("Address:", er.Error())
		flag.PrintDefaults()
		return
	}

	nv, btcsig, er := btc.ParseMessageSignature(*sign)
	if er != nil {
		println("ParseMessageSignature:", er.Error())
		return
	}

	if *mess != "" {
		msg = []byte(*mess)
	} else if *mfil != "" {
		msg, er = ioutil.ReadFile(*mfil)
		if er != nil {
			println(er.Error())
			return
		}
	} else {
		if *verb {
			fmt.Println("Enter the message:")
		}
		msg, _ = ioutil.ReadAll(os.Stdin)
	}

	if *unix {
		if *verb {
			fmt.Println("Enforcing Unix text format")
		}
		msg = bytes.Replace(msg, []byte{'\r'}, nil, -1)
	}

	hash := make([]byte, 32)
	if *litecoin {
		ltc.HashFromMessage(msg, hash)
	} else {
		btc.HashFromMessage(msg, hash)
	}

	compressed := false
	if nv >= 31 {
		if *verb {
			fmt.Println("compressed key")
		}
		nv -= 4
		compressed = true
	}

	pub := btcsig.RecoverPublicKey(hash[:], int(nv-27))
	if pub != nil {
		pk := pub.Bytes(compressed)
		ok := btc.EcdsaVerify(pk, btcsig.Bytes(), hash)
		if ok {
			sa := btc.zx(pk, ad.Version)
			if ad.Hash160 != sa.Hash160 {
				fmt.Println("BAD signature for", ad.String())
				if bytes.IndexByte(msg, '\r') != -1 {
					fmt.Println("You have CR chars in the message. Try to verify with -u switch.")
				}
				os.Exit(1)
			} else {
				fmt.Println("Signature OK")
			}
		} else {
			println("BAD signature")
			os.Exit(1)
		}
	} else {
		println("BAD, BAD, BAD signature")
		os.Exit(1)
	}
}
