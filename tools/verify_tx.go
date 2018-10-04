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

// File:		verify_tx.go
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

// +build windows

// On Windows OS copy this file to gocoin\client\usif\textui to enable consensus checking
// Make sure you have proper "libbitcoinconsensus-0.dll" in a folder where OS can find it.

package main

import (
	"encoding/hex"
	"fmt"
	"syscall"
	"unsafe"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/script"
)

const (
	DllName  = "libbitcoinconsensus-0.dll"
	ProcName = "bitcoinconsensus_verify_script"
)

var (
	bitcoinconsensus_verify_script *syscall.Proc
	use_consensus_lib              bool
)

func consensus_verify_script(pkScr []byte, i int, tx *bch.Tx, ver_flags uint32) bool {
	txTo := tx.Serialize()

	var pkscr_ptr, pkscr_len uintptr // default to 0/null
	if pkScr != nil {
		pkscr_ptr = uintptr(unsafe.Pointer(&pkScr[0]))
		pkscr_len = uintptr(len(pkScr))
	}
	r1, _, _ := syscall.Syscall9(bitcoinconsensus_verify_script.Addr(), 7,
		pkscr_ptr, pkscr_len,
		uintptr(unsafe.Pointer(&txTo[0])), uintptr(len(txTo)),
		uintptr(i), uintptr(ver_flags), 0, 0, 0)

	return r1 == 1
}

func load_dll() {
	dll, er := syscall.LoadDLL(DllName)
	if er != nil {
		println(er.Error())
		println("WARNING: Consensus verificatrion disabled")
		return
	}
	bitcoinconsensus_verify_script, er = dll.FindProc(ProcName)
	if er != nil {
		println(er.Error())
		println("WARNING: Consensus verificatrion disabled")
		return
	}
	fmt.Println("Using", DllName, "to ensure consensus rules")
	use_consensus_lib = true
}

func main() {
	load_dll()
	pkscript, _ := hex.DecodeString("76a9147d22f6c9cca35cb4071971fe442da58546aaeb5988ac")
	d, _ := hex.DecodeString("0100000002232e0afdd9bcad5e3ace8a19ab8ad0ed8cebd6213b098e36cdc8b25af1d5cd30010000006b483045022077768255f192427bd2841555cfc86fdb7332e18c5c530b3b6028cd034a339f9c022100b3876037f63559ca8a2766a86c8dc62d41c869abc539ab983ce8eccf448f117f012102a33ac1e78cd3ff49bde292da2efcf273509d0869fe81571dfb49528c8287a8fcffffffff2fc90cf473e6ce6177818f705f6e96c7ad42f921f23b660ea27f653346e6a8a9010000006a47304402206d5be8061f712fba560b9966e037f7c53cff377b0c15d8c62bd0a2bcb195048602200522601341cdf574e3a39ba0397d8fe5608e37fd46b3fda2684386207ca9bf69012102a33ac1e78cd3ff49bde292da2efcf273509d0869fe81571dfb49528c8287a8fcffffffff0200a86100000000001976a914ff8e92b694527dd77660f873eb3a86eda5ed459f88ac70110100000000001976a9147d22f6c9cca35cb4071971fe442da58546aaeb5988ac00000000")
	tx, _ := bch.NewTx(d)
	i := 0
	value := uint64(1000000)
	flags := uint32(script.STANDARD_VERIFY_FLAGS)
	println(flags)
	res := script.VerifyTxScript(pkscript, value, i, tx, flags)
	println("Gocoin:", res)
	if use_consensus_lib {
		res = consensus_verify_script(pkscript, i, tx, flags)
		println("Consen:", res)
	}
}
