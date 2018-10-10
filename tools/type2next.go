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

// File:		type2next.go
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

// This tool outpus Type-2 deterministic addresses, as described here:
// https://bitcointalk.org/index.php?topic=19137.0
// At input it takes "A_public_key" and "B_secret" - both values as hex encoded strings.
// Optionally, you can add a third parameter - number of public keys you want to calculate.
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Specify secret and public_key to get the next Type-2 deterministic address")
		fmt.Println("Add -t as the third argument to work with Testnet addresses.")
		return
	}
	public_key, er := hex.DecodeString(os.Args[2])
	if er != nil {
		println("Error parsing public_key:", er.Error())
		os.Exit(1)
	}

	if len(public_key) == 33 && (public_key[0] == 2 || public_key[0] == 3) {
		fmt.Println("Compressed")
	} else if len(public_key) == 65 && (public_key[0] == 4) {
		fmt.Println("Uncompressed")
	} else {
		println("Incorrect public key")
	}

	secret, er := hex.DecodeString(os.Args[1])
	if er != nil {
		println("Error parsing secret:", er.Error())
		os.Exit(1)
	}

	testnet := len(os.Args) > 3 && os.Args[3] == "-t"

	// Old address
	public_key = bch.DeriveNextPublic(public_key, secret)

	// New address
	fmt.Println(bch.NewAddrFromPubkey(public_key, bch.AddrVerPubkey(testnet)).String())
	// New key
	fmt.Println(hex.EncodeToString(public_key))

}
