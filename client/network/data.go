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

// File:		data.go
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
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

func (c *OneConnection) ProcessGetData(pl []byte) {
	//var notfound []byte

	// println(c.PeerAddr.Ip(), "getdata")
	b := bytes.NewReader(pl)
	cnt, e := bch.ReadVLen(b)
	if e != nil {
		println("ProcessGetData:", e.Error(), c.PeerAddr.Ip())
		return
	}
	for i := 0; i < int(cnt); i++ {
		var typ uint32
		var h [36]byte

		n, _ := b.Read(h[:])
		if n != 36 {
			println("ProcessGetData: pl too short", c.PeerAddr.Ip())
			return
		}

		typ = binary.LittleEndian.Uint32(h[:4])
		c.Mutex.Lock()
		c.InvStore(typ, h[4:36])
		c.Mutex.Unlock()

		common.CountSafe(fmt.Sprintf("GetdataType-%x", typ))
		if typ == MSG_BLOCK || typ == MSG_WITNESS_BLOCK {
			hash := bch.NewUint256(h[4:])
			crec, _, er := common.BchBlockChain.BchBlocks.BchBlockGetExt(hash)

			if er == nil {
				bl := crec.Data
				if typ == MSG_BLOCK {
					// remove witness data from the block
					if crec.BchBlock == nil {
						crec.BchBlock, _ = bch.NewBchBlock(bl)
					}
					if crec.BchBlock.NoWitnessData == nil {
						crec.BchBlock.BuildNoWitnessData()
					}
					println("block size", len(crec.Data), "->", len(bl))
					bl = crec.BchBlock.NoWitnessData
				}
				c.SendRawMsg("block", bl)
			} else {
				//fmt.Println("BlockGetExt-2 failed for", hash.String(), er.Error())
				//notfound = append(notfound, h[:]...)
			}
		} else if typ == MSG_TX || typ == MSG_WITNESS_TX {
			// transaction
			TxMutex.Lock()
			if tx, ok := TransactionsToSend[bch.NewUint256(h[4:]).BIdx()]; ok && tx.BchBlocked == 0 {
				tx.SentCnt++
				tx.Lastsent = time.Now()
				TxMutex.Unlock()
				if tx.SegWit == nil || typ == MSG_WITNESS_TX {
					c.SendRawMsg("tx", tx.Raw)
				} else {
					c.SendRawMsg("tx", tx.Serialize())
				}
			} else {
				TxMutex.Unlock()
				//notfound = append(notfound, h[:]...)
			}
		} else if typ == MSG_CMPCT_BLOCK {
			if !c.SendCmpctBlk(bch.NewUint256(h[4:])) {
				println(c.ConnID, c.PeerAddr.Ip(), c.Node.Agent, "asked for CmpctBlk we don't have", bch.NewUint256(h[4:]).String())
				if c.Misbehave("GetCmpctBlk", 100) {
					break
				}
			}
		} else {
			if typ > 0 && typ <= 3 /*3 is a filtered block(we dont support it)*/ {
				//notfound = append(notfound, h[:]...)
			}
		}
	}

	/*
		if len(notfound)>0 {
			buf := new(bytes.Buffer)
			bch.WriteVlen(buf, uint64(len(notfound)/36))
			buf.Write(notfound)
			c.SendRawMsg("notfound", buf.Bytes())
		}
	*/
}

// This function is called from a net conn thread
func netBlockReceived(conn *OneConnection, b []byte) {
	println("netBlockReceived")
	if len(b) < 100 {
		conn.DoS("ShortBlock")
		return
	}

	hash := bch.NewSha2Hash(b[:80])
	idx := hash.BIdx()
	println("got block data", hash.String())

	MutexRcv.Lock()

	// the blocks seems to be fine
	if rb, got := ReceivedBlocks[idx]; got {
		rb.Cnt++
		common.CountSafe("BlockSameRcvd")
		conn.Mutex.Lock()
		delete(conn.GetBlockInProgress, idx)
		conn.Mutex.Unlock()
		MutexRcv.Unlock()
		return
	}

	// remove from BlocksToGet:
	b2g := BchBlocksToGet[idx]
	if b2g == nil {
		//println("Block", hash.String(), " from", conn.PeerAddr.Ip(), conn.Node.Agent, " was not expected")

		var hdr [81]byte
		var sta int
		copy(hdr[:80], b[:80])
		sta, b2g = conn.ProcessNewHeader(hdr[:])
		if b2g == nil {
			if sta == PH_STATUS_FATAL {
				println("Unrequested Block: FAIL - Ban", conn.PeerAddr.Ip(), conn.Node.Agent)
				conn.DoS("BadUnreqBlock")
			} else {
				common.CountSafe("ErrUnreqBlock")
			}
			//conn.Disconnect()
			MutexRcv.Unlock()
			return
		}
		if sta == PH_STATUS_NEW {
			b2g.SendInvs = true
		}
		//println(c.ConnID, " - taking this new block")
		common.CountSafe("UnxpectedBlockNEW")
	}

	println("Debugging block", b2g.BchBlockTreeNode.Height, " len", len(b), " got from", conn.PeerAddr.Ip(), b2g.InProgress)

	b2g.BchBlock.Raw = b
	if conn.X.Authorized {
		b2g.BchBlock.Trusted = true
	}

	er := common.BchBlockChain.PostCheckBlock(b2g.BchBlock)
	if er != nil {
		b2g.InProgress--
		println("Corrupt block received from", conn.PeerAddr.Ip(), er.Error())
		//ioutil.WriteFile(hash.String() + ".bin", b, 0700)
		conn.DoS("BadBlock")

		// we don't need to remove from conn.GetBlockInProgress as we're disconnecting

		if b2g.BchBlock.MerkleRootMatch() {
			println("It was a wrongly mined one - clean it up")
			DelB2G(idx) //remove it from BlocksToGet
			if b2g.BchBlockTreeNode == LastCommitedHeader {
				LastCommitedHeader = LastCommitedHeader.Parent
			}
			common.BchBlockChain.DeleteBranch(b2g.BchBlockTreeNode, delB2G_callback)
		}

		MutexRcv.Unlock()
		return
	}

	orb := &OneReceivedBlock{TmStart: b2g.Started, TmPreproc: b2g.TmPreproc,
		TmDownload: conn.LastMsgTime, FromConID: conn.ConnID, DoInvs: b2g.SendInvs}

	conn.Mutex.Lock()
	bip := conn.GetBlockInProgress[idx]
	if bip == nil {
		//println(conn.ConnID, "received unrequested block", hash.String())
		common.CountSafe("UnreqBlockRcvd")
		conn.counters["NewBlock!"]++
		orb.TxMissing = -2
	} else {
		delete(conn.GetBlockInProgress, idx)
		conn.counters["NewBlock"]++
		orb.TxMissing = -1
	}
	conn.blocksreceived = append(conn.blocksreceived, time.Now())
	conn.Mutex.Unlock()

	ReceivedBlocks[idx] = orb
	DelB2G(idx) //remove it from BchBlocksToGet if no more pending downloads

	store_on_disk := len(BchBlocksToGet) > 10 && common.GetBool(&common.CFG.Memory.CacheOnDisk) && len(b2g.BchBlock.Raw) > 16*1024
	MutexRcv.Unlock()

	var bei *bch.BchBlockExtraInfo

	if store_on_disk {
		if e := ioutil.WriteFile(common.TempBlocksDir()+hash.String(), b2g.BchBlock.Raw, 0600); e == nil {
			bei = new(bch.BchBlockExtraInfo)
			*bei = b2g.BchBlock.BchBlockExtraInfo
			b2g.BchBlock = nil
		} else {
			println("write tmp block:", e.Error())
		}
	}

	NetBlocks <- &BchBlockRcvd{Conn: conn, BchBlock: b2g.BchBlock, BchBlockTreeNode: b2g.BchBlockTreeNode, OneReceivedBlock: orb, BchBlockExtraInfo: bei}
}

// Read VLen followed by the number of locators
// parse the payload of getblocks and getheaders messages
func parseLocatorsPayload(pl []byte) (h2get []*bch.Uint256, hashstop *bch.Uint256, er error) {
	var cnt uint64
	var h [32]byte
	var ver uint32

	b := bytes.NewReader(pl)

	// version
	if er = binary.Read(b, binary.LittleEndian, &ver); er != nil {
		return
	}

	// hash count
	cnt, er = bch.ReadVLen(b)
	if er != nil {
		return
	}

	// block locator hashes
	if cnt > 0 {
		h2get = make([]*bch.Uint256, cnt)
		for i := 0; i < int(cnt); i++ {
			if _, er = b.Read(h[:]); er != nil {
				return
			}
			h2get[i] = bch.NewUint256(h[:])
		}
	}

	// hash_stop
	if _, er = b.Read(h[:]); er != nil {
		return
	}
	hashstop = bch.NewUint256(h[:])

	return
}

// Call it with locked MutexRcv
func getBlockToFetch(max_height uint32, cnt_in_progress, avg_block_size uint) (lowest_found *OneBlockToGet) {
	for _, v := range BchBlocksToGet {
		if v.InProgress == cnt_in_progress && v.BchBlock.Height <= max_height &&
			(lowest_found == nil || v.BchBlock.Height < lowest_found.BchBlock.Height) {
			lowest_found = v
		}
	}
	return
}

func (c *OneConnection) GetBlockData() (yes bool) {
	//MAX_GETDATA_FORWARD
	// Need to send getdata...?
	MutexRcv.Lock()
	defer MutexRcv.Unlock()

	if LowestIndexToBlocksToGet == 0 || len(BchBlocksToGet) == 0 {
		c.IncCnt("FetchNoBlocksToGet", 1)
		// wake up in one minute, just in case
		c.nextGetData = time.Now().Add(60 * time.Second)
		return
	}

	c.Mutex.Lock()
	if c.X.BchBlocksExpired > 0 { // Do not fetch blocks from nodes that had not given us some in the past
		c.Mutex.Unlock()
		c.IncCnt("FetchHasBlocksExpired", 1)
		return
	}
	cbip := len(c.GetBlockInProgress)
	c.Mutex.Unlock()

	if cbip >= MAX_PEERS_BLOCKS_IN_PROGRESS {
		c.IncCnt("FetchMaxCountInProgress", 1)
		// wake up in a few seconds, maybe some blocks will complete by then
		c.nextGetData = time.Now().Add(1 * time.Second)
		return
	}

	avg_block_size := common.AverageBlockSize.Get()
	block_data_in_progress := cbip * avg_block_size

	if block_data_in_progress > 0 && (block_data_in_progress+avg_block_size) > MAX_GETDATA_FORWARD {
		c.IncCnt("FetchMaxBytesInProgress", 1)
		// wake up in a few seconds, maybe some blocks will complete by then
		c.nextGetData = time.Now().Add(1 * time.Second) // wait for some blocks to complete
		return
	}

	var cnt uint64
	var block_type uint32

	if (c.Node.Services & SERVICE_SEGWIT) != 0 {
		block_type = MSG_WITNESS_BLOCK
	} else {
		block_type = MSG_BLOCK
	}

	// We can issue getdata for this peer
	// Let's look for the lowest height block in BchBlocksToGet that isn't being downloaded yet

	common.Last.Mutex.Lock()
	max_height := common.Last.BchBlock.Height + uint32(MAX_BLOCKS_FORWARD_SIZ/avg_block_size)
	if max_height > common.Last.BchBlock.Height+MAX_BLOCKS_FORWARD_CNT {
		max_height = common.Last.BchBlock.Height + MAX_BLOCKS_FORWARD_CNT
	}
	common.Last.Mutex.Unlock()
	if max_height > c.Node.Height {
		max_height = c.Node.Height
	}
	if max_height > LastCommitedHeader.Height {
		max_height = LastCommitedHeader.Height
	}

	if common.BchBlockChain.Consensus.Enforce_SEGWIT != 0 && (c.Node.Services&SERVICE_SEGWIT) == 0 { // no segwit node
		if max_height >= common.BchBlockChain.Consensus.Enforce_SEGWIT-1 {
			max_height = common.BchBlockChain.Consensus.Enforce_SEGWIT - 1
			if max_height <= common.Last.BchBlock.Height {
				c.IncCnt("FetchNoWitness", 1)
				c.nextGetData = time.Now().Add(3600 * time.Second) // never do getdata
				return
			}
		}
	}

	invs := new(bytes.Buffer)
	var cnt_in_progress uint

	for {
		var lowest_found *OneBlockToGet

		// Get block to fetch:

		for bh := LowestIndexToBlocksToGet; bh <= max_height; bh++ {
			if idxlst, ok := IndexToBlocksToGet[bh]; ok {
				for _, idx := range idxlst {
					v := BchBlocksToGet[idx]
					if v.InProgress == cnt_in_progress && (lowest_found == nil || v.BchBlock.Height < lowest_found.BchBlock.Height) {
						c.Mutex.Lock()
						if _, ok := c.GetBlockInProgress[idx]; !ok {
							lowest_found = v
						}
						c.Mutex.Unlock()
					}
				}
			}
		}

		if lowest_found == nil {
			cnt_in_progress++
			if cnt_in_progress >= uint(common.CFG.Net.MaxBlockAtOnce) {
				break
			}
			continue
		}

		binary.Write(invs, binary.LittleEndian, block_type)
		invs.Write(lowest_found.BchBlockHash.Hash[:])
		lowest_found.InProgress++
		cnt++

		c.Mutex.Lock()
		c.GetBlockInProgress[lowest_found.BchBlockHash.BIdx()] =
			&oneBlockDl{hash: lowest_found.BchBlockHash, start: time.Now(), SentAtPingCnt: c.X.PingSentCnt}
		cbip = len(c.GetBlockInProgress)
		c.Mutex.Unlock()

		if cbip >= MAX_PEERS_BLOCKS_IN_PROGRESS {
			break // no more than 2000 blocks in progress / peer
		}
		block_data_in_progress += avg_block_size
		if block_data_in_progress > MAX_GETDATA_FORWARD {
			break
		}
	}

	if cnt == 0 {
		//println(c.ConnID, "fetch nothing", cbip, block_data_in_progress, max_height-common.Last.BchBlock.Height, cnt_in_progress)
		c.IncCnt("FetchNothing", 1)
		// wake up in a few seconds, maybe it will be different next time
		c.nextGetData = time.Now().Add(5 * time.Second)
		return
	}

	bu := new(bytes.Buffer)
	bch.WriteVlen(bu, uint64(cnt))
	pl := append(bu.Bytes(), invs.Bytes()...)
	//println(c.ConnID, "fetching", cnt, "new blocks ->", cbip)
	c.SendRawMsg("getdata", pl)
	yes = true

	return
}
