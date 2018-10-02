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

// File:		block.go
// Description:	Bictoin Cash Block Package

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

package bch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
)

type BchBlock struct {
	Raw               []byte
	Hash              *Uint256
	Txs               []*Tx
	TxCount, TxOffset int  // Number of transactions and byte offset to the first one
	Trusted           bool // if the block is trusted, we do not check signatures and some other things...
	LastKnownHeight   uint32

	// 100's / 1000's blocks [not sequenced] *******
	// Meantime you receive randoms. Memory flooding. Disk cache structure for storing on disk. (So as to not completely exhaust RAM)

	BchBlockExtraInfo        // If we cache block on disk (between downloading and comitting), this data has to be preserved
	MedianPastTime    uint32 // Set in PreCheckBlock() .. last used in PostCheckBlock()

	// These flags are set in BuildTxList() used later (e.g. by script.VerifyTxScript):
	NoWitnessSize  int
	BchBlockWeight uint
	TotalInputs    int

	NoWitnessData []byte // This is set by BuildNoWitnessData()
}

type BchBlockExtraInfo struct {
	VerifyFlags uint32
	Height      uint32
}

func BchNewBlock(data []byte) (bl *BchBlock, er error) {
	if data == nil {
		er = errors.New("nil pointer")
		return
	}
	bl = new(BchBlock)
	bl.Hash = NewSha2Hash(data[:80])
	er = bl.UpdateContent(data)
	return
}

func (bl *BchBlock) UpdateContent(data []byte) error {
	if len(data) < 81 {
		return errors.New("BCH Block too short")
	}
	bl.Raw = data
	bl.TxCount, bl.TxOffset = VLen(data[80:])
	if bl.TxOffset == 0 {
		return errors.New("BCH Block's txn_count field corrupt - RPC_Result:bad-blk-length")
	}
	bl.TxOffset += 80
	return nil
}

func (bl *BchBlock) Version() uint32 {
	return binary.LittleEndian.Uint32(bl.Raw[0:4])
}

func (bl *BchBlock) ParentHash() []byte {
	return bl.Raw[4:36]
}

func (bl *BchBlock) MerkleRoot() []byte {
	return bl.Raw[36:68]
}

func (bl *BchBlock) BchBlockTime() uint32 {
	return binary.LittleEndian.Uint32(bl.Raw[68:72])
}

func (bl *BchBlock) Bits() uint32 {
	return binary.LittleEndian.Uint32(bl.Raw[72:76])
}

// Parses block's transactions and adds them to the structure, calculating hashes BTW.
// It would be more elegant to use bytes.Reader here, but this solution is ~20% faster.
func (bl *BchBlock) BuildTxList() (e error) {
	if bl.TxCount == 0 {
		bl.TxCount, bl.TxOffset = VLen(bl.Raw[80:])
		if bl.TxCount == 0 || bl.TxOffset == 0 {
			e = errors.New("Block's txn_count field corrupt - RPC_Result:bad-blk-length")
			return
		}
		bl.TxOffset += 80
	}
	bl.Txs = make([]*Tx, bl.TxCount)

	offs := bl.TxOffset

	var wg sync.WaitGroup
	var data2hash, witness2hash []byte

	bl.NoWitnessSize = 80 + VLenSize(uint64(bl.TxCount))
	bl.BchBlockWeight = 4 * uint(bl.NoWitnessSize)

	for i := 0; i < bl.TxCount; i++ {
		var n int
		bl.Txs[i], n = NewTx(bl.Raw[offs:])
		if bl.Txs[i] == nil || n == 0 {
			e = errors.New("NewTx failed")
			break
		}
		bl.Txs[i].Raw = bl.Raw[offs : offs+n]
		bl.Txs[i].Size = uint32(n)
		if i == 0 {
			for _, ou := range bl.Txs[0].TxOut {
				ou.WasCoinbase = true
			}
		} else {
			// Coinbase tx does not have an input
			bl.TotalInputs += len(bl.Txs[i].TxIn)
		}
		if bl.Txs[i].SegWit != nil {
			data2hash = bl.Txs[i].Serialize()
			bl.Txs[i].NoWitSize = uint32(len(data2hash))
			if i > 0 {
				witness2hash = bl.Txs[i].Raw
			}
		} else {
			data2hash = bl.Txs[i].Raw
			bl.Txs[i].NoWitSize = bl.Txs[i].Size
			witness2hash = nil
		}
		bl.BchBlockWeight += uint(3*bl.Txs[i].NoWitSize + bl.Txs[i].Size)
		bl.NoWitnessSize += len(data2hash)
		wg.Add(1)
		go func(tx *Tx, b, w []byte) {
			tx.Hash.Calc(b) // Calculate tx hash in a background
			if w != nil {
				tx.wTxID.Calc(w)
			}
			wg.Done()
		}(bl.Txs[i], data2hash, witness2hash)
		offs += n
	}

	wg.Wait()

	return
}

// The block data in non-segwit format
func (bl *BchBlock) BuildNoWitnessData() (e error) {
	if bl.TxCount == 0 {
		e = bl.BuildTxList()
		if e != nil {
			return
		}
	}
	old_format_block := new(bytes.Buffer)
	old_format_block.Write(bl.Raw[:80])
	WriteVlen(old_format_block, uint64(bl.TxCount))
	for _, tx := range bl.Txs {
		tx.WriteSerialized(old_format_block)
	}
	bl.NoWitnessData = old_format_block.Bytes()
	if bl.NoWitnessSize == 0 {
		bl.NoWitnessSize = len(bl.NoWitnessData)
	} else if bl.NoWitnessSize != len(bl.NoWitnessData) {
		panic("NoWitnessSize corrupt")
	}
	return
}

func GetBlockReward(height uint32) uint64 {
	return 50e8 >> (height / 210000)
}

func (bl *BchBlock) MerkleRootMatch() bool {
	if bl.TxCount == 0 {
		return false
	}
	merkle, mutated := bl.GetMerkle()
	return !mutated && bytes.Equal(merkle, bl.MerkleRoot())
}

func (bl *BchBlock) GetMerkle() (res []byte, mutated bool) {
	mtr := make([][32]byte, len(bl.Txs), 3*len(bl.Txs)) // make the buffer 3 times longer as we use append() inside CalcMerkle
	for i, tx := range bl.Txs {
		mtr[i] = tx.Hash.Hash
	}
	res, mutated = CalcMerkle(mtr)
	return
}
