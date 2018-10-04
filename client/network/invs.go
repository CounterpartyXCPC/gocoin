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

// File:		invs.go
// Description:	Bictoin Cash network Package

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

package network

import (
	"fmt"
	//"time"
	"bytes"
	"encoding/binary"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
)

const (
	MSG_WITNESS_FLAG = 0x40000000

	MSG_TX            = 1
	MSG_BLOCK         = 2
	MSG_CMPCT_BLOCK   = 4
	MSG_WITNESS_TX    = MSG_TX | MSG_WITNESS_FLAG
	MSG_WITNESS_BLOCK = MSG_BLOCK | MSG_WITNESS_FLAG
)

func blockReceived(bh *bch.Uint256) (ok bool) {
	MutexRcv.Lock()
	_, ok = ReceivedBlocks[bh.BIdx()]
	MutexRcv.Unlock()
	return
}

func hash2invid(hash []byte) uint64 {
	return binary.LittleEndian.Uint64(hash[4:12])
}

// Make sure c.Mutex is locked when calling it
func (c *OneConnection) InvStore(typ uint32, hash []byte) {
	inv_id := hash2invid(hash)
	if len(c.InvDone.History) < MAX_INV_HISTORY {
		c.InvDone.History = append(c.InvDone.History, inv_id)
		c.InvDone.Map[inv_id] = typ
		c.InvDone.Idx++
		return
	}
	if c.InvDone.Idx == MAX_INV_HISTORY {
		c.InvDone.Idx = 0
	}
	delete(c.InvDone.Map, c.InvDone.History[c.InvDone.Idx])
	c.InvDone.History[c.InvDone.Idx] = inv_id
	c.InvDone.Map[inv_id] = typ
	c.InvDone.Idx++
}

func (c *OneConnection) ProcessInv(pl []byte) {
	if len(pl) < 37 {
		//println(c.PeerAddr.Ip(), "inv payload too short", len(pl))
		c.DoS("InvEmpty")
		return
	}
	c.Mutex.Lock()
	c.X.InvsRecieved++
	c.Mutex.Unlock()

	cnt, of := bch.VLen(pl)
	if len(pl) != of+36*cnt {
		println("inv payload length mismatch", len(pl), of, cnt)
	}

	for i := 0; i < cnt; i++ {
		typ := binary.LittleEndian.Uint32(pl[of : of+4])
		c.Mutex.Lock()
		c.InvStore(typ, pl[of+4:of+36])
		ahr := c.X.AllHeadersReceived
		c.Mutex.Unlock()
		common.CountSafe(fmt.Sprint("InvGot-", typ))
		if typ == MSG_BLOCK {
			bhash := bch.NewUint256(pl[of+4 : of+36])
			if !ahr {
				common.CountSafe("InvBlockIgnored")
			} else {
				if !blockReceived(bhash) {
					MutexRcv.Lock()
					if b2g, ok := BchBlocksToGet[bhash.BIdx()]; ok {
						if c.Node.Height < b2g.BchBlock.Height {
							c.Node.Height = b2g.BchBlock.Height
						}
						common.CountSafe("InvBlockFresh")
						println(c.PeerAddr.Ip(), c.Node.Version, "also knows the block", b2g.BchBlock.Height, bhash.String())
						c.MutexSetBool(&c.X.GetBlocksDataNow, true)
					} else {
						common.CountSafe("InvBlockNew")
						c.ReceiveHeadersNow()
						println(c.PeerAddr.Ip(), c.Node.Version, "possibly new block", bhash.String())
					}
					MutexRcv.Unlock()
				} else {
					common.CountSafe("InvBlockOld")
				}
			}
		} else if typ == MSG_TX {
			if common.AcceptTx() {
				c.TxInvNotify(pl[of+4 : of+36])
			} else {
				common.CountSafe("InvTxIgnored")
			}
		}
		of += 36
	}

	return
}

func NetRouteInv(typ uint32, h *bch.Uint256, fromConn *OneConnection) uint32 {
	var fee_spkb uint64
	if typ == MSG_TX {
		TxMutex.Lock()
		if tx, ok := TransactionsToSend[h.BIdx()]; ok {
			fee_spkb = (1000 * tx.Fee) / uint64(tx.VSize())
		} else {
			println("NetRouteInv: txid", h.String(), "not in mempool")
		}
		TxMutex.Unlock()
	}
	return NetRouteInvExt(typ, h, fromConn, fee_spkb)
}

// This function is called from the main thread (or from an UI)
func NetRouteInvExt(typ uint32, h *bch.Uint256, fromConn *OneConnection, fee_spkb uint64) (cnt uint32) {
	common.CountSafe(fmt.Sprint("NetRouteInv", typ))

	// Prepare the inv
	inv := new([36]byte)
	binary.LittleEndian.PutUint32(inv[0:4], typ)
	copy(inv[4:36], h.Bytes())

	// Append it to PendingInvs in each open connection
	Mutex_net.Lock()
	for _, v := range OpenCons {
		if v != fromConn { // except the one that this inv came from
			send_inv := true
			v.Mutex.Lock()
			if typ == MSG_TX {
				if v.Node.DoNotRelayTxs {
					send_inv = false
					common.CountSafe("SendInvNoTxNode")
				} else if v.X.MinFeeSPKB > 0 && uint64(v.X.MinFeeSPKB) > fee_spkb {
					send_inv = false
					common.CountSafe("SendInvFeeTooLow")
				}

				/* This is to prevent sending own txs to "spying" peers:
				else if fromConn==nil && v.X.InvsRecieved==0 {
					send_inv = false
					common.CountSafe("SendInvOwnBlocked")
				}
				*/
			}
			if send_inv {
				if len(v.PendingInvs) < 500 {
					if typ, ok := v.InvDone.Map[hash2invid(inv[4:36])]; ok {
						common.CountSafe(fmt.Sprint("SendInvSame-", typ))
					} else {
						v.PendingInvs = append(v.PendingInvs, inv)
						cnt++
					}
				} else {
					common.CountSafe("SendInvFull")
				}
			}
			v.Mutex.Unlock()
		}
	}
	Mutex_net.Unlock()
	return
}

// Call this function only when BlockIndexAccess is locked
func addInvBlockBranch(inv map[[32]byte]bool, bl *bch_chain.BchBlockTreeNode, stop *bch.Uint256) {
	if len(inv) >= 500 || bl.BchBlockHash.Equal(stop) {
		return
	}
	inv[bl.BchBlockHash.Hash] = true
	for i := range bl.Childs {
		if len(inv) >= 500 {
			return
		}
		addInvBlockBranch(inv, bl.Childs[i], stop)
	}
}

func (c *OneConnection) GetBlocks(pl []byte) {
	h2get, hashstop, e := parseLocatorsPayload(pl)

	if e != nil || len(h2get) < 1 || hashstop == nil {
		println("GetBlocks: error parsing payload from", c.PeerAddr.Ip())
		c.DoS("BadGetBlks")
		return
	}

	invs := make(map[[32]byte]bool, 500)
	for i := range h2get {
		common.BchBlockChain.BchBlockIndexAccess.Lock()
		if bl, ok := common.BchBlockChain.BchBlockIndex[h2get[i].BIdx()]; ok {
			// make sure that this block is in our main chain
			common.Last.Mutex.Lock()
			end := common.Last.BchBlock
			common.Last.Mutex.Unlock()
			for ; end != nil && end.Height >= bl.Height; end = end.Parent {
				if end == bl {
					addInvBlockBranch(invs, bl, hashstop) // Yes - this is the main chain
					if len(invs) > 0 {
						common.BchBlockChain.BchBlockIndexAccess.Unlock()

						inv := new(bytes.Buffer)
						bch.WriteVlen(inv, uint64(len(invs)))
						for k := range invs {
							binary.Write(inv, binary.LittleEndian, uint32(2))
							inv.Write(k[:])
						}
						c.SendRawMsg("inv", inv.Bytes())
						return
					}
				}
			}
		}
		common.BchBlockChain.BchBlockIndexAccess.Unlock()
	}

	common.CountSafe("GetblksMissed")
	return
}

func (c *OneConnection) SendInvs() (res bool) {
	b_txs := new(bytes.Buffer)
	b_blk := new(bytes.Buffer)
	var c_blk []*bch.Uint256

	c.Mutex.Lock()
	if len(c.PendingInvs) > 0 {
		for i := range c.PendingInvs {
			var inv_sent_otherwise bool
			typ := binary.LittleEndian.Uint32((*c.PendingInvs[i])[:4])
			c.InvStore(typ, (*c.PendingInvs[i])[4:36])
			if typ == MSG_BLOCK {
				if c.Node.SendCmpctVer >= 1 && c.Node.HighBandwidth {
					c_blk = append(c_blk, bch.NewUint256((*c.PendingInvs[i])[4:]))
					inv_sent_otherwise = true
				} else if c.Node.SendHeaders {
					// convert block inv to block header
					common.BchBlockChain.BchBlockIndexAccess.Lock()
					bl := common.BchBlockChain.BchBlockIndex[bch.NewUint256((*c.PendingInvs[i])[4:]).BIdx()]
					if bl != nil {
						b_blk.Write(bl.BchBlockHeader[:])
						b_blk.Write([]byte{0}) // 0 txs
					}
					common.BchBlockChain.BchBlockIndexAccess.Unlock()
					inv_sent_otherwise = true
				}
			}

			if !inv_sent_otherwise {
				b_txs.Write((*c.PendingInvs[i])[:])
			}
		}
		res = true
	}
	c.PendingInvs = nil
	c.Mutex.Unlock()

	if len(c_blk) > 0 {
		for _, h := range c_blk {
			c.SendCmpctBlk(h)
		}
	}

	if b_blk.Len() > 0 {
		common.CountSafe("InvSentAsHeader")
		b := new(bytes.Buffer)
		bch.WriteVlen(b, uint64(b_blk.Len()/81))
		c.SendRawMsg("headers", append(b.Bytes(), b_blk.Bytes()...))
		//println("sent block's header(s)", b_blk.Len(), uint64(b_blk.Len()/81))
	}

	if b_txs.Len() > 0 {
		b := new(bytes.Buffer)
		bch.WriteVlen(b, uint64(b_txs.Len()/36))
		c.SendRawMsg("inv", append(b.Bytes(), b_txs.Bytes()...))
	}

	return
}
