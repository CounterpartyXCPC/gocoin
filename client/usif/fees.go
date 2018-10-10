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

// File:		fees.go
// Description:	Bictoin Cash usif Package

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

package usif

import (
	"bufio"
	"encoding/gob"
	"os"
	"sync"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

const (
	BLKFES_FILE_NAME = "blkfees.gob"
)

var (
	BchBlockFeesMutex sync.Mutex
	BchBlockFees      map[uint32][][3]uint64 = make(map[uint32][][3]uint64) // [0]=Weight  [1]-Fee  [2]-Group
	BchBlockFeesDirty bool                                                  // it true, clean up old data
)

func ProcessBlockFees(height uint32, bl *bch.BchBlock) {
	if len(bl.Txs) < 2 {
		return
	}

	txs := make(map[[32]byte]int, len(bl.Txs)) // group_id -> transaciton_idx
	txs[bl.Txs[0].Hash.Hash] = 0

	fees := make([][3]uint64, len(bl.Txs)-1)

	for i := 1; i < len(bl.Txs); i++ {
		txs[bl.Txs[i].Hash.Hash] = i
		fees[i-1][0] = uint64(3*bl.Txs[i].NoWitSize + bl.Txs[i].Size)
		fees[i-1][1] = uint64(bl.Txs[i].Fee)
		fees[i-1][2] = uint64(i)
	}

	for i := 1; i < len(bl.Txs); i++ {
		for _, inp := range bl.Txs[i].TxIn {
			if paretidx, yes := txs[inp.Input.Hash]; yes {
				if fees[paretidx-1][2] < fees[i-1][2] { // only update it for a lower index
					fees[i-1][2] = fees[paretidx-1][2]
				}
			}
		}
	}

	BchBlockFeesMutex.Lock()
	BchBlockFees[height] = fees
	BchBlockFeesDirty = true
	BchBlockFeesMutex.Unlock()
}

func ExpireBlockFees() {
	var height uint32
	common.Last.Lock()
	height = common.Last.BchBlock.Height
	common.Last.Unlock()

	if height <= 144 {
		return
	}
	height -= 144

	BchBlockFeesMutex.Lock()
	if BchBlockFeesDirty {
		for k := range BchBlockFees {
			if k < height {
				delete(BchBlockFees, k)
			}
		}
		BchBlockFeesDirty = false
	}
	BchBlockFeesMutex.Unlock()
}

func SaveBlockFees() {
	f, er := os.Create(common.GocoinCashHomeDir + BLKFES_FILE_NAME)
	if er != nil {
		println("SaveBlockFees:", er.Error())
		return
	}

	ExpireBlockFees()
	buf := bufio.NewWriter(f)
	er = gob.NewEncoder(buf).Encode(BchBlockFees)

	if er != nil {
		println("SaveBlockFees:", er.Error())
	}

	buf.Flush()
	f.Close()

}

func LoadBlockFees() {
	f, er := os.Open(common.GocoinCashHomeDir + BLKFES_FILE_NAME)
	if er != nil {
		println("LoadBlockFees:", er.Error())
		return
	}

	buf := bufio.NewReader(f)
	er = gob.NewDecoder(buf).Decode(&BchBlockFees)
	if er != nil {
		println("LoadBlockFees:", er.Error())
	}

	f.Close()
}
