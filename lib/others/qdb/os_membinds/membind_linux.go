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

// File:        membind_linux.go
// Description: Bictoin Cash Cash qdb Package

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

// +build linux

/*
If this file does not build and you don't know what to do, simply delete it and rebuild.
*/

package qdb

/*
#include <stdlib.h>
#include <string.h>

static void *alloc_ptr(void *c, unsigned long l) {
	void *ptr = malloc(l);
	memcpy(ptr, c, l);
	return ptr;
}

static void *my_alloc(unsigned long l) {
	return malloc(l);
}

*/
import "C"

import (
	"unsafe"
)

func gcc_HeapAlloc(le uint32) data_ptr_t {
	return data_ptr_t(C.my_alloc(C.ulong(le)))
}

func gcc_HeapFree(ptr data_ptr_t) {
	C.free(unsafe.Pointer(ptr))
}

func gcc_AllocPtr(v []byte) data_ptr_t {
	ptr := unsafe.Pointer(&v[0]) // see https://github.com/golang/go/issues/15172
	return data_ptr_t(C.alloc_ptr(ptr, C.ulong(len(v))))
}

func init() {
	if membind_use_wrapper {
		panic("Another wrapper already initialized")
	}
	println("Using malloc() qdb memory bindings")
	_heap_alloc = gcc_HeapAlloc
	_heap_free = gcc_HeapFree
	_heap_store = gcc_AllocPtr
	membind_use_wrapper = true
}
