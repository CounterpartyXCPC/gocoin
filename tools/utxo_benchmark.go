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

// File:		utxo_benchmark.go
// Description:	Bictoin Cash Cash main Package

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

package main

import (
	"encoding/binary"
	"os"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

func main() {
	var tmp uint32
	var dir = ""

	println("UtxoIdxLen:", utxo.UtxoIdxLen)
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if len(os.Args) < 3 {
		utxo.MembindInit()
	} else {
		println("Using native Go heap for UTXO records")
	}

	sta := time.Now()
	db := utxo.NewUnspentDb(&utxo.NewUnspentOpts{Dir: dir})
	if db == nil {
		println("place UTXO.db or UTXO.old in the current folder")
		return
	}

	println(len(db.HashMap), "UTXO records/txs loaded in", time.Now().Sub(sta).String())

	print("Going through the map...")
	sta = time.Now()
	for k, v := range db.HashMap {
		if v != nil {
			tmp += binary.LittleEndian.Uint32(k[:])
		}
	}
	tim := time.Now().Sub(sta)
	println("\rGoing through the map done in", tim.String(), tmp)

	print("Going through the map for the slice...")
	tmp = 0
	sta = time.Now()
	for _, v := range db.HashMap {
		tmp += binary.LittleEndian.Uint32(v)
	}
	println("\rGoing through the map for the slice done in", time.Now().Sub(sta).String(), tmp)

	print("Decoding all records in static mode ...")
	tmp = 0
	sta = time.Now()
	for k, v := range db.HashMap {
		tmp += utxo.NewUtxoRecStatic(k, v).InBlock
	}
	println("\rDecoding all records in static mode done in", time.Now().Sub(sta).String(), tmp)

	print("Decoding all records in dynamic mode ...")
	tmp = 0
	sta = time.Now()
	for k, v := range db.HashMap {
		tmp += utxo.NewUtxoRec(k, v).InBlock
	}
	println("\rDecoding all records in dynamic mode done in", time.Now().Sub(sta).String(), tmp)

	al, sy := sys.MemUsed()
	println("Mem Used:", al>>20, "/", sy>>20)
}
