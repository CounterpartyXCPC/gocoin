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

// File:        db_disk.go
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

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Qdb is a fast persistent storage database.

The records are binary blobs that can have a variable length, up to 4GB.

The key must be a unique 64-bit value, most likely a hash of the actual key.

They data is stored on a disk, in a folder specified during the call to NewDB().
There are can be three possible files in that folder
 * qdb.0, qdb.1 - these files store a compact version of the entire database
 * qdb.log - this one stores the changes since the most recent qdb.0 or qdb.1

*/
package qdb

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func (db *DB) seq2fn(seq uint32) string {
	return fmt.Sprintf("%s%08x.dat", db.Dir, seq)
}

func (db *DB) checklogfile() {
	// If could not open, create it
	if db.LogFile == nil {
		fn := db.seq2fn(db.DataSeq)
		db.LogFile, _ = os.Create(fn)
		binary.Write(db.LogFile, binary.LittleEndian, uint32(db.DataSeq))
		db.LastValidLogPos = 4
	}
}

// load record from disk, if not loaded yet
func (db *DB) loadrec(idx *oneIdx) {
	if idx.data == nil {
		var f *os.File
		if f, _ = db.DatFiles[idx.DataSeq]; f == nil {
			fn := db.seq2fn(idx.DataSeq)
			f, _ = os.Open(fn)
			if f == nil {
				println("file", fn, "not found")
				os.Exit(1)
			}
			db.DatFiles[idx.DataSeq] = f
		}
		idx.LoadData(f)
	}
}

// add record at the end of the log
func (db *DB) addtolog(f io.Writer, key KeyType, val []byte) (fpos int64) {
	if f == nil {
		db.checklogfile()
		db.LogFile.Seek(db.LastValidLogPos, os.SEEK_SET)
		f = db.LogFile
	}

	fpos = db.LastValidLogPos
	f.Write(val)
	db.LastValidLogPos += int64(len(val)) // 4 bytes for CRC

	return
}

// add record at the end of the log
func (db *DB) cleanupold(used map[uint32]bool) {
	filepath.Walk(db.Dir, func(path string, info os.FileInfo, err error) error {
		fn := info.Name()
		if len(fn) == 12 && fn[8:12] == ".dat" {
			v, er := strconv.ParseUint(fn[:8], 16, 32)
			if er == nil && uint32(v) != db.DataSeq {
				if _, ok := used[uint32(v)]; !ok {
					//println("deleting", v, path)
					if f, _ := db.DatFiles[uint32(v)]; f != nil {
						f.Close()
						delete(db.DatFiles, uint32(v))
					}
					os.Remove(path)
				}
			}
		}
		return nil
	})
}
