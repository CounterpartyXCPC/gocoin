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

// File:		vars.go
// Description:	Bictoin Cash network Package

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

package network

import (
	"sync"
	"time"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

type OneReceivedBlock struct {
	TmStart        time.Time // when we receioved message letting us about this block
	TmPreproc      time.Time // when we added this block to BlocksToGet
	TmDownload     time.Time // when we finished dowloading of this block
	TmQueue        time.Time // when we started comitting this block
	TmAccepted     time.Time // when the block was commited to blockchain
	Cnt            uint
	TxMissing      int
	FromConID      uint32
	NonWitnessSize int
	DoInvs         bool
}

type BchBlockRcvd struct {
	Conn *OneConnection
	*bch.BchBlock
	*bch_chain.BchBlockTreeNode
	*OneReceivedBlock
	*bch.BchBlockExtraInfo
}

type TxRcvd struct {
	conn *OneConnection
	*bch.Tx
	trusted, local bool
}

type OneBlockToGet struct {
	Started time.Time
	*bch.BchBlock
	*bch_chain.BchBlockTreeNode
	InProgress uint
	TmPreproc  time.Time // how long it took to start downloading this block
	SendInvs   bool
}

var (
	ReceivedBlocks           map[BIDX]*OneReceivedBlock = make(map[BIDX]*OneReceivedBlock, 400e3)
	BchBlocksToGet           map[BIDX]*OneBlockToGet    = make(map[BIDX]*OneBlockToGet)
	IndexToBlocksToGet       map[uint32][]BIDX          = make(map[uint32][]BIDX)
	LowestIndexToBlocksToGet uint32
	LastCommitedHeader       *bch_chain.BchBlockTreeNode
	MutexRcv                 sync.Mutex

	NetBlocks chan *BchBlockRcvd = make(chan *BchBlockRcvd, MAX_BLOCKS_FORWARD_CNT+10)
	NetTxs    chan *TxRcvd       = make(chan *TxRcvd, 2000)

	CachedBlocks    []*BchBlockRcvd
	CachedBlocksLen sys.SyncInt

	DiscardedBlocks map[BIDX]bool = make(map[BIDX]bool)

	HeadersReceived sys.SyncInt
)

func AddB2G(b2g *OneBlockToGet) {
	bidx := b2g.BchBlock.Hash.BIdx()
	BchBlocksToGet[bidx] = b2g
	bh := b2g.BchBlockTreeNode.Height
	IndexToBlocksToGet[bh] = append(IndexToBlocksToGet[bh], bidx)
	if LowestIndexToBlocksToGet == 0 || bh < LowestIndexToBlocksToGet {
		LowestIndexToBlocksToGet = bh
	}

	/* TODO: this was causing deadlock. Removing it for now as maybe it is not even needed.
	// Trigger each connection to as the peer for block data
	Mutex_net.Lock()
	for _, v := range OpenCons {
		v.MutexSetBool(&v.X.GetBlocksDataNow, true)
	}
	Mutex_net.Unlock()
	*/
}

func DelB2G(idx BIDX) {
	b2g := BchBlocksToGet[idx]
	if b2g == nil {
		println("DelB2G - not found")
		return
	}

	bh := b2g.BchBlockTreeNode.Height
	iii := IndexToBlocksToGet[bh]
	if len(iii) > 1 {
		var n []BIDX
		for _, cidx := range iii {
			if cidx != idx {
				n = append(n, cidx)
			}
		}
		if len(n)+1 != len(iii) {
			println("DelB2G - index not found")
		}
		IndexToBlocksToGet[bh] = n
	} else {
		if iii[0] != idx {
			println("DelB2G - index not matching")
		}
		delete(IndexToBlocksToGet, bh)
		if bh == LowestIndexToBlocksToGet {
			if len(IndexToBlocksToGet) > 0 {
				for LowestIndexToBlocksToGet++; ; LowestIndexToBlocksToGet++ {
					if _, ok := IndexToBlocksToGet[LowestIndexToBlocksToGet]; ok {
						break
					}
				}
			} else {
				LowestIndexToBlocksToGet = 0
			}
		}
	}

	delete(BchBlocksToGet, idx)
}
