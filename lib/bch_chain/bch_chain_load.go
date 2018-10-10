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

// File:		bch_chain_load.go
// Description:	Bictoin Cash bch_chain Package

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

package bch_chain

import (
	"errors"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

func nextBlock(ch *Chain, hash, header []byte, height, blen, txs uint32) {
	bh := bch.NewUint256(hash[:])
	if _, ok := ch.BchBlockIndex[bh.BIdx()]; ok {
		println("nextBlock:", bh.String(), "- already in")
		return
	}
	v := new(BchBlockTreeNode)
	v.BchBlockHash = bh
	v.Height = height
	v.BchBlockSize = blen
	v.TxCount = txs
	copy(v.BchBlockHeader[:], header)
	ch.BchBlockIndex[v.BchBlockHash.BIdx()] = v
}

// Loads block index from the disk
func (ch *Chain) loadBlockIndex() {
	ch.BchBlockIndex = make(map[[bch.Uint256IdxLen]byte]*BchBlockTreeNode, BlockMapInitLen)
	ch.BchBlockTreeRoot = new(BchBlockTreeNode)
	ch.BchBlockTreeRoot.BchBlockHash = ch.Genesis
	ch.RebuildGenesisHeader()
	ch.BchBlockIndex[ch.Genesis.BIdx()] = ch.BchBlockTreeRoot

	ch.BchBlocks.LoadBlockIndex(ch, nextBlock)
	tlb := ch.Unspent.LastBlockHash
	//println("Building tree from", len(ch.BchBlockIndex), "nodes")
	for k, v := range ch.BchBlockIndex {
		if AbortNow {
			return
		}
		if v == ch.BchBlockTreeRoot {
			// skip root block (should be only one)
			continue
		}

		par, ok := ch.BchBlockIndex[bch.NewUint256(v.BchBlockHeader[4:36]).BIdx()]
		if !ok {
			println("ERROR: Block", v.Height, v.BchBlockHash.String(), "has no Parent")
			println("...", bch.NewUint256(v.BchBlockHeader[4:36]).String(), "- removing it from blocksDB")
			delete(ch.BchBlockIndex, k)
			continue
		}
		v.Parent = par
		v.Parent.addChild(v)
	}
	if tlb == nil {
		//println("No last block - full rescan will be needed")
		ch.SetLast(ch.BchBlockTreeRoot)
		return
	} else {
		//println("Last Block Hash:", bch.NewUint256(tlb).String())
		last, ok := ch.BchBlockIndex[bch.NewUint256(tlb).BIdx()]
		if !ok {
			panic("Last Block Hash not found")
		}
		ch.SetLast(last)
	}
}

func (ch *Chain) GetRawTx(BchBlockHeight uint32, txid *bch.Uint256) (data []byte, er error) {
	// Find the block with the indicated Height in the main tree
	ch.BchBlockIndexAccess.Lock()
	n := ch.LastBlock()
	if n.Height < BchBlockHeight {
		println(n.Height, BchBlockHeight)
		ch.BchBlockIndexAccess.Unlock()
		er = errors.New("GetRawTx: block height too big")
		return
	}
	for n.Height > BchBlockHeight {
		n = n.Parent
	}
	ch.BchBlockIndexAccess.Unlock()

	bd, _, e := ch.BchBlocks.BchBlockGet(n.BchBlockHash)
	if e != nil {
		er = errors.New("GetRawTx: block not in the database")
		return
	}

	bl, e := bch.NewBchBlock(bd)
	if e != nil {
		er = errors.New("GetRawTx: NewBlock failed")
		return
	}

	e = bl.BuildTxList()
	if e != nil {
		er = errors.New("GetRawTx: BuildTxList failed")
		return
	}

	// Find the transaction we need and store it in the file
	for i := range bl.Txs {
		if bl.Txs[i].Hash.Equal(txid) {
			data = bl.Txs[i].Serialize()
			return
		}
	}
	er = errors.New("GetRawTx: BuildTxList failed")
	return
}
