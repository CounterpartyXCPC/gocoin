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

// File:		membind_linux.go
// Description:	Bictoin Cash utxo Package

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

/*
If this file does not build and you don't know what to do, simply delete it and rebuild.
*/

package utxo

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"unsafe"
)

func init() {
	MembindInit = func() {
		fmt.Println("Using malloc() and free() for UTXO records")

		malloc = func(le uint32) []byte {
			atomic.AddInt64(&extraMemoryConsumed, int64(le)+24)
			atomic.AddInt64(&extraMemoryAllocCnt, 1)
			ptr := uintptr(C.malloc(C.size_t(le + 24)))
			*(*reflect.SliceHeader)(unsafe.Pointer(ptr)) = reflect.SliceHeader{Data: ptr + 24, Len: int(le), Cap: int(le)}
			return *(*[]byte)(unsafe.Pointer(ptr))
		}

		free = func(ptr []byte) {
			atomic.AddInt64(&extraMemoryConsumed, -int64(len(ptr)+24))
			atomic.AddInt64(&extraMemoryAllocCnt, -1)
			C.free(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr[0])) - 24))
		}

		malloc_and_copy = func(v []byte) []byte {
			sl := malloc(uint32(len(v)))
			copy(sl, v)
			return sl
		}
	}
}
