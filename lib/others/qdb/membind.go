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

// File:        membind.go
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

package qdb

import (
	"os"
	"reflect"
	"sync/atomic"
	"unsafe"
)

var (
	membind_use_wrapper bool
	_heap_alloc         func(le uint32) data_ptr_t
	_heap_free          func(ptr data_ptr_t)
	_heap_store         func(v []byte) data_ptr_t
)

type data_ptr_t unsafe.Pointer

func (v *oneIdx) FreeData() {
	if v.data == nil {
		return
	}
	if membind_use_wrapper {
		_heap_free(v.data)
		atomic.AddInt64(&ExtraMemoryConsumed, -int64(v.datlen))
		atomic.AddInt64(&ExtraMemoryAllocCnt, -1)
	}
	v.data = nil
}

func (v *oneIdx) Slice() (res []byte) {
	if membind_use_wrapper {
		res = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(v.data), Len: int(v.datlen), Cap: int(v.datlen)}))
	} else {
		res = *(*[]byte)(v.data)
	}
	return
}

func newIdx(v []byte, f uint32) (r *oneIdx) {
	r = new(oneIdx)
	r.datlen = uint32(len(v))
	r.SetData(v)
	r.flags = f
	return
}

func (r *oneIdx) SetData(v []byte) {
	if membind_use_wrapper {
		r.data = _heap_store(v)
		atomic.AddInt64(&ExtraMemoryConsumed, int64(r.datlen))
		atomic.AddInt64(&ExtraMemoryAllocCnt, 1)
	} else {
		r.data = data_ptr_t(&v)
	}
}

func (v *oneIdx) LoadData(f *os.File) {
	if membind_use_wrapper {
		v.data = _heap_alloc(v.datlen)
		atomic.AddInt64(&ExtraMemoryConsumed, int64(v.datlen))
		atomic.AddInt64(&ExtraMemoryAllocCnt, 1)
		f.Seek(int64(v.datpos), os.SEEK_SET)
		f.Read(*(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(v.data), Len: int(v.datlen), Cap: int(v.datlen)})))
	} else {
		ptr := make([]byte, int(v.datlen))
		v.data = data_ptr_t(&ptr)
		f.Seek(int64(v.datpos), os.SEEK_SET)
		f.Read(ptr)
	}
}
