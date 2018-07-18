package chain

import (
	"fmt"
	"time"
	"bytes"
	"errors"
	"encoding/binary"
	"github.com/piotrnar/gocoin/lib/btc"
	"github.com/piotrnar/gocoin/lib/script"
)

// Make sure to call this function with ch.BlockIndexAccess locked
func (ch *Chain) PreCheckBlock(bl *btc.Block) (er error, dos bool, maybelater bool) {
	// Size limits
	if len(bl.Raw)<81 {
		er = errors.New("CheckBlock() : size limits failed - RPC_Result:bad-blk-length")
		dos = true
		return
	}

	ver := bl.Version()
	if ver == 0 {
		er = errors.New("CheckBlock() : Block version 0 not allowed - RPC_Result:bad-version")
		dos = true
		return
	}

	// Check proof-of-work
	if !btc.CheckProofOfWork(bl.Hash, bl.Bits()) {
		er = errors.New("CheckBlock() : proof of work failed - RPC_Result:high-hash")
		dos = true
		return
	}

	// Check timestamp (must not be higher than now +2 hours)
	if int64(bl.BlockTime()) > time.Now().Unix() + 2 * 60 * 60 {
		er = errors.New("CheckBlock() : block timestamp too far in the future - RPC_Result:time-too-new")
		dos = true
		return
	}

	if prv, pres := ch.BlockIndex[bl.Hash.BIdx()]; pres {
		if prv.Parent == nil {
			// This is genesis block
			er = errors.New("Genesis")
			return
		} else {
			er = errors.New("CheckBlock: "+bl.Hash.String()+" already in - RPC_Result:duplicate")
			return
		}
	}

	prevblk, ok := ch.BlockIndex[btc.NewUint256(bl.ParentHash()).BIdx()]
	if !ok {
		er = errors.New("CheckBlock: "+bl.Hash.String()+" parent not found - RPC_Result:bad-prevblk")
		maybelater = true
		return
	}

	bl.Height = prevblk.Height+1

	// Reject the block if it reaches into the chain deeper than our unwind buffer
	lst_now := ch.LastBlock()
	if prevblk != lst_now && int(lst_now.Height)-int(bl.Height) >= MovingCheckopintDepth {
		er = errors.New(fmt.Sprint("CheckBlock: btc.Block ", bl.Hash.String(),
			" hooks too deep into the chain: ", bl.Height, "/", lst_now.Height, " ",
			btc.NewUint256(bl.ParentHash()).String(), " - RPC_Result:bad-prevblk"))
		return
	}

	// Check proof of work
	gnwr := ch.GetNextWorkRequired(prevblk, bl.BlockTime())
	if bl.Bits() != gnwr {
		er = errors.New("CheckBlock: incorrect proof of work - RPC_Result:bad-diffbits")
		dos = true
		return
	}

	// Check timestamp against prev
	bl.MedianPastTime = prevblk.GetMedianTimePast()
	if bl.BlockTime() <= bl.MedianPastTime {
		er = errors.New("CheckBlock: block's timestamp is too early - RPC_Result:time-too-old")
		dos = true
		return
	}

	if ver < 2 && bl.Height >= ch.Consensus.BIP34Height ||
		ver < 3 && bl.Height >= ch.Consensus.BIP66Height ||
		ver < 4 && bl.Height >= ch.Consensus.BIP65Height {
		// bad block version
		erstr := fmt.Sprintf("0x%08x", ver)
		er = errors.New("CheckBlock() : Rejected Version="+erstr+" block - RPC_Result:bad-version("+erstr+")")
		dos = true
		return
	}

	return
}


func (ch *Chain) ApplyBlockFlags(bl *btc.Block) {
	if bl.BlockTime() >= BIP16SwitchTime {
		bl.VerifyFlags = script.VER_P2SH
	} else {
		bl.VerifyFlags = 0
	}

	if bl.Height >= ch.Consensus.BIP66Height {
		bl.VerifyFlags |= script.VER_DERSIG
	}

	if bl.Height >= ch.Consensus.BIP65Height {
		bl.VerifyFlags |= script.VER_CLTV
	}

	if ch.Consensus.Enforce_CSV != 0 && bl.Height >= ch.Consensus.Enforce_CSV {
		bl.VerifyFlags |= script.VER_CSV
	}
}


func (ch *Chain) PostCheckBlock(bl *btc.Block) (er error) {
	// Size limits
	if len(bl.Raw)<81 {
		er = errors.New("CheckBlock() : size limits failed low - RPC_Result:bad-blk-length")
		return
	}

	if bl.Txs==nil {
		er = bl.BuildTxList()
		if er != nil {
			return
		}
		if bl.BlockWeight > ch.MaxBlockWeight(bl.Height) {
			er = errors.New("CheckBlock() : weight limits failed - RPC_Result:bad-blk-weight")
			return
		}
		//fmt.Println("New block", bl.Height, " Weight:", bl.BlockWeight, " Raw:", len(bl.Raw))
	}

	if !bl.Trusted {
		// We need to be satoshi compatible
		if len(bl.Txs)==0 || !bl.Txs[0].IsCoinBase() {
			er = errors.New("CheckBlock() : first tx is not coinbase: "+bl.Hash.String()+" - RPC_Result:bad-cb-missing")
			return
		}

		// Enforce rule that the coinbase starts with serialized block height
		if bl.Height>=ch.Consensus.BIP34Height {
			var exp [6]byte
			var exp_len int
			binary.LittleEndian.PutUint32(exp[1:5], bl.Height)
			for exp_len=5; exp_len>1; exp_len-- {
				if exp[exp_len]!=0 || exp[exp_len-1]>=0x80 {
					break
				}
			}
			exp[0] = byte(exp_len)
			exp_len++

			if !bytes.HasPrefix(bl.Txs[0].TxIn[0].ScriptSig, exp[:exp_len]) {
				er = errors.New("CheckBlock() : Unexpected block number in coinbase: "+bl.Hash.String()+" - RPC_Result:bad-cb-height")
				return
			}
		}

		// And again...
		for i:=1; i<len(bl.Txs); i++ {
			if bl.Txs[i].IsCoinBase() {
				er = errors.New("CheckBlock() : more than one coinbase: "+bl.Hash.String()+" - RPC_Result:bad-cb-multiple")
				return
			}
		}
	}

	// Check Merkle Root, even for trusted blocks - that's important, as they may come from untrasted peers
	merkle, mutated := bl.GetMerkle()
	if mutated {
		er = errors.New("CheckBlock(): duplicate transaction - RPC_Result:bad-txns-duplicate")
		return
	}

	if !bytes.Equal(merkle, bl.MerkleRoot()) {
		er = errors.New("CheckBlock() : Merkle Root mismatch - RPC_Result:bad-txnmrklroot")
		return
	}

	ch.ApplyBlockFlags(bl)

	if !bl.Trusted {
		var blockTime uint32
		
		if (bl.VerifyFlags&script.VER_CSV) != 0 {
			blockTime = bl.MedianPastTime
		} else {
			blockTime = bl.BlockTime()
		}

		// Check transactions - this is the most time consuming task
		er = CheckTransactions(bl.Txs, bl.Height, blockTime)
	}
	return
}


func (ch *Chain) CheckBlock(bl *btc.Block) (er error, dos bool, maybelater bool) {
	er, dos, maybelater = ch.PreCheckBlock(bl)
	if er == nil {
		er = ch.PostCheckBlock(bl)
		if er != nil { // all post-check errors are DoS kind
			dos = true
		}
	}
	return
}
