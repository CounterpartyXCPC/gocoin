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

// File:        db_test.go
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
	"bytes"
	cr "crypto/rand"
	"encoding/hex"
	"fmt"
	mr "math/rand"
	"os"
	"testing"
	"time"
)

const (
	dbname   = "test"
	oneRound = 10000
	delRound = 1000
)

func getRecSize() int {
	return 4
	//return mr.Intn(4096)
}

func kim(v []byte) bool {
	return (mr.Int63() & 1) == 0
}

func dumpidx(db *DB) {
	println("index")
	for k, v := range db.Idx.Index {
		println(k2s(k), v.datpos, v.datlen)
	}
}

func TestDatabase(t *testing.T) {
	var key KeyType
	var val, v []byte
	var db *DB
	var e error

	os.RemoveAll(dbname)
	mr.Seed(time.Now().UnixNano())

	db, e = NewDB(dbname, true)
	if e != nil {
		t.Error("Cannot create db")
		return
	}

	// Add oneRound random records
	for i := 0; i < oneRound; i++ {
		vlen := getRecSize()
		val = make([]byte, vlen)
		key = KeyType(mr.Int63())
		cr.Read(val[:])
		db.Put(key, val)
	}
	db.Close()

	// Reopen DB, verify, defrag and close
	db, e = NewDB(dbname, true)
	if e != nil {
		t.Error("Cannot reopen db")
		return
	}
	if db.Count() != oneRound {
		t.Error("Bad count", db.Count(), oneRound)
		return
	}
	//dumpidx(db)
	v = db.Get(key)
	if !bytes.Equal(val, v) {
		t.Error("Key data mismatch ", k2s(key), "/", hex.EncodeToString(val), "/", hex.EncodeToString(v))
		return
	}
	if db.Count() != oneRound {
		t.Error("Wrong number of records", db.Count(), oneRound)
	}
	db.Defrag(false)
	db.Close()

	// Reopen DB, verify, add oneRound more records and Close
	db, e = NewDB(dbname, true)
	if e != nil {
		t.Error("Cannot reopen db")
		return
	}
	v = db.Get(key)
	if !bytes.Equal(val[:], v[:]) {
		t.Error("Key data mismatch")
	}
	if db.Count() != oneRound {
		t.Error("Wrong number of records", db.Count())
	}
	db.NoSync()
	for i := 0; i < oneRound; i++ {
		vlen := getRecSize()
		val = make([]byte, vlen)
		key = KeyType(mr.Int63())
		cr.Read(val[:])
		db.Put(key, val)
	}
	db.Sync()
	db.Close()

	// Reopen DB, verify, defrag and close
	db, e = NewDB(dbname, true)
	if e != nil {
		t.Error("Cannot reopen db")
		return
	}
	v = db.Get(key)
	if !bytes.Equal(val[:], v[:]) {
		t.Error("Key data mismatch")
	}
	if db.Count() != 2*oneRound {
		t.Error("Wrong number of records", db.Count())
		return
	}
	db.Defrag(true)
	db.Close()

	// Reopen DB, verify, close...
	db, e = NewDB(dbname, true)
	if e != nil {
		t.Error("Cannot reopen db")
		return
	}
	v = db.Get(key)
	if !bytes.Equal(val[:], v[:]) {
		t.Error("Key data mismatch")
	}
	if db.Count() != 2*oneRound {
		t.Error("Wrong number of records", db.Count())
	}
	db.Close()

	// Reopen, delete 100 records, close...
	db, e = NewDB(dbname, true)
	if e != nil {
		t.Error("Cannot reopen db")
		return
	}

	var keys []KeyType
	db.Browse(func(key KeyType, v []byte) uint32 {
		keys = append(keys, key)
		if len(keys) < delRound {
			return 0
		} else {
			return BR_ABORT
		}
	})
	for i := range keys {
		db.Del(keys[i])
	}
	db.Close()

	// Reopen DB, verify, close...
	db, e = NewDB(dbname, true)
	if db.Count() != 2*oneRound-delRound {
		t.Error("Wrong number of records", db.Count())
	}
	db.Close()

	// Reopen DB, verify, close...
	db, e = NewDB(dbname, true)
	db.Defrag(false)
	if db.Count() != 2*oneRound-delRound {
		t.Error("Wrong number of records", db.Count())
	}
	db.Close()

	// Reopen DB, verify, close...
	db, e = NewDB(dbname, true)
	if db.Count() != 2*oneRound-delRound {
		t.Error("Wrong number of records", db.Count())
	}
	db.Close()

	os.RemoveAll(dbname)
}

func k2s(k KeyType) string {
	return fmt.Sprintf("%16x", k)
}
