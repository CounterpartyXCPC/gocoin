package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
	"github.com/dchest/siphash"
)

var (
	CompactBlocksMutex sync.Mutex
)

type CmpctBlockCollector struct {
	Header  []byte
	Txs     []interface{} // either []byte of uint64
	K0, K1  uint64
	Sid2idx map[uint64]int
	Missing int
}

func ShortIDToU64(d []byte) uint64 {
	return uint64(d[0]) | (uint64(d[1]) << 8) | (uint64(d[2]) << 16) |
		(uint64(d[3]) << 24) | (uint64(d[4]) << 32) | (uint64(d[5]) << 40)
}

func (col *CmpctBlockCollector) Assemble() []byte {
	bdat := new(bytes.Buffer)
	bdat.Write(col.Header)
	bch.WriteVlen(bdat, uint64(len(col.Txs)))
	for _, txd := range col.Txs {
		bdat.Write(txd.([]byte))
	}
	return bdat.Bytes()
}

func GetchBlockForBIP152(hash *bch.Uint256) (crec *bch_chain.BlckCachRec) {
	CompactBlocksMutex.Lock()
	defer CompactBlocksMutex.Unlock()

	crec, _, _ = common.BchBlockChain.BchBlocks.BchBlockGetExt(hash)
	if crec == nil {
		//fmt.Println("BlockGetExt failed for", hash.String(), er.Error())
		return
	}

	if crec.BchBlock == nil {
		crec.BchBlock, _ = bch.NewBchBlock(crec.Data)
		if crec.BchBlock == nil {
			fmt.Println("GetchBlockForBIP152: bch.NewBchBlock() failed for", hash.String())
			return
		}
	}

	if len(crec.BchBlock.Txs) == 0 {
		if crec.BchBlock.BuildTxList() != nil {
			fmt.Println("GetchBlockForBIP152: bl.BuildTxList() failed for", hash.String())
			return
		}
	}

	if len(crec.BIP152) != 24 {
		crec.BIP152 = make([]byte, 24)
		copy(crec.BIP152[:8], crec.Data[48:56]) // set the nonce to 8 middle-bytes of block's merkle_root
		sha := sha256.New()
		sha.Write(crec.Data[:80])
		sha.Write(crec.BIP152[:8])
		copy(crec.BIP152[8:24], sha.Sum(nil)[0:16])
	}

	return
}

func (c *OneConnection) SendCmpctBlk(hash *bch.Uint256) bool {
	crec := GetchBlockForBIP152(hash)
	if crec == nil {
		//fmt.Println(c.ConnID, "cmpctblock not sent:", c.Node.Agent, hash.String())
		return false
	}

	k0 := binary.LittleEndian.Uint64(crec.BIP152[8:16])
	k1 := binary.LittleEndian.Uint64(crec.BIP152[16:24])

	msg := new(bytes.Buffer)
	msg.Write(crec.Data[:80])
	msg.Write(crec.BIP152[:8])
	bch.WriteVlen(msg, uint64(len(crec.BchBlock.Txs)-1)) // all except coinbase
	for i := 1; i < len(crec.BchBlock.Txs); i++ {
		var lsb [8]byte
		var hasz *bch.Uint256
		if c.Node.SendCmpctVer == 2 {
			hasz = crec.BchBlock.Txs[i].WTxID()
		} else {
			hasz = &crec.BchBlock.Txs[i].Hash
		}
		binary.LittleEndian.PutUint64(lsb[:], siphash.Hash(k0, k1, hasz.Hash[:]))
		msg.Write(lsb[:6])
	}
	msg.Write([]byte{1}) // one preffiled tx
	msg.Write([]byte{0}) // coinbase - index 0
	if c.Node.SendCmpctVer == 2 {
		msg.Write(crec.BchBlock.Txs[0].Raw) // coinbase - index 0
	} else {
		crec.BchBlock.Txs[0].WriteSerialized(msg) // coinbase - index 0
	}
	c.SendRawMsg("cmpctblock", msg.Bytes())
	return true
}

func (c *OneConnection) ProcessGetBlockTxn(pl []byte) {
	if len(pl) < 34 {
		println(c.ConnID, "GetBlockTxnShort")
		c.DoS("GetBlockTxnShort")
		return
	}
	hash := bch.NewUint256(pl[:32])
	crec := GetchBlockForBIP152(hash)
	if crec == nil {
		fmt.Println(c.ConnID, "GetBlockTxn aborting for", hash.String())
		return
	}

	req := bytes.NewReader(pl[32:])
	indexes_length, _ := bch.ReadVLen(req)
	if indexes_length == 0 {
		println(c.ConnID, "GetBlockTxnEmpty")
		c.DoS("GetBlockTxnEmpty")
		return
	}

	var exp_idx uint64
	msg := new(bytes.Buffer)

	msg.Write(hash.Hash[:])
	bch.WriteVlen(msg, indexes_length)

	for {
		idx, er := bch.ReadVLen(req)
		if er != nil {
			println(c.ConnID, "GetBlockTxnERR")
			c.DoS("GetBlockTxnERR")
			return
		}
		idx += exp_idx
		if int(idx) >= len(crec.BchBlock.Txs) {
			println(c.ConnID, "GetBlockTxnIdx+")
			c.DoS("GetBlockTxnIdx+")
			return
		}
		if c.Node.SendCmpctVer == 2 {
			msg.Write(crec.BchBlock.Txs[idx].Raw) // coinbase - index 0
		} else {
			crec.BchBlock.Txs[idx].WriteSerialized(msg) // coinbase - index 0
		}
		if indexes_length == 1 {
			break
		}
		indexes_length--
		exp_idx = idx + 1
	}

	c.SendRawMsg("blocktxn", msg.Bytes())
}

func delB2G_callback(hash *bch.Uint256) {
	DelB2G(hash.BIdx())
}

func (c *OneConnection) ProcessCmpctBlock(pl []byte) {
	println("ProcessCmpctBlock")
	if len(pl) < 90 {
		println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "cmpctblock error A", hex.EncodeToString(pl))
		c.DoS("CmpctBlkErrA")
		return
	}

	MutexRcv.Lock()
	defer MutexRcv.Unlock()

	var tmp_hdr [81]byte
	copy(tmp_hdr[:80], pl[:80])
	sta, b2g := c.ProcessNewHeader(tmp_hdr[:]) // ProcessNewHeader() needs byte(0) after the header,
	// but don't try to change it to ProcessNewHeader(append(pl[:80], 0)) as it'd overwrite pl[80]

	if b2g == nil {
		common.CountSafe("CmpctBlockHdrNo")
		if sta == PH_STATUS_ERROR {
			c.ReceiveHeadersNow()       // block doesn't connect so ask for the headers
			c.Misbehave("BadCmpct", 50) // do it 20 times and you are banned
		} else if sta == PH_STATUS_FATAL {
			c.DoS("BadCmpct")
		}
		return
	}
	if sta == PH_STATUS_NEW {
		b2g.SendInvs = true
	}

	if common.BchBlockChain.Consensus.Enforce_SEGWIT != 0 && c.Node.SendCmpctVer < 2 {
		if b2g.BchBlock.Height >= common.BchBlockChain.Consensus.Enforce_SEGWIT {
			common.CountSafe("CmpctBlockIgnore")
			println("Ignore compact block", b2g.BchBlock.Height, "from non-segwit node", c.ConnID)
			if (c.Node.Services & SERVICE_SEGWIT) != 0 {
				// it only makes sense to ask this node for block's data, if it supports segwit
				c.MutexSetBool(&c.X.GetBlocksDataNow, true)
			}
			return
		}
	}

	// if we got here, we shall download this block
	if c.Node.Height < b2g.BchBlock.Height {
		c.Mutex.Lock()
		c.Node.Height = b2g.BchBlock.Height
		c.Mutex.Unlock()
	}

	if b2g.InProgress >= uint(common.CFG.Net.MaxBlockAtOnce) {
		common.CountSafe("CmpctBlockMaxInProg")
		//fmt.Println(c.ConnID, " - too many in progress")
		return
	}

	var n, idx, shortidscnt, shortidx_idx, prefilledcnt int

	col := new(CmpctBlockCollector)
	col.Header = b2g.BchBlock.Raw[:80]

	offs := 88
	shortidscnt, n = bch.VLen(pl[offs:])
	if shortidscnt < 0 || n > 3 {
		println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "cmpctblock error B", hex.EncodeToString(pl))
		c.DoS("CmpctBlkErrB")
		return
	}
	offs += n
	shortidx_idx = offs
	shortids := make(map[uint64][]byte, shortidscnt)
	for i := 0; i < int(shortidscnt); i++ {
		if len(pl[offs:offs+6]) < 6 {
			println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "cmpctblock error B2", hex.EncodeToString(pl))
			c.DoS("CmpctBlkErrB2")
			return
		}
		shortids[ShortIDToU64(pl[offs:offs+6])] = nil
		offs += 6
	}

	prefilledcnt, n = bch.VLen(pl[offs:])
	if prefilledcnt < 0 || n > 3 {
		println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "cmpctblock error C", hex.EncodeToString(pl))
		c.DoS("CmpctBlkErrC")
		return
	}
	offs += n

	col.Txs = make([]interface{}, prefilledcnt+shortidscnt)

	exp := 0
	for i := 0; i < int(prefilledcnt); i++ {
		idx, n = bch.VLen(pl[offs:])
		if idx < 0 || n > 3 {
			println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "cmpctblock error D", hex.EncodeToString(pl))
			c.DoS("CmpctBlkErrD")
			return
		}
		idx += exp
		offs += n
		n = bch.TxSize(pl[offs:])
		if n == 0 {
			println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "cmpctblock error E", hex.EncodeToString(pl))
			c.DoS("CmpctBlkErrE")
			return
		}
		col.Txs[idx] = pl[offs : offs+n]
		offs += n
		exp = int(idx) + 1
	}

	// calculate K0 and K1 params for siphash-4-2
	sha := sha256.New()
	sha.Write(pl[:88])
	kks := sha.Sum(nil)
	col.K0 = binary.LittleEndian.Uint64(kks[0:8])
	col.K1 = binary.LittleEndian.Uint64(kks[8:16])

	var cnt_found int

	TxMutex.Lock()

	for _, v := range TransactionsToSend {
		var hash2take *bch.Uint256
		if c.Node.SendCmpctVer == 2 {
			hash2take = v.Tx.WTxID()
		} else {
			hash2take = &v.Tx.Hash
		}
		sid := siphash.Hash(col.K0, col.K1, hash2take.Hash[:]) & 0xffffffffffff
		if ptr, ok := shortids[sid]; ok {
			if ptr != nil {
				common.CountSafe("ShortIDSame")
				println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "Same short ID - abort")
				return
			}
			shortids[sid] = v.Raw
			cnt_found++
		}
	}

	for _, v := range TransactionsRejected {
		if v.Tx == nil {
			continue
		}
		var hash2take *bch.Uint256
		if c.Node.SendCmpctVer == 2 {
			hash2take = v.WTxID()
		} else {
			hash2take = &v.Hash
		}
		sid := siphash.Hash(col.K0, col.K1, hash2take.Hash[:]) & 0xffffffffffff
		if ptr, ok := shortids[sid]; ok {
			if ptr != nil {
				common.CountSafe("ShortIDSame")
				println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "Same short ID - abort")
				return
			}
			shortids[sid] = v.Raw
			cnt_found++
			common.CountSafe(fmt.Sprint("CmpctBlkUseRej-", v.Reason))
		}
	}

	var msg *bytes.Buffer

	missing := len(shortids) - cnt_found
	//fmt.Println(c.ConnID, c.Node.SendCmpctVer, "ShortIDs", cnt_found, "/", shortidscnt, "  Prefilled", prefilledcnt, "  Missing", missing, "  MemPool:", len(TransactionsToSend))
	col.Missing = missing
	if missing > 0 {
		msg = new(bytes.Buffer)
		msg.Write(b2g.BchBlock.Hash.Hash[:])
		bch.WriteVlen(msg, uint64(missing))
		exp = 0
		col.Sid2idx = make(map[uint64]int, missing)
	}
	for n = 0; n < len(col.Txs); n++ {
		switch col.Txs[n].(type) {
		case []byte: // prefilled transaction

		default:
			sid := ShortIDToU64(pl[shortidx_idx : shortidx_idx+6])
			if ptr, ok := shortids[sid]; ok {
				if ptr != nil {
					col.Txs[n] = ptr
				} else {
					col.Txs[n] = sid
					col.Sid2idx[sid] = n
					if missing > 0 {
						bch.WriteVlen(msg, uint64(n-exp))
						exp = n + 1
					}
				}
			} else {
				panic(fmt.Sprint("Tx idx ", n, " is missing - this should not happen!!!"))
			}
			shortidx_idx += 6
		}
	}
	TxMutex.Unlock()

	if missing == 0 {
		//sta := time.Now()
		b2g.BchBlock.UpdateContent(col.Assemble())
		//sto := time.Now()
		bidx := b2g.BchBlock.Hash.BIdx()
		er := common.BchBlockChain.PostCheckBlock(b2g.BchBlock)
		if er != nil {
			println(c.ConnID, "Corrupt CmpctBlkA")
			ioutil.WriteFile(b2g.Hash.String()+".bin", b2g.BchBlock.Raw, 0700)

			if b2g.BchBlock.MerkleRootMatch() {
				println("It was a wrongly mined one - clean it up")
				DelB2G(bidx) //remove it from BlocksToGet
				if b2g.BchBlockTreeNode == LastCommitedHeader {
					LastCommitedHeader = LastCommitedHeader.Parent
				}
				common.BchBlockChain.DeleteBranch(b2g.BchBlockTreeNode, delB2G_callback)
			}

			//c.DoS("BadCmpctBlockA")
			return
		}
		//fmt.Println(c.ConnID, "Instatnt PostCheckBlock OK #", b2g.BchBlock.Height, sto.Sub(sta), time.Now().Sub(sta))
		c.Mutex.Lock()
		c.counters["NewCBlock"]++
		c.blocksreceived = append(c.blocksreceived, time.Now())
		c.Mutex.Unlock()
		orb := &OneReceivedBlock{TmStart: b2g.Started, TmPreproc: time.Now(), FromConID: c.ConnID, DoInvs: b2g.SendInvs}
		ReceivedBlocks[bidx] = orb
		DelB2G(bidx) //remove it from BlocksToGet if no more pending downloads
		if c.X.Authorized {
			b2g.BchBlock.Trusted = true
		}
		NetBlocks <- &BlockRcvd{Conn: c, BchBlock: b2g.BchBlock, BchBlockTreeNode: b2g.BchBlockTreeNode, OneReceivedBlock: orb}
	} else {
		if b2g.TmPreproc.IsZero() { // do not overwrite TmPreproc if already set
			b2g.TmPreproc = time.Now()
		}
		b2g.InProgress++
		c.Mutex.Lock()
		c.GetBlockInProgress[b2g.BchBlock.Hash.BIdx()] = &oneBlockDl{hash: b2g.BchBlock.Hash, start: time.Now(), col: col, SentAtPingCnt: c.X.PingSentCnt}
		c.Mutex.Unlock()
		c.SendRawMsg("getblocktxn", msg.Bytes())
		//fmt.Println(c.ConnID, "Send getblocktxn for", col.Missing, "/", shortidscnt, "missing txs.  ", msg.Len(), "bytes")
	}
}

func (c *OneConnection) ProcessBlockTxn(pl []byte) {
	if len(pl) < 33 {
		println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "blocktxn error A", hex.EncodeToString(pl))
		c.DoS("BlkTxnErrLen")
		return
	}
	hash := bch.NewUint256(pl[:32])
	le, n := bch.VLen(pl[32:])
	if le < 0 || n > 3 {
		println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "blocktxn error B", hex.EncodeToString(pl))
		c.DoS("BlkTxnErrCnt")
		return
	}
	MutexRcv.Lock()
	defer MutexRcv.Unlock()

	idx := hash.BIdx()

	c.Mutex.Lock()
	bip := c.GetBlockInProgress[idx]
	if bip == nil {
		println(time.Now().Format("2006-01-02 15:04:05"), c.ConnID, "BlkTxnNoBIP:", c.PeerAddr.Ip(), c.Node.Agent, hash.String())
		c.Mutex.Unlock()
		c.counters["BlkTxnNoBIP"]++
		c.Misbehave("BlkTxnErrBip", 100)
		return
	}
	col := bip.col
	if col == nil {
		c.Mutex.Unlock()
		println("BlkTxnNoCOL:", c.PeerAddr.Ip(), c.Node.Agent, hash.String())
		common.CountSafe("UnxpectedBlockTxn")
		c.counters["BlkTxnNoCOL"]++
		c.Misbehave("BlkTxnNoCOL", 100)
		return
	}
	delete(c.GetBlockInProgress, idx)
	c.Mutex.Unlock()

	// the blocks seems to be fine
	if rb, got := ReceivedBlocks[idx]; got {
		rb.Cnt++
		common.CountSafe("BlkTxnSameRcvd")
		//fmt.Println(c.ConnID, "BlkTxn size", len(pl), "for", hash.String()[48:],"- already have")
		return
	}

	b2g := BlocksToGet[idx]
	if b2g == nil {
		panic("BlockTxn: Block missing from BlocksToGet")
		return
	}
	//b2g.InProgress--

	//fmt.Println(c.ConnID, "BlockTxn size", len(pl), "-", le, "new txs for block #", b2g.BchBlock.Height)

	offs := 32 + n
	for offs < len(pl) {
		n = bch.TxSize(pl[offs:])
		if n == 0 {
			println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "blocktxn corrupt TX")
			c.DoS("BlkTxnErrTx")
			return
		}
		raw_tx := pl[offs : offs+n]
		var tx_hash bch.Uint256
		tx_hash.Calc(raw_tx)
		if common.GetBool(&common.CFG.TXPool.Debug) {
			if f, _ := os.OpenFile("missing_txs.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660); f != nil {
				_tx, _ := bch.NewTx(raw_tx)
				_tx.SetHash(raw_tx)
				fmt.Fprintf(f, "%s: Tx %s was missing in bock %d\n",
					time.Now().Format("2006-01-02 15:04:05"), _tx.Hash.String(), b2g.BchBlock.Height)
				f.Close()
			}
		}
		offs += n

		sid := siphash.Hash(col.K0, col.K1, tx_hash.Hash[:]) & 0xffffffffffff
		if idx, ok := col.Sid2idx[sid]; ok {
			col.Txs[idx] = raw_tx
		} else {
			common.CountSafe("ShortIDUnknown")
			println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "blocktxn TX (short) ID unknown")
			return
		}
	}

	//println(c.ConnID, "Received the rest of compact block version", c.Node.SendCmpctVer)

	//sta := time.Now()
	b2g.BchBlock.UpdateContent(col.Assemble())
	//sto := time.Now()
	er := common.BchBlockChain.PostCheckBlock(b2g.BchBlock)
	if er != nil {
		println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "Corrupt CmpctBlkB")
		//c.DoS("BadCmpctBlockB")
		ioutil.WriteFile(b2g.Hash.String()+".bin", b2g.BchBlock.Raw, 0700)

		if b2g.BchBlock.MerkleRootMatch() {
			println("It was a wrongly mined one - clean it up")
			DelB2G(idx) //remove it from BlocksToGet
			if b2g.BchBlockTreeNode == LastCommitedHeader {
				LastCommitedHeader = LastCommitedHeader.Parent
			}
			common.BchBlockChain.DeleteBranch(b2g.BchBlockTreeNode, delB2G_callback)
		}

		return
	}
	DelB2G(idx)
	//fmt.Println(c.ConnID, "PostCheckBlock OK #", b2g.BchBlock.Height, sto.Sub(sta), time.Now().Sub(sta))
	c.Mutex.Lock()
	c.counters["NewTBlock"]++
	c.blocksreceived = append(c.blocksreceived, time.Now())
	c.Mutex.Unlock()
	orb := &OneReceivedBlock{TmStart: b2g.Started, TmPreproc: b2g.TmPreproc,
		TmDownload: c.LastMsgTime, TxMissing: col.Missing, FromConID: c.ConnID, DoInvs: b2g.SendInvs}
	ReceivedBlocks[idx] = orb
	if c.X.Authorized {
		b2g.BchBlock.Trusted = true
	}
	NetBlocks <- &BlockRcvd{Conn: c, BchBlock: b2g.BchBlock, BchBlockTreeNode: b2g.BchBlockTreeNode, OneReceivedBlock: orb}
}
