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

// File:		txaddsig.go
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

func raw_tx_from_file(fn string) *bch.Tx {
	d, er := ioutil.ReadFile(fn)
	if er != nil {
		fmt.Println(er.Error())
		return nil
	}

	dat, er := hex.DecodeString(string(d))
	if er != nil {
		fmt.Println("hex.DecodeString failed - assume binary transaction file")
		dat = d
	}
	tx, txle := bch.NewTx(dat)

	if tx != nil && txle != len(dat) {
		fmt.Println("WARNING: Raw transaction length mismatch", txle, len(dat))
	}

	return tx
}

func write_tx_file(tx *bch.Tx) {
	signedrawtx := tx.Serialize()
	tx.SetHash(signedrawtx)

	hs := tx.Hash.String()
	fmt.Println(hs)

	f, _ := os.Create(hs[:8] + ".txt")
	if f != nil {
		f.Write([]byte(hex.EncodeToString(signedrawtx)))
		f.Close()
		fmt.Println("Transaction data stored in", hs[:8]+".txt")
	}
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("This tool needs to be executed with 4 arguments:")
		fmt.Println(" 1) Name of the unsigned transaction file")
		fmt.Println(" 2) Input index to add the key & signature to")
		fmt.Println(" 3) Hex dump of the canonical signature")
		fmt.Println(" 4) Hex dump of the public key")
		return
	}
	tx := raw_tx_from_file(os.Args[1])
	if tx == nil {
		return
	}

	in, er := strconv.ParseUint(os.Args[2], 10, 32)
	if er != nil {
		println("Input index:", er.Error())
		return
	}

	if int(in) >= len(tx.TxIn) {
		println("Input index too big:", int(in), "/", len(tx.TxIn))
		return
	}

	sig, er := hex.DecodeString(os.Args[3])
	if er != nil {
		println("Signature:", er.Error())
		return
	}

	pk, er := hex.DecodeString(os.Args[4])
	if er != nil {
		println("Public key:", er.Error())
		return
	}

	buf := new(bytes.Buffer)
	bch.WriteVlen(buf, uint64(len(sig)))
	buf.Write(sig)
	bch.WriteVlen(buf, uint64(len(pk)))
	buf.Write(pk)

	tx.TxIn[in].ScriptSig = buf.Bytes()

	write_tx_file(tx)
}
