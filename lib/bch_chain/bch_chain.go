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
// Description:	Bictoin Cash bch_chain Package

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

package bch_chain

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
)

var AbortNow bool // set it to true to abort any activity

type Chain struct {
	BchBlocks *BchBlockDB     // blockchain.dat and blockchain.idx
	Unspent   *utxo.UnspentDB // unspent folder

	BchBlockTreeRoot *BchBlockTreeNode
	blockTreeEnd     *BchBlockTreeNode
	blockTreeAccess  sync.Mutex
	Genesis          *bch.Uint256

	BchBlockIndexAccess sync.Mutex
	BchBlockIndex       map[[bch.Uint256IdxLen]byte]*BchBlockTreeNode

	CB NewChanOpts // callbacks used by Unspent database

	// UAHF-User activated hard fork [Bitcoin Cash]	|	aka: ""
	// UASF-User activated soft fork [BIP 148]	|	aka: ""
	// The New York Agreement []	|	aka: ""
	// SegWit [BIP 141]		|	aka: ""
	// BIP 141 [Segregated Witness]		|	aka: ""
	// BIP 91 [SegWit2x]	|	aka: "Consensus.BIP91Height"
	// BIP 148 [UASF]		|	aka: ""

	// The official date and time for the fork is:
	// Fork Date: 2017-08-01  12:20 p.m. UTC

	// Block Size Limit Increase
	// Bitcoin Cash started off with an immediate increase of the block size limit to 8MB.

	// Replay and Wipeout Protection
	// When BCH split, it was applying what was described as a well thought out replay and wipeout protection plan
	// for both chains. With this, everyone involved will have minimum disruptions and both the chains can peacefully
	// coexist from there. Until the coming hash wars. ;)

	// New Transaction Type
	// As part of the replay protection technology, Bitcoin Cash has introduced a new transaction type with additional
	// benefits such as input value signing for improved hardware wallet security, and elimination of the quadratic hashing
	// problem. (Source-https://www.bitcoincash.org/)

	// Q. How is transaction replay being handled between the new and the old blockchain?
	// A. Bitcoin Cash transactions use a new flag SIGHASH_FORKID, which is non standard to the legacy blockchain.
	// This prevents Bitcoin Cash transactions from being replayed on the Bitcoin blockchain and vice versa.

	Consensus struct {
		Window, EnforceUpgrade, RejectBlock uint
		MaxPOWBits                          uint32
		MaxPOWValue                         *big.Int
		GensisTimestamp                     uint32
		Enforce_CSV                         uint32 // if non zero CVS verifications will be enforced from this block onwards
		Enforce_SEGWIT                      uint32 // if non zero SEGWIT verifications will be enforced from this block onwards
		Enforce_UAHF                        uint32 // if non zero UAHF verifications will be enforced from this block onwards
		Enforce_DAA                         uint32 // if non zero DAA verifications will be enforced from this block onwards
		Enforce_MagneticAnomaly             uint32
		Enforce_GreatWall                   uint32
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
	BchBlockMinedCB  func(*bch.BchBlock) // used to remove mined txs from memory pool
}

// This is the very first function one should call in order to use this package
func NewChainExt(dbrootdir string, genesis *bch.Uint256, rescan bool, opts *NewChanOpts, bdbopts *BchBlockDBOpts) (ch *Chain) {
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
		// August 27, 2012 (Testnet) Introduction of Consensus Block and TX Versioning Mechanism
		ch.Consensus.BIP34Height = 21111 // 0000000023b3a96d3484e5abb3755c413e7d41500f8e2a5c3f0dd01299cd8ef8
		// October 31, 2015 (Testnet) Introduction of n-block/n-time-based locked transactions
		ch.Consensus.BIP65Height = 581885 // 00000000007f6655f22f98e72ed80d8b06dc761d5da09df0fa1dc4be4f861eb6
		// April 13, 2015 (Testnet) Introduction of strict ECDSA validation by Distinguished Encoding Rules (DER)
		ch.Consensus.BIP66Height = 330776 // 000000002104c8c45e99a8853285a3b592602a3ccde2b832481da85e9e4ba182
		// April 16, 2016 (Testnet) Introduction of CSV Check Sequence Verify, Relative n-(time/block)lock, 111 block median time inclusion.
		ch.Consensus.Enforce_CSV = 770112 // 00000000025e930139bac5c6c31a403776da130831ab85be56578f3fa75369bb
		// ** @todo THIS NEEDS TO BE REMOVED (BTC Only. Not active on BCH)
		ch.Consensus.Enforce_SEGWIT = 834624 // 00000000002b980fcd729daaa248fd9316a5200e9b367f4ff2c42453e84201ca
		// August 1, 2017 (Testnet) User Activated Hard Fork (UAHF) Active. Next Block (1155876) is First Bitcoin Cash Block on Test Network
		ch.Consensus.Enforce_UAHF = 1155875 // 00000000f17c850672894b9a75b63a1e72830bbd5f4c8889b5c1a80e7faef138
		ch.Consensus.Enforce_DAA = 1188697  // 0000000000170ed0918077bde7b4d36cc4c91be69fa09211f748240dabe047fb
		// Nov 15, 2018 Upcoming Bitcoin Cash scheduled hard fork
		ch.Consensus.Enforce_MagneticAnomaly = 1542300000 // 0000000000xxxxxxxxxxxxxxxxxxxxxxxxxxxxxtbd
		// Wed, 15 May 2019 12:00:00 UTC hard fork
		ch.Consensus.Enforce_GreatWall = 1557921600 // 0000000000xxxxxxxxxxxxxxxxxxxxxxxxxxxxxtbd
		ch.Consensus.BIP9_Treshold = 1512           // 00000000f57741182427e6ae2cc67e296ddc428d35a445a35e63a3228eb2c59f
	} else {
		// March 25, 2013 (BIP34) Introduction of Consensus Block and TX Versioning Mechanism
		// 		Gavin Andresen <gavinandresen@gmail.com>
		ch.Consensus.BIP34Height = 227931 // 000000000000024b89b42a942fe0d9fea3bb44ab7bd1b19115dd6a759c0808b8
		// December 15, 2015 (BIP65) Introduction of n-block/n-time-based locked transactions
		// 		Peter Todd <pete@petertodd.org>
		ch.Consensus.BIP65Height = 388381 // 000000000000000004c2b624ed5d7756c508d90fd0da2c7c679febfa6c4735f0
		// July 4, 2015 (BIP66) Introduction of strict ECDSA validation by Distinguished Encoding Rules (DER)
		// 		Pieter Wuille <pieter.wuille@gmail.com>
		ch.Consensus.BIP66Height = 363725 // 00000000000000000379eaa19dce8c9b722d46ae6a57c2f1a988119488b50931
		// July 5, 2016 (CSV) See Enforce_CSV [BIP68, BIP112, BIP113] notes below:
		// 	1. (BIP68) Introduction of relative n-block/n-time-based locked transactions
		// 		Mark Friedenbach <mark@friedenbach.org>
		// 		BtcDrak <btcdrak@gmail.com>
		// 		Nicolas Dorier <nicolas.dorier@gmail.com>
		//		kinoshitajona <kinoshitajona@gmail.com>
		// 	2. (BIP112) Introduction of opcode CHECKSEQUENCEVERIFY (aka 'CSV')
		// 		BtcDrak <btcdrak@gmail.com>
		// 		Mark Friedenbach <mark@friedenbach.org>
		// 		Eric Lombrozo <elombrozo@gmail.com>
		// 	3. (BIP113) Introduction of past 11 block timestamp median as method of time-lock transaction block inclusion eligibility
		// 		Thomas Kerin <me@thomaskerin.io>
		// 		Mark Friedenbach <mark@friedenbach.org>
		ch.Consensus.Enforce_CSV = 419328 // 000000000000000004a1b34462cb8aeebd5799177f7a29cf28f2d1961716b5b5
		// ** @todo THIS NEEDS TO BE REMOVED (BTC Only. Not active on BCH)
		ch.Consensus.Enforce_SEGWIT = 481824
		// August 1, 2017 (BCH) User Activated Hard Fork (UAHF) Active. Next Block (478559) is First Bitcoin Cash Block on Main Network
		ch.Consensus.Enforce_UAHF = 478558 // 0000000000000000011865af4122fe3b144e2cbeea86142e8ff2fb4107352d43
		// November 13, 2017 (DAA) Difficulty Adjustment Algorithm to replace Emergency Difficulty Adjustment (EDA)
		ch.Consensus.Enforce_DAA = 504031 // 0000000000000000011ebf65b60d0a3de80b8175be709d653b4c1a1beeb6ab9c
		// Nov 15, 2018 Upcoming Bitcoin Cash scheduled hard fork
		ch.Consensus.Enforce_MagneticAnomaly = 1542300000 // 0000000000xxxxxxxxxxxxxxxxxxxxxxxxxxxxxtbd
		// Wed, 15 May 2019 12:00:00 UTC hard fork
		ch.Consensus.Enforce_GreatWall = 1557921600 // 0000000000xxxxxxxxxxxxxxxxxxxxxxxxxxxxxtbd
		// July 23, 2017 (Mainnet) Introduce reduced Segwit threshold by Miner Activated Soft Fork (MAST) James Hilliard <james.hilliard1@gmail.com>
		ch.Consensus.BIP91Height = 477120 // 0000000000000000015411ca4b35f7b48ecab015b14de5627b647e262ba0ec40
		// January 26, 2009 (Mainnet) Introduction of mechanism for parallel "soft fork" deployment and orderly bit-flag space re-use
		// 		Pieter Wuille <pieter.wuille@gmail.com>
		// 		Peter Todd <pete@petertodd.org>
		// 		Greg Maxwell <greg@xiph.org>
		// 		Rusty Russell <rusty@rustcorp.com.au>
		ch.Consensus.BIP9_Treshold = 1916 // 00000000800cca5d11742408e3965a84424269df7cecca5896649b1521d22297
	}

	ch.BchBlocks = NewBlockDBExt(dbrootdir, bdbopts)

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
		ch.SetLast(ch.BchBlockTreeRoot)
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
	end, _ := ch.BchBlockTreeRoot.FindFarthestNode()
	if end.Height > ch.LastBlock().Height {
		ch.ParseTillBlock(end)
	} else {
		ch.Unspent.LastBlockHeight = end.Height
	}

	return
}

// Calculate an imaginary header of the genesis block (for Timestamp() and Bits() functions from chain_tree.go)
func (ch *Chain) RebuildGenesisHeader() {
	binary.LittleEndian.PutUint32(ch.BchBlockTreeRoot.BchBlockHeader[0:4], 1) // Version
	// [4:36] - prev_block
	// [36:68] - merkle_root
	binary.LittleEndian.PutUint32(ch.BchBlockTreeRoot.BchBlockHeader[68:72], ch.Consensus.GensisTimestamp) // Timestamp
	binary.LittleEndian.PutUint32(ch.BchBlockTreeRoot.BchBlockHeader[72:76], ch.Consensus.MaxPOWBits)      // Bits
	// [76:80] - nonce
}

// Call this function periodically (i.e. each second)
// when your client is idle, to defragment databases.
func (ch *Chain) Idle() bool {
	ch.BchBlocks.Idle()
	return ch.Unspent.Idle()
}

// Return blockchain stats in one string.
func (ch *Chain) Stats() (s string) {
	last := ch.LastBlock()
	ch.BchBlockIndexAccess.Lock()
	s = fmt.Sprintf("CHAIN: blocks:%d  Height:%d  MedianTime:%d\n",
		len(ch.BchBlockIndex), last.Height, last.GetMedianTimePast())
	ch.BchBlockIndexAccess.Unlock()
	s += ch.BchBlocks.GetStats()
	s += ch.Unspent.GetStats()
	return
}

// Close the databases.
func (ch *Chain) Close() {
	ch.BchBlocks.Close()
	ch.Unspent.Close()
}

// Returns true if we are on Testnet3 chain
func (ch *Chain) testnet() bool {
	return ch.Genesis.Hash[0] == 0x43 // it's simple, but works
}

// For SegWit2X
func (ch *Chain) MaxBlockWeight(height uint32) uint {
	if ch.Consensus.S2XHeight != 0 && height >= ch.Consensus.S2XHeight {
		return 2 * bch.MAX_BLOCK_WEIGHT
	} else {
		return bch.MAX_BLOCK_WEIGHT
	}
}

// For SegWit2X
func (ch *Chain) MaxBlockSigopsCost(height uint32) uint32 {
	if ch.Consensus.S2XHeight != 0 && height >= ch.Consensus.S2XHeight {
		return 2 * bch.MAX_BLOCK_SIGOPS_COST
	} else {
		return bch.MAX_BLOCK_SIGOPS_COST
	}
}

func (ch *Chain) LastBlock() (res *BchBlockTreeNode) {
	ch.blockTreeAccess.Lock()
	res = ch.blockTreeEnd
	ch.blockTreeAccess.Unlock()
	return
}

func (ch *Chain) SetLast(val *BchBlockTreeNode) {
	ch.blockTreeAccess.Lock()
	ch.blockTreeEnd = val
	ch.blockTreeAccess.Unlock()
	return
}
