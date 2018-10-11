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

// File:		importblocks.go
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

// This tool can import blockchain database from satoshi client to gocoin
package main

import (
	"fmt"
	"os"
	"time"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/blockdb"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

const Trust = true // Set this to false if you want to re-check all scripts

var (
	Magic               [4]byte
	GocoinCashHomeDir   string
	BtcRootDir          string
	GenesisBlock        *bch.Uint256
	prev_EcdsaVerifyCnt uint64
)

func stat(totnsec, pernsec int64, totbytes, perbytes uint64, height uint32) {
	totmbs := float64(totbytes) / (1024 * 1024)
	perkbs := float64(perbytes) / (1024)
	var x string
	cn := bch.EcdsaVerifyCnt() - prev_EcdsaVerifyCnt
	if cn > 0 {
		x = fmt.Sprintf("|  %d -> %d us/ecdsa", cn, uint64(pernsec)/cn/1e3)
		prev_EcdsaVerifyCnt += cn
	}
	fmt.Printf("%.1fMB of data processed. We are at height %d. Processing speed %.3fMB/sec, recent: %.1fKB/s %s\n",
		totmbs, height, totmbs/(float64(totnsec)/1e9), perkbs/(float64(pernsec)/1e9), x)
}

func import_blockchain(dir string) {
	BlockDatabase := blockdb.NewBchBlockDB(dir, Magic)
	chain := bch_chain.NewChainExt(GocoinCashHomeDir, GenesisBlock, false, nil, nil)

	var bl *bch.BchBlock
	var er error
	var dat []byte
	var totbytes, perbytes uint64

	fmt.Println("Be patient while importing Satoshi's database... ")
	start := time.Now().UnixNano()
	prv := start
	for {
		now := time.Now().UnixNano()
		if now-prv >= 10e9 {
			stat(now-start, now-prv, totbytes, perbytes, bch_chain.LastBlock().Height)
			prv = now // show progress each 10 seconds
			perbytes = 0
		}

		dat, er = BlockDatabase.FetchNextBlock()
		if dat == nil || er != nil {
			println("END of DB file")
			break
		}

		bl, er = bch.NewBchBlock(dat[:])
		if er != nil {
			println("Block inconsistent:", er.Error())
			break
		}

		bl.Trusted = Trust

		er, _, _ = bch_chain.CheckBlock(bl)

		if er != nil {
			if er.Error() != "Genesis" {
				println("CheckBlock failed:", er.Error())
				os.Exit(1) // Such a thing should not happen, so let's better abort here.
			}
			continue
		}

		er = bch_chain.AcceptBlock(bl)
		if er != nil {
			println("AcceptBlock failed:", er.Error())
			os.Exit(1) // Such a thing should not happen, so let's better abort here.
		}

		totbytes += uint64(len(bl.Raw))
		perbytes += uint64(len(bl.Raw))
	}

	stop := time.Now().UnixNano()
	stat(stop-start, stop-prv, totbytes, perbytes, bch_chain.LastBlock().Height)

	fmt.Println("Satoshi's database import finished in", (stop-start)/1e9, "seconds")

	fmt.Println("Now saving the new database...")
	chain.Close()
	fmt.Println("Database saved. No more imports should be needed.")
}

func RemoveLastSlash(p string) string {
	if len(p) > 0 && os.IsPathSeparator(p[len(p)-1]) {
		return p[:len(p)-1]
	}
	return p
}

func exists(fn string) bool {
	_, e := os.Lstat(fn)
	return e == nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Specify at least one parameter - a path to the blk0000?.dat files.")
		fmt.Println("By default it should be:", sys.BitcoinHome()+"blocks")
		fmt.Println()
		fmt.Println("If you specify a second parameter, that's where output data will be stored.")
		fmt.Println("Otherwise the output data will go to Gocoin's default data folder.")
		return
	}

	BtcRootDir = RemoveLastSlash(os.Args[1])
	fn := BtcRootDir + string(os.PathSeparator) + "blk00000.dat"
	fmt.Println("Looking for file", fn, "...")
	f, e := os.Open(fn)
	if e != nil {
		println(e.Error())
		os.Exit(1)
	}
	_, e = f.Read(Magic[:])
	f.Close()
	if e != nil {
		println(e.Error())
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		GocoinCashHomeDir = RemoveLastSlash(os.Args[2]) + string(os.PathSeparator)
	} else {
		GocoinCashHomeDir = sys.BitcoinHome() + "gocoin" + string(os.PathSeparator)
	}

	if Magic == [4]byte{0x0B, 0x11, 0x09, 0x07} {
		// testnet3
		fmt.Println("There are Testnet3 blocks")
		GenesisBlock = bch.NewUint256FromString("000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943")
		GocoinCashHomeDir += "tstnet" + string(os.PathSeparator)
	} else if Magic == [4]byte{0xF9, 0xBE, 0xB4, 0xD9} {
		fmt.Println("There are valid Bitcoin blocks")
		GenesisBlock = bch.NewUint256FromString("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
		GocoinCashHomeDir += "bchnet" + string(os.PathSeparator)
	} else {
		println("blk00000.dat has an unexpected magic")
		os.Exit(1)
	}

	fmt.Println("Importing blockchain data into", GocoinCashHomeDir, "...")

	if exists(GocoinCashHomeDir+"blockchain.dat") ||
		exists(GocoinCashHomeDir+"blockchain.idx") ||
		exists(GocoinCashHomeDir+"unspent") {
		println("Destination folder contains some database files.")
		println("Either move them somewhere else or delete manually.")
		println("None of the following files/folders must exist before you proceed:")
		println(" *", GocoinCashHomeDir+"blockchain.dat")
		println(" *", GocoinCashHomeDir+"blockchain.idx")
		println(" *", GocoinCashHomeDir+"unspent")
		os.Exit(1)
	}

	import_blockchain(BtcRootDir)
}
