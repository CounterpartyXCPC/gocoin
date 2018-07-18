package btc

import (
	"sync"
	"bytes"
	"errors"
	"encoding/binary"
)

type Block struct {
	Raw []byte
	Hash *Uint256
	Txs []*Tx
	TxCount, TxOffset int  // Number of transactions and byte offset to the first one
	Trusted bool // if the block is trusted, we do not check signatures and some other things...
	LastKnownHeight uint32

	BlockExtraInfo // If we cache block on disk (between downloading and comitting), this data has to be preserved

	MedianPastTime uint32 // Set in PreCheckBlock() .. last used in PostCheckBlock()

	// These flags are set in BuildTxList() used later (e.g. by script.VerifyTxScript):
	BlockWeight uint
	TotalInputs int
}


type BlockExtraInfo struct {
	VerifyFlags uint32
	Height uint32
}


func NewBlock(data []byte) (bl *Block, er error) {
	if data==nil {
		er = errors.New("nil pointer")
		return
	}
	bl = new(Block)
	bl.Hash = NewSha2Hash(data[:80])
	er = bl.UpdateContent(data)
	return
}


func (bl *Block) UpdateContent(data []byte) error {
	if len(data)<81 {
		return errors.New("Block too short")
	}
	bl.Raw = data
	bl.TxCount, bl.TxOffset = VLen(data[80:])
	if bl.TxOffset == 0 {
		return errors.New("Block's txn_count field corrupt - RPC_Result:bad-blk-length")
	}
	bl.TxOffset += 80
	return nil
}

func (bl *Block)Version() uint32 {
	return binary.LittleEndian.Uint32(bl.Raw[0:4])
}

func (bl *Block)ParentHash() []byte {
	return bl.Raw[4:36]
}

func (bl *Block)MerkleRoot() []byte {
	return bl.Raw[36:68]
}

func (bl *Block)BlockTime() uint32 {
	return binary.LittleEndian.Uint32(bl.Raw[68:72])
}

func (bl *Block)Bits() uint32 {
	return binary.LittleEndian.Uint32(bl.Raw[72:76])
}


// Parses block's transactions and adds them to the structure, calculating hashes BTW.
// It would be more elegant to use bytes.Reader here, but this solution is ~20% faster.
func (bl *Block) BuildTxList() (e error) {
	if bl.TxCount==0 {
		bl.TxCount, bl.TxOffset = VLen(bl.Raw[80:])
		if bl.TxCount==0 || bl.TxOffset==0 {
			e = errors.New("Block's txn_count field corrupt - RPC_Result:bad-blk-length")
			return
		}
		bl.TxOffset += 80
	}
	bl.Txs = make([]*Tx, bl.TxCount)

	offs := bl.TxOffset

	var wg sync.WaitGroup
	var data2hash []byte

	for i := 0; i < bl.TxCount; i++ {
		var n int
		bl.Txs[i], n = NewTx(bl.Raw[offs:])
		if bl.Txs[i] == nil || n==0 {
			e = errors.New("NewTx failed")
			break
		}
		bl.Txs[i].Raw = bl.Raw[offs:offs+n]
		bl.Txs[i].Size = uint32(n)
		if i == 0 {
			for _, ou := range bl.Txs[0].TxOut {
				ou.WasCoinbase = true
			}
		} else {
			// Coinbase tx does not have an input
			bl.TotalInputs += len(bl.Txs[i].TxIn)
		}
		
                data2hash = bl.Txs[i].Raw
				
		wg.Add(1)
		go func(tx *Tx, b, w []byte) {
			tx.Hash.Calc(b) // Calculate tx hash in a background
			if w != nil {
				tx.wTxID.Calc(w)
			}
			wg.Done()
		}(bl.Txs[i], data2hash)
		offs += n
	}

	wg.Wait()

	return
}





func GetBlockReward(height uint32) (uint64) {
	return 50e8 >> (height/210000)
}


func (bl *Block) MerkleRootMatch() bool {
	if bl.TxCount==0 {
		return false
	}
	merkle, mutated := bl.GetMerkle()
	return !mutated && bytes.Equal(merkle, bl.MerkleRoot())
}

func (bl *Block) GetMerkle() (res []byte, mutated bool) {
	mtr := make([][32]byte, len(bl.Txs), 3*len(bl.Txs)) // make the buffer 3 times longer as we use append() inside CalcMerkle
	for i, tx := range bl.Txs {
		mtr[i] = tx.Hash.Hash
	}
	res, mutated = CalcMerkle(mtr)
	return
}
