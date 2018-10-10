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

// File:		bch_chain_tree.go
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
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

type BchBlockTreeNode struct {
	BchBlockHash *bch.Uint256
	Height       uint32
	Parent       *BchBlockTreeNode
	Childs       []*BchBlockTreeNode

	BchBlockSize uint32 // if this is zero, only header is known so far
	TxCount      uint32
	SigopsCost   uint32

	BchBlockHeader [80]byte

	Trusted bool
}

func (ch *Chain) ParseTillBlock(end *BchBlockTreeNode) {
	var crec *BlckCachRec
	var er error
	var trusted bool
	var tot_bytes uint64

	last := ch.LastBlock()
	var total_size_to_process uint64
	fmt.Print("Calculating size of blockchain overhead...")
	for n := end; n != nil && n != last; n = n.Parent {
		l, _ := ch.BchBlocks.BchBlockLength(n.BchBlockHash, false)
		total_size_to_process += uint64(l)
	}
	fmt.Println("\rApplying", total_size_to_process>>20, "MB of transactions data from", end.Height-last.Height, "blocks to UTXO.db")
	sta := time.Now()
	prv := sta
	for !AbortNow && last != end {
		cur := time.Now()
		if cur.Sub(prv) >= 10*time.Second {
			mbps := float64(tot_bytes) / float64(cur.Sub(sta)/1e3)
			sec_left := int64(float64(total_size_to_process) / 1e6 / mbps)
			fmt.Printf("ParseTillBlock %d / %d ... %.2f MB/s - %d:%02d:%02d left\n", last.Height,
				end.Height, mbps, sec_left/3600, (sec_left/60)%60, sec_left%60)
			prv = cur
		}

		nxt := last.FindPathTo(end)
		if nxt == nil {
			break
		}

		if nxt.BchBlockSize == 0 {
			println("ParseTillBlock: ", nxt.Height, nxt.BchBlockHash.String(), "- not yet commited")
			break
		}

		crec, trusted, er = ch.BchBlocks.BchBlockGetInternal(nxt.BchBlockHash, true)
		if er != nil {
			panic("Db.BchBlockGet(): " + er.Error())
		}
		tot_bytes += uint64(len(crec.Data))
		l, _ := ch.BchBlocks.BchBlockLength(nxt.BchBlockHash, false)
		total_size_to_process -= uint64(l)

		bl, er := bch.NewBchBlock(crec.Data)
		if er != nil {
			ch.DeleteBranch(nxt, nil)
			break
		}
		bl.Height = nxt.Height

		// Recover the flags to be used when verifying scripts for non-trusted blocks (stored orphaned blocks)
		ch.ApplyBlockFlags(bl)

		// Do not recover MedianPastTime as it is only checked in PostCheckBlock()
		// ... that had to be done before the block was stored on disk.

		er = bl.BuildTxList()
		if er != nil {
			ch.DeleteBranch(nxt, nil)
			break
		}

		bl.Trusted = trusted

		changes, sigopscost, er := ch.ProcessBlockTransactions(bl, nxt.Height, end.Height)
		if er != nil {
			println("ProcessBlockTransactionsB", nxt.BchBlockHash.String(), nxt.Height, er.Error())
			ch.DeleteBranch(nxt, nil)
			break
		}
		nxt.SigopsCost = sigopscost
		if !trusted {
			ch.BchBlocks.BchBlockTrusted(bl.Hash.Hash[:])
		}

		ch.Unspent.CommitBlockTxs(changes, bl.Hash.Hash[:])

		ch.SetLast(nxt)
		last = nxt

		if ch.CB.BchBlockMinedCB != nil {
			bl.Height = nxt.Height
			bl.LastKnownHeight = end.Height
			ch.CB.BchBlockMinedCB(bl)
		}
	}

	if !AbortNow && last != end {
		end, _ = ch.BchBlockTreeRoot.FindFarthestNode()
		fmt.Println("ParseTillBlock failed - now go to", end.Height)
		ch.MoveToBlock(end)
	}
}

func (n *BchBlockTreeNode) BchBlockVersion() uint32 {
	return binary.LittleEndian.Uint32(n.BchBlockHeader[0:4])
}

func (n *BchBlockTreeNode) Timestamp() uint32 {
	return binary.LittleEndian.Uint32(n.BchBlockHeader[68:72])
}

func (n *BchBlockTreeNode) Bits() uint32 {
	return binary.LittleEndian.Uint32(n.BchBlockHeader[72:76])
}

// Returns median time of the last 11 blocks
func (pindex *BchBlockTreeNode) GetMedianTimePast() uint32 {
	var pmedian [MedianTimeSpan]int
	pbegin := MedianTimeSpan
	pend := MedianTimeSpan
	for i := 0; i < MedianTimeSpan && pindex != nil; i++ {
		pbegin--
		pmedian[pbegin] = int(pindex.Timestamp())
		pindex = pindex.Parent
	}
	sort.Ints(pmedian[pbegin:pend])
	return uint32(pmedian[pbegin+((pend-pbegin)/2)])
}

// Looks for the fartherst node
func (n *BchBlockTreeNode) FindFarthestNode() (*BchBlockTreeNode, int) {
	//fmt.Println("FFN:", n.Height, "kids:", len(n.Childs))
	if len(n.Childs) == 0 {
		return n, 0
	}
	res, depth := n.Childs[0].FindFarthestNode()
	if len(n.Childs) > 1 {
		for i := 1; i < len(n.Childs); i++ {
			_re, _dept := n.Childs[i].FindFarthestNode()
			if _dept > depth {
				res = _re
				depth = _dept
			}
		}
	}
	return res, depth + 1
}

// Returns the next node that leads to the given destiantion
func (n *BchBlockTreeNode) FindPathTo(end *BchBlockTreeNode) *BchBlockTreeNode {
	if n == end {
		return nil
	}

	if end.Height <= n.Height {
		panic("FindPathTo: End block is not higher then current")
	}

	if len(n.Childs) == 0 {
		panic("FindPathTo: Unknown path to block " + end.BchBlockHash.String())
	}

	if len(n.Childs) == 1 {
		return n.Childs[0] // if there is only one child, do it fast
	}

	for {
		// more then one children: go from the end until you reach the current node
		if end.Parent == n {
			return end
		}
		end = end.Parent
	}

	return nil
}

// Check whether the given node has all its parent blocks already comitted
func (ch *Chain) HasAllParents(dst *BchBlockTreeNode) bool {
	for {
		dst = dst.Parent
		if ch.OnActiveBranch(dst) {
			return true
		}
		if dst == nil || dst.TxCount == 0 {
			return false
		}
	}
}

// returns true if the given node is on the active branch
func (ch *Chain) OnActiveBranch(dst *BchBlockTreeNode) bool {
	top := ch.LastBlock()
	for {
		if dst == top {
			return true
		}
		if dst.Height >= top.Height {
			return false
		}
		top = top.Parent
	}
}

// Performs channel reorg
func (ch *Chain) MoveToBlock(dst *BchBlockTreeNode) {
	cur := dst
	for cur.Height > ch.LastBlock().Height {
		cur = cur.Parent

		// if cur.TxCount is zero, it means we dont yet have this block's data
		if cur.TxCount == 0 {
			fmt.Println("MoveToBlock cannot continue A")
			fmt.Println("Trying to go:", dst.BchBlockHash.String())
			fmt.Println("Cannot go at:", cur.BchBlockHash.String())
			return
		}
	}

	// At this point both "ch.blockTreeEnd" and "cur" should be at the same height
	for tmp := ch.LastBlock(); tmp != cur; tmp = tmp.Parent {
		if cur.Parent.TxCount == 0 {
			fmt.Println("MoveToBlock cannot continue B")
			fmt.Println("Trying to go:", dst.BchBlockHash.String())
			fmt.Println("Cannot go at:", cur.Parent.BchBlockHash.String())
			return
		}
		cur = cur.Parent
	}

	// At this point "cur" is at the highest common block
	for ch.LastBlock() != cur {
		if AbortNow {
			return
		}
		ch.UndoLastBlock()
	}
	ch.ParseTillBlock(dst)
}

func (ch *Chain) UndoLastBlock() {
	last := ch.LastBlock()
	fmt.Println("Undo block", last.Height, last.BchBlockHash.String(), last.BchBlockSize>>10, "KB")

	crec, _, er := ch.BchBlocks.BchBlockGetInternal(last.BchBlockHash, true)
	if er != nil {
		panic(er.Error())
	}

	bl, _ := bch.NewBchBlock(crec.Data)
	bl.BuildTxList()

	ch.Unspent.UndoBlockTxs(bl, last.Parent.BchBlockHash.Hash[:])
	ch.SetLast(last.Parent)
}

// make sure ch.BchBlockIndexAccess is locked before calling it
func (cur *BchBlockTreeNode) delAllChildren(ch *Chain, deleteCallback func(*bch.Uint256)) {
	for i := range cur.Childs {
		if deleteCallback != nil {
			deleteCallback(cur.Childs[i].BchBlockHash)
		}
		cur.Childs[i].delAllChildren(ch, deleteCallback)
		delete(ch.BchBlockIndex, cur.Childs[i].BchBlockHash.BIdx())
		ch.BchBlocks.BchBlockInvalid(cur.BchBlockHash.Hash[:])
	}
	cur.Childs = nil
}

func (ch *Chain) DeleteBranch(cur *BchBlockTreeNode, deleteCallback func(*bch.Uint256)) {
	// first disconnect it from the Parent
	ch.BchBlocks.BchBlockInvalid(cur.BchBlockHash.Hash[:])
	ch.BchBlockIndexAccess.Lock()
	delete(ch.BchBlockIndex, cur.BchBlockHash.BIdx())
	cur.Parent.delChild(cur)
	cur.delAllChildren(ch, deleteCallback)
	ch.BchBlockIndexAccess.Unlock()
}

func (n *BchBlockTreeNode) addChild(c *BchBlockTreeNode) {
	n.Childs = append(n.Childs, c)
}

func (n *BchBlockTreeNode) delChild(c *BchBlockTreeNode) {
	newChds := make([]*BchBlockTreeNode, len(n.Childs)-1)
	xxx := 0
	for i := range n.Childs {
		if n.Childs[i] != c {
			newChds[xxx] = n.Childs[i]
			xxx++
		}
	}
	if xxx != len(n.Childs)-1 {
		panic("Child not found")
	}
	n.Childs = newChds
}
