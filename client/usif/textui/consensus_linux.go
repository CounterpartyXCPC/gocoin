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

// File:		consensus_linux.go
// Description:	Bictoin Cash textui Package

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

// +build linux

// Place the bitcoin consensus lib (libbitcoinconsensus.so) where OS can find it.
// If this file does not build and you don't know what to do, just delete it

package textui

/*
#cgo LDFLAGS: -ldl

#include <stdio.h>
#include <dlfcn.h>


typedef signed long long int64_t;

unsigned int (*_bitcoinconsensus_version)();

int (*_bitcoinconsensus_verify_script_with_amount)(const unsigned char *scriptPubKey, unsigned int scriptPubKeyLen, int64_t amount,
                                    const unsigned char *txTo        , unsigned int txToLen,
                                    unsigned int nIn, unsigned int flags, void* err);

int bitcoinconsensus_verify_script_with_amount(const unsigned char *scriptPubKey, unsigned int scriptPubKeyLen, int64_t amount,
                                    const unsigned char *txTo        , unsigned int txToLen,
                                    unsigned int nIn, unsigned int flags) {
	return _bitcoinconsensus_verify_script_with_amount(scriptPubKey, scriptPubKeyLen, amount, txTo, txToLen, nIn, flags, NULL);
}

unsigned int bitcoinconsensus_version() {
	return _bitcoinconsensus_version();
}

int init_bitcoinconsensus_so() {
	void *so = dlopen("libbitcoinconsensus.so", RTLD_LAZY);
	if (so) {
		*(void **)(&_bitcoinconsensus_version) = dlsym(so, "bitcoinconsensus_version");
		*(void **)(&_bitcoinconsensus_verify_script_with_amount) = dlsym(so, "bitcoinconsensus_verify_script_with_amount");
		if (!_bitcoinconsensus_version) {
			printf("libbitcoinconsensus.so not found\n");
			return 0;
		}
		if (!_bitcoinconsensus_verify_script_with_amount) {
			printf("libbitcoinconsensus.so is too old. Use one of bitcoin-core 0.13.1\n");
			return 0;
		}
		return 1;
	}
	return 0;
}

*/
import "C"

import (
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/script"
)

var (
	ConsensusChecks uint64
	ConsensusExpErr uint64
	ConsensusErrors uint64
	mut             sync.Mutex
)

func check_consensus(pkScr []byte, amount uint64, i int, tx *bch.Tx, ver_flags uint32, result bool) {
	var tmp []byte
	if len(pkScr) != 0 {
		tmp = make([]byte, len(pkScr))
		copy(tmp, pkScr)
	}
	tx_raw := tx.Raw
	if tx_raw == nil {
		tx_raw = tx.Serialize()
	}
	go func(pkScr []byte, txTo []byte, amount uint64, i int, ver_flags uint32, result bool) {
		var pkscr_ptr *C.uchar // default to null
		var pkscr_len C.uint   // default to 0
		if pkScr != nil {
			pkscr_ptr = (*C.uchar)(unsafe.Pointer(&pkScr[0]))
			pkscr_len = C.uint(len(pkScr))
		}
		r1 := int(C.bitcoinconsensus_verify_script_with_amount(pkscr_ptr, pkscr_len, C.int64_t(amount),
			(*C.uchar)(unsafe.Pointer(&txTo[0])), C.uint(len(txTo)), C.uint(i), C.uint(ver_flags)))
		res := r1 == 1
		atomic.AddUint64(&ConsensusChecks, 1)
		if !result {
			atomic.AddUint64(&ConsensusExpErr, 1)
		}
		if res != result {
			atomic.AddUint64(&ConsensusErrors, 1)
			common.CountSafe("TxConsensusERR")
			mut.Lock()
			println("Compare to consensus failed!")
			println("Gocoin:", result, "   ConsLIB:", res)
			println("pkScr", hex.EncodeToString(pkScr))
			println("txTo", hex.EncodeToString(txTo))
			println("amount:", amount, "  input_idx:", i, "  ver_flags:", ver_flags)
			println()
			mut.Unlock()
		}
	}(tmp, tx_raw, amount, i, ver_flags, result)
}

func verify_script_with_amount(pkScr []byte, amount uint64, i int, tx *bch.Tx, ver_flags uint32) (result bool) {
	txTo := tx.Raw
	if txTo == nil {
		txTo = tx.Serialize()
	}
	var pkscr_ptr *C.uchar // default to null
	var pkscr_len C.uint   // default to 0
	if pkScr != nil {
		pkscr_ptr = (*C.uchar)(unsafe.Pointer(&pkScr[0]))
		pkscr_len = C.uint(len(pkScr))
	}
	r1 := int(C.bitcoinconsensus_verify_script_with_amount(pkscr_ptr, pkscr_len, C.int64_t(amount),
		(*C.uchar)(unsafe.Pointer(&txTo[0])), C.uint(len(txTo)), C.uint(i), C.uint(ver_flags)))

	result = (r1 == 1)
	return
}

func consensus_stats(s string) {
	fmt.Println("Consensus Checks:", atomic.LoadUint64(&ConsensusChecks))
	fmt.Println("Consensus ExpErr:", atomic.LoadUint64(&ConsensusExpErr))
	fmt.Println("Consensus Errors:", atomic.LoadUint64(&ConsensusErrors))
}

func init() {
	if C.init_bitcoinconsensus_so() == 0 {
		common.Log.Println("Not using libbitcoinconsensus.so to cross-check consensus rules")
		return
	}
	common.Log.Println("Using libbitcoinconsensus.so version", C.bitcoinconsensus_version(), "to cross-check consensus")
	script.VerifyConsensus = check_consensus
	newUi("cons", false, consensus_stats, "See statistics of the consensus cross-checks")
}
