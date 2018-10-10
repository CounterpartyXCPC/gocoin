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

// File:		block_test.go
// Description:	Bictoin Cash Block Test Package

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

package bch

import (
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

const block_hash = "0000000000000000000884ad62c7036a7e2022bca3f0bd68628414150e8a0ea6"

var _block_filename = ""

func block_filename() string {
	if _block_filename == "" {
		_block_filename = os.TempDir() + string(os.PathSeparator) + block_hash
	}
	return _block_filename
}

// Download block from blockchain.info and store it in the TEMP folder
func fetch_block(b *testing.B) {
	url := "https://blockchain.info/block/" + block_hash + "?format=hex"
	r, er := http.Get(url)
	if er == nil {
		if r.StatusCode == 200 {
			rawhex, er := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if er == nil {
				raw, er := hex.DecodeString(string(rawhex))
				if er == nil {
					er = ioutil.WriteFile(block_filename(), raw, 0600)
				}
			}
		} else {
			b.Fatal("Unexpected HTTP Status code", r.StatusCode, url)
		}
	} else {
		b.Fatal(er.Error())
	}
	return
}

func BenchmarkBuildTxList(b *testing.B) {
	raw, e := ioutil.ReadFile(block_filename())
	if e != nil {
		fetch_block(b)
		if raw, e = ioutil.ReadFile(block_filename()); e != nil {
			b.Fatal(e.Error())
		}
	}
	b.SetBytes(int64(len(raw)))
	bl, e := NewBchBlock(raw)
	if e != nil {
		b.Fatal(e.Error())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bl.TxCount = 0
		bl.BuildTxList()
	}
}

func BenchmarkCalcMerkle(b *testing.B) {
	raw, e := ioutil.ReadFile(block_filename())
	if e != nil {
		fetch_block(b)
		if raw, e = ioutil.ReadFile(block_filename()); e != nil {
			b.Fatal(e.Error())
		}
	}
	bl, e := NewBchBlock(raw)
	if e != nil {
		b.Fatal(e.Error())
	}
	bl.BuildTxList()
	mtr := make([][32]byte, len(bl.Txs), 3*len(bl.Txs)) // make the buffer 3 times longer as we use append() inside CalcMerkle
	for i, tx := range bl.Txs {
		mtr[i] = tx.Hash.Hash
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcMerkle(mtr)
	}
}
