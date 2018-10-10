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

// File:		membind_windows.go
// Description:	Bictoin Cash utxo Package

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

package utxo

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"syscall"
	"unsafe"
)

func init() {
	MembindInit = func() {
		var (
			hHeap             uintptr
			funcHeapAllocAddr uintptr
			funcHeapFreeAddr  uintptr
		)

		dll, er := syscall.LoadDLL("kernel32.dll")
		if er != nil {
			return
		}
		fun, _ := dll.FindProc("GetProcessHeap")
		hHeap, _, _ = fun.Call()

		fun, _ = dll.FindProc("HeapAlloc")
		funcHeapAllocAddr = fun.Addr()

		fun, _ = dll.FindProc("HeapFree")
		funcHeapFreeAddr = fun.Addr()

		fmt.Println("Using kernel32.dll heap functions for UTXO records")
		malloc = func(le uint32) []byte {
			atomic.AddInt64(&extraMemoryConsumed, int64(le)+24)
			atomic.AddInt64(&extraMemoryAllocCnt, 1)
			ptr, _, _ := syscall.Syscall(funcHeapAllocAddr, 3, hHeap, 0, uintptr(le+24))
			*(*reflect.SliceHeader)(unsafe.Pointer(ptr)) = reflect.SliceHeader{Data: ptr + 24, Len: int(le), Cap: int(le)}
			return *(*[]byte)(unsafe.Pointer(ptr))
		}

		free = func(ptr []byte) {
			atomic.AddInt64(&extraMemoryConsumed, -int64(len(ptr)+24))
			atomic.AddInt64(&extraMemoryAllocCnt, -1)
			syscall.Syscall(funcHeapFreeAddr, 3, hHeap, 0, uintptr(unsafe.Pointer(&ptr[0]))-24)
		}

		malloc_and_copy = func(v []byte) []byte {
			ptr := malloc(uint32(len(v)))
			copy(ptr, v)
			return ptr
		}
	}
}
