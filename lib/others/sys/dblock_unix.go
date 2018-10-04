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

// File:        dblock_unix.go
// Description: Bictoin Cash Cash sys Package

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

// +build !windows

package sys

import (
	"os"
	"syscall"
)

var (
	DbLockFileName string
	DbLockFileHndl *os.File
)

func LockDatabaseDir(GocoinCashHomeDir string) {
	os.MkdirAll(GocoinCashHomeDir, 0770)
	DbLockFileName = GocoinCashHomeDir + ".lock"
	DbLockFileHndl, _ = os.Open(DbLockFileName)
	if DbLockFileHndl == nil {
		DbLockFileHndl, _ = os.Create(DbLockFileName)
	}
	if DbLockFileHndl == nil {
		goto error
	}

	if e := syscall.Flock(int(DbLockFileHndl.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		goto error
	}
	return

error:
	println("Could not lock the databse folder for writing. Another instance might be running.")
	println("If it is not the case, remove this file:", DbLockFileName)
	os.Exit(1)
}

func UnlockDatabaseDir() {
	syscall.Flock(int(DbLockFileHndl.Fd()), syscall.LOCK_UN)
	DbLockFileHndl.Close()
	os.Remove(DbLockFileName)
}
