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

// File:        index.go
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

package qdb

import (
	"io/ioutil"
	"os"
)

type QdbIndex struct {
	db                 *DB
	IdxFilePath        string
	file               *os.File
	DatfileIndex       int
	VersionSequence    uint32
	MaxDatfileSequence uint32

	Index map[KeyType]*oneIdx

	DiskSpaceNeeded uint64
	ExtraSpaceUsed  uint64
}

func NewDBidx(db *DB, recs uint) (idx *QdbIndex) {
	idx = new(QdbIndex)
	idx.db = db
	idx.IdxFilePath = db.Dir + "qdbidx."
	if recs == 0 {
		idx.Index = make(map[KeyType]*oneIdx)
	} else {
		idx.Index = make(map[KeyType]*oneIdx, recs)
	}
	used := make(map[uint32]bool, 10)
	idx.loaddat(used)
	idx.loadlog(used)
	idx.db.cleanupold(used)
	return
}

func (idx *QdbIndex) load(walk QdbWalkFunction) {
	dats := make(map[uint32][]byte)
	idx.browse(func(k KeyType, v *oneIdx) bool {
		if walk != nil || (v.flags&NO_CACHE) == 0 {
			dat := dats[v.DataSeq]
			if dat == nil {
				dat, _ = ioutil.ReadFile(idx.db.seq2fn(v.DataSeq))
				if dat == nil {
					println("Database corrupt - missing file:", idx.db.seq2fn(v.DataSeq))
					os.Exit(1)
				}
				dats[v.DataSeq] = dat
			}
			v.SetData(dat[v.datpos : v.datpos+v.datlen])
			if walk != nil {
				res := walk(k, v.Slice())
				v.aply_browsing_flags(res)
				v.freerec()
			}
		}
		return true
	})
}

func (idx *QdbIndex) size() int {
	return len(idx.Index)
}

func (idx *QdbIndex) get(k KeyType) *oneIdx {
	return idx.Index[k]
}

func (idx *QdbIndex) memput(k KeyType, rec *oneIdx) {
	if prv, ok := idx.Index[k]; ok {
		prv.FreeData()
		dif := uint64(24 + prv.datlen)
		if !idx.db.VolatileMode {
			idx.ExtraSpaceUsed += dif
			idx.DiskSpaceNeeded -= dif
		}
	}
	idx.Index[k] = rec

	if !idx.db.VolatileMode {
		idx.DiskSpaceNeeded += uint64(24 + rec.datlen)
	}
	if rec.DataSeq > idx.MaxDatfileSequence {
		idx.MaxDatfileSequence = rec.DataSeq
	}
}

func (idx *QdbIndex) memdel(k KeyType) {
	if cur, ok := idx.Index[k]; ok {
		cur.FreeData()
		dif := uint64(12 + cur.datlen)
		if !idx.db.VolatileMode {
			idx.ExtraSpaceUsed += dif
			idx.DiskSpaceNeeded -= dif
		}
		delete(idx.Index, k)
	}
}

func (idx *QdbIndex) put(k KeyType, rec *oneIdx) {
	idx.memput(k, rec)
	if idx.db.VolatileMode {
		return
	}
	idx.addtolog(nil, k, rec)
}

func (idx *QdbIndex) del(k KeyType) {
	idx.memdel(k)
	if idx.db.VolatileMode {
		return
	}
	idx.deltolog(nil, k)
}

func (idx *QdbIndex) browse(walk func(key KeyType, idx *oneIdx) bool) {
	for k, v := range idx.Index {
		if !walk(k, v) {
			break
		}
	}
}

func (idx *QdbIndex) close() {
	if idx.file != nil {
		idx.file.Close()
		idx.file = nil
	}
	idx.Index = nil
}
