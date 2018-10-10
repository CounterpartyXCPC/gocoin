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

// File:		unspent4_to_UTXO.go
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
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/qdb"
)

var (
	block_height uint64
	block_hash   []byte
)

func load_map4() (ndb map[qdb.KeyType][]byte) {
	var odb *qdb.DB
	ndb = make(map[qdb.KeyType][]byte, 21e6)
	for i := 0; i < 16; i++ {
		fmt.Print("\r", i, " of 16 ... ")
		er := qdb.NewDBExt(&odb, &qdb.NewDBOpts{Dir: fmt.Sprintf("unspent4/%06d", i),
			Volatile: true, LoadData: true, WalkFunction: func(key qdb.KeyType, val []byte) uint32 {
				if _, ok := ndb[key]; ok {
					panic("duplicate")
				}
				ndb[key] = val
				return 0
			}})
		if er != nil {
			fmt.Println(er.Error())
			return
		}
		odb.Close()
	}
	fmt.Print("\r                                                              \r")
	return
}

func load_last_block() {
	var maxbl_fn string

	fis, _ := ioutil.ReadDir("unspent4/")
	var maxbl, undobl int
	for _, fi := range fis {
		if !fi.IsDir() && fi.Size() >= 32 {
			ss := strings.SplitN(fi.Name(), ".", 2)
			cb, er := strconv.ParseUint(ss[0], 10, 32)
			if er == nil && int(cb) > maxbl {
				maxbl = int(cb)
				maxbl_fn = fi.Name()
				if len(ss) == 2 && ss[1] == "tmp" {
					undobl = maxbl
				}
			}
		}
	}
	if maxbl == 0 {
		fmt.Println("This unspent4 database is corrupt")
		return
	}
	if undobl == maxbl {
		fmt.Println("This unspent4 database is not properly closed")
		return
	}

	block_height = uint64(maxbl)
	block_hash = make([]byte, 32)

	f, _ := os.Open("unspent4/" + maxbl_fn)
	f.Read(block_hash)
	f.Close()

}

func save_map(ndb map[qdb.KeyType][]byte) {
	var cnt_dwn, cnt_dwn_from, perc int
	of, er := os.Create("UTXO.db")
	if er != nil {
		fmt.Println("Create file:", er.Error())
		return
	}

	cnt_dwn_from = len(ndb) / 100
	wr := bufio.NewWriter(of)
	binary.Write(wr, binary.LittleEndian, uint64(block_height))
	wr.Write(block_hash)
	binary.Write(wr, binary.LittleEndian, uint64(len(ndb)))
	for k, v := range ndb {
		bch.WriteVlen(wr, uint64(len(v)+8))
		binary.Write(wr, binary.LittleEndian, k)
		//binary.Write(wr, binary.LittleEndian, uint32(len(v)))
		_, er = wr.Write(v)
		if er != nil {
			fmt.Println("\n\007Fatal error:", er.Error())
			break
		}
		if cnt_dwn == 0 {
			fmt.Print("\rSaving UTXO.db - ", perc, "% complete ... ")
			cnt_dwn = cnt_dwn_from
			perc++
		} else {
			cnt_dwn--
		}
	}
	wr.Flush()
	of.Close()

	fmt.Print("\r                                                              \r")
}

func main() {
	var sta time.Time

	if fi, er := os.Stat("unspent4"); er != nil || !fi.IsDir() {
		fmt.Println("ERROR: Input database not found.")
		fmt.Println("Make sure to have unspent4/ directory, where you run this tool from")
		return
	}

	load_last_block()
	if len(block_hash) != 32 {
		fmt.Println("ERROR: Could not recover last block's data from the input database", len(block_hash))
		return
	}

	fmt.Println("Loading input database. Block", block_height, bch.NewUint256(block_hash).String())
	sta = time.Now()
	ndb := load_map4()
	fmt.Println(len(ndb), "records loaded in", time.Now().Sub(sta).String())

	sta = time.Now()
	save_map(ndb)
	fmt.Println("Saved in in", time.Now().Sub(sta).String())
}
