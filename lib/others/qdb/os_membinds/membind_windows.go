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

// File:        membind_windows.go
// Description: Bictoin Cash Cash qdb Package

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

// +build windows

package qdb

import (
	"reflect"
	"syscall"
	"unsafe"
)

var (
	funcGlobalAlloc *syscall.Proc
	funcGlobalFree  *syscall.Proc
)

func win_HeapAlloc(le uint32) data_ptr_t {
	ptr, _, _ := funcGlobalAlloc.Call(0, uintptr(le))
	return data_ptr_t(ptr)
}

func win_HeapFree(ptr data_ptr_t) {
	funcGlobalFree.Call(uintptr(ptr))
}

func win_AllocPtr(v []byte) data_ptr_t {
	ptr := win_HeapAlloc(uint32(len(v)))
	sl := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(ptr), Len: int(len(v)), Cap: int(len(v))}))
	copy(sl, v)
	return ptr
}

func init() {
	if membind_use_wrapper {
		return
	}
	dll, er := syscall.LoadDLL("kernel32.dll")
	if er != nil {
		return
	}
	funcGlobalAlloc, _ = dll.FindProc("GlobalAlloc")
	funcGlobalFree, _ = dll.FindProc("GlobalFree")
	if funcGlobalAlloc == nil || funcGlobalFree == nil {
		return
	}
	println("Using kernel32.dll for qdb memory bindings")
	_heap_alloc = win_HeapAlloc
	_heap_free = win_HeapFree
	_heap_store = win_AllocPtr
	membind_use_wrapper = true
}
