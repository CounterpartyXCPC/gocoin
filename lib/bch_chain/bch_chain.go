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

// File:		bch_chain.go
// Description:	Bictoin Cash Chain Package

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

package chain

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"

	btc "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
)

var AbortNow bool // set it to true to abort any activity

type Chain struct {
	Blocks  *BlockDB        // blockchain.dat and blockchain.idx
	Unspent *utxo.UnspentDB // unspent folder

	BlockTreeRoot   *BlockTreeNode
	blockTreeEnd    *BlockTreeNode
	blockTreeAccess sync.Mutex
	Genesis         *btc.Uint256

	BlockIndexAccess sync.Mutex
	BlockIndex       map[[btc.Uint256IdxLen]byte]*BlockTreeNode

	CB NewChanOpts // callbacks used by Unspent database

	// UAHF-User activated hard fork	|	aka: ""
	// UASF-User activated soft fork	|	aka: ""
	// The New York Agreement	|	aka: ""
	// SegWit	|	aka: ""
	// BIP 141	|	aka: ""
	// BIP 91 [SegWit2x]	|	aka: "Consensus.BIP91Height"
	// BIP 148	|	aka: ""

	// The official date and time for the fork is:
	// Fork Date: 2017-08-01  12:20 p.m. UTC

	//
	//
	//
	//
	//

	// Block Size Limit Increase
	// Bitcoin Cash started off with an immediate increase of the block size limit to 8MB.

	Consensus struct {
		Window, EnforceUpgrade, RejectBlock uint
		MaxPOWBits                          uint32
		MaxPOWValue                         *big.Int
		GensisTimestamp                     uint32
		Enforce_CSV                         uint32 // if non zero CVS verifications will be enforced from this block onwards
		Enforce_SEGWIT                      uint32 // if non zero CVS verifications will be enforced from this block onwards
		BIP9_Treshold                       uint32 // It is not really used at this moment, but maybe one day...
		BIP34Height                         uint32
		BIP65Height                         uint32
		BIP66Height                         uint32
		BIP91Height                         uint32
		S2XHeight                           uint32
	}
}

type NewChanOpts struct {
	UTXOVolatileMode bool
	UndoBlocks       uint // undo this many blocks when opening the chain
	UTXOCallbacks    utxo.CallbackFunctions
	BlockMinedCB     func(*btc.Block) // used to remove mined txs from memory pool
}

// This is the very first function one should call in order to use this package
func NewChainExt(dbrootdir string, genesis *btc.Uint256, rescan bool, opts *NewChanOpts, bdbopts *BlockDBOpts) (ch *Chain) {
	ch = new(Chain)
	ch.Genesis = genesis

	if opts == nil {
		opts = &NewChanOpts{}
	}

	ch.CB = *opts

	ch.Consensus.GensisTimestamp = 1231006505
	ch.Consensus.MaxPOWBits = 0x1d00ffff
	ch.Consensus.MaxPOWValue, _ = new(big.Int).SetString("00000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)
	if ch.testnet() {
		ch.Consensus.BIP34Height = 21111
		ch.Consensus.BIP65Height = 581885
		ch.Consensus.BIP66Height = 330776
		ch.Consensus.Enforce_CSV = 770112
		ch.Consensus.Enforce_SEGWIT = 834624
		ch.Consensus.BIP9_Treshold = 1512
	} else {
		ch.Consensus.BIP34Height = 227931
		ch.Consensus.BIP65Height = 388381
		ch.Consensus.BIP66Height = 363725
		ch.Consensus.Enforce_CSV = 419328
		ch.Consensus.Enforce_SEGWIT = 481824 // https://www.reddit.com/r/Bitcoin/comments/6okd1n/bip91_lock_in_is_guaranteed_as_of_block_476768/
		ch.Consensus.BIP91Height = 477120
		ch.Consensus.BIP9_Treshold = 1916
	}

	ch.Blocks = NewBlockDBExt(dbrootdir, bdbopts)

	ch.Unspent = utxo.NewUnspentDb(&utxo.NewUnspentOpts{
		Dir: dbrootdir, Rescan: rescan, VolatimeMode: opts.UTXOVolatileMode,
		CB: opts.UTXOCallbacks, AbortNow: &AbortNow})

	if AbortNow {
		return
	}

	ch.loadBlockIndex()
	if AbortNow {
		return
	}

	if rescan {
		ch.SetLast(ch.BlockTreeRoot)
	}

	if AbortNow {
		return
	}

	if opts.UndoBlocks > 0 {
		fmt.Println("Undo", opts.UndoBlocks, "block(s) and exit...")
		for opts.UndoBlocks > 0 {
			ch.UndoLastBlock()
			opts.UndoBlocks--
		}
		return
	}

	// And now re-apply the blocks which you have just reverted :)
	end, _ := ch.BlockTreeRoot.FindFarthestNode()
	if end.Height > ch.LastBlock().Height {
		ch.ParseTillBlock(end)
	} else {
		ch.Unspent.LastBlockHeight = end.Height
	}

	return
}

// Calculate an imaginary header of the genesis block (for Timestamp() and Bits() functions from chain_tree.go)
func (ch *Chain) RebuildGenesisHeader() {
	binary.LittleEndian.PutUint32(ch.BlockTreeRoot.BlockHeader[0:4], 1) // Version
	// [4:36] - prev_block
	// [36:68] - merkle_root
	binary.LittleEndian.PutUint32(ch.BlockTreeRoot.BlockHeader[68:72], ch.Consensus.GensisTimestamp) // Timestamp
	binary.LittleEndian.PutUint32(ch.BlockTreeRoot.BlockHeader[72:76], ch.Consensus.MaxPOWBits)      // Bits
	// [76:80] - nonce
}

// Call this function periodically (i.e. each second)
// when your client is idle, to defragment databases.
func (ch *Chain) Idle() bool {
	ch.Blocks.Idle()
	return ch.Unspent.Idle()
}

// Return blockchain stats in one string.
func (ch *Chain) Stats() (s string) {
	last := ch.LastBlock()
	ch.BlockIndexAccess.Lock()
	s = fmt.Sprintf("CHAIN: blocks:%d  Height:%d  MedianTime:%d\n",
		len(ch.BlockIndex), last.Height, last.GetMedianTimePast())
	ch.BlockIndexAccess.Unlock()
	s += ch.Blocks.GetStats()
	s += ch.Unspent.GetStats()
	return
}

// Close the databases.
func (ch *Chain) Close() {
	ch.Blocks.Close()
	ch.Unspent.Close()
}

// Returns true if we are on Testnet3 chain
func (ch *Chain) testnet() bool {
	return ch.Genesis.Hash[0] == 0x43 // it's simple, but works
}

// For SegWit2X
func (ch *Chain) MaxBlockWeight(height uint32) uint {
	if ch.Consensus.S2XHeight != 0 && height >= ch.Consensus.S2XHeight {
		return 2 * btc.MAX_BLOCK_WEIGHT
	} else {
		return btc.MAX_BLOCK_WEIGHT
	}
}

// For SegWit2X
func (ch *Chain) MaxBlockSigopsCost(height uint32) uint32 {
	if ch.Consensus.S2XHeight != 0 && height >= ch.Consensus.S2XHeight {
		return 2 * btc.MAX_BLOCK_SIGOPS_COST
	} else {
		return btc.MAX_BLOCK_SIGOPS_COST
	}
}

func (ch *Chain) LastBlock() (res *BlockTreeNode) {
	ch.blockTreeAccess.Lock()
	res = ch.blockTreeEnd
	ch.blockTreeAccess.Unlock()
	return
}

func (ch *Chain) SetLast(val *BlockTreeNode) {
	ch.blockTreeAccess.Lock()
	ch.blockTreeEnd = val
	ch.blockTreeAccess.Unlock()
	return
}
