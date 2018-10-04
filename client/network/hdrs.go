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

// File:		hdrs.go
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
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
)

const (
	PH_STATUS_NEW   = 1
	PH_STATUS_FRESH = 2
	PH_STATUS_OLD   = 3
	PH_STATUS_ERROR = 4
	PH_STATUS_FATAL = 5
)

func (c *OneConnection) ProcessNewHeader(hdr []byte) (int, *OneBlockToGet) {
	var ok bool
	var b2g *OneBlockToGet
	bl, _ := bch.NewBchBlock(hdr)

	c.Mutex.Lock()
	c.InvStore(MSG_BLOCK, bl.Hash.Hash[:])
	c.Mutex.Unlock()

	if _, ok = ReceivedBlocks[bl.Hash.BIdx()]; ok {
		common.CountSafe("HeaderOld")
		//fmt.Println("", i, bl.Hash.String(), "-already received")
		return PH_STATUS_OLD, nil
	}

	if b2g, ok = BchBlocksToGet[bl.Hash.BIdx()]; ok {
		common.CountSafe("HeaderFresh")
		//fmt.Println(c.PeerAddr.Ip(), "block", bl.Hash.String(), " not new but get it")
		return PH_STATUS_FRESH, b2g
	}

	common.CountSafe("HeaderNew")
	fmt.Println("", bl.Hash.String(), " - NEW!")

	common.BchBlockChain.BchBlockIndexAccess.Lock()
	defer common.BchBlockChain.BchBlockIndexAccess.Unlock()

	if er, dos, _ := common.BchBlockChain.PreCheckBlock(bl); er != nil {
		common.CountSafe("PreCheckBlockFail")
		//println("PreCheckBlock err", dos, er.Error())
		if dos {
			return PH_STATUS_FATAL, nil
		} else {
			return PH_STATUS_ERROR, nil
		}
	}

	node := common.BchBlockChain.AcceptHeader(bl)
	b2g = &OneBlockToGet{Started: c.LastMsgTime, BchBlock: bl, BchBlockTreeNode: node, InProgress: 0}
	AddB2G(b2g)
	LastCommitedHeader = node

	if common.LastTrustedBlockMatch(node.BchBlockHash) {
		common.SetUint32(&common.LastTrustedBlockHeight, node.Height)
		for node != nil {
			node.Trusted = true
			node = node.Parent
		}
	}
	b2g.BchBlock.Trusted = b2g.BchBlockTreeNode.Trusted

	return PH_STATUS_NEW, b2g
}

func (c *OneConnection) HandleHeaders(pl []byte) (new_headers_got int) {
	var highest_block_found uint32

	c.MutexSetBool(&c.X.GetHeadersInProgress, false)

	b := bytes.NewReader(pl)
	cnt, e := bch.ReadVLen(b)
	if e != nil {
		println("HandleHeaders:", e.Error(), c.PeerAddr.Ip())
		return
	}

	HeadersReceived.Add(1)

	if cnt > 0 {
		MutexRcv.Lock()
		defer MutexRcv.Unlock()

		for i := 0; i < int(cnt); i++ {
			var hdr [81]byte

			n, _ := b.Read(hdr[:])
			if n != 81 {
				println("HandleHeaders: pl too short", c.PeerAddr.Ip())
				c.DoS("HdrErr1")
				return
			}

			if hdr[80] != 0 {
				fmt.Println("Unexpected value of txn_count from", c.PeerAddr.Ip())
				c.DoS("HdrErr2")
				return
			}

			sta, b2g := c.ProcessNewHeader(hdr[:])
			if b2g == nil {
				if sta == PH_STATUS_FATAL {
					//println("c.DoS(BadHeader)")
					c.DoS("BadHeader")
					return
				} else if sta == PH_STATUS_ERROR {
					//println("c.Misbehave(BadHeader)")
					c.Misbehave("BadHeader", 50) // do it 20 times and you are banned
				}
			} else {
				if sta == PH_STATUS_NEW {
					if cnt == 1 {
						b2g.SendInvs = true
					}
					new_headers_got++
				}
				if b2g.BchBlock.Height > highest_block_found {
					highest_block_found = b2g.BchBlock.Height
				}
				if c.Node.Height < b2g.BchBlock.Height {
					c.Mutex.Lock()
					c.Node.Height = b2g.BchBlock.Height
					c.Mutex.Unlock()
				}
				c.MutexSetBool(&c.X.GetBlocksDataNow, true)
				if b2g.TmPreproc.IsZero() { // do not overwrite TmPreproc (in case of PH_STATUS_FRESH)
					b2g.TmPreproc = time.Now()
				}
			}
		}
	} else {
		common.CountSafe("EmptyHeadersRcvd")
		HeadersReceived.Add(4)
	}

	c.Mutex.Lock()
	c.X.LastHeadersEmpty = highest_block_found <= c.X.LastHeadersHeightAsk
	c.X.TotalNewHeadersCount += new_headers_got
	if new_headers_got == 0 {
		c.X.AllHeadersReceived = true
	}
	c.Mutex.Unlock()

	return
}

func (c *OneConnection) ReceiveHeadersNow() {
	c.Mutex.Lock()
	c.X.AllHeadersReceived = false
	c.Mutex.Unlock()
}

// Handle getheaders protocol command
// https://en.bitcoin.it/wiki/Protocol_specification#getheaders
func (c *OneConnection) GetHeaders(pl []byte) {
	h2get, hashstop, e := parseLocatorsPayload(pl)
	if e != nil || hashstop == nil {
		println("GetHeaders: error parsing payload from", c.PeerAddr.Ip())
		c.DoS("BadGetHdrs")
		return
	}

	var best_block, last_block *bch_chain.BchBlockTreeNode

	//common.Last.Mutex.Lock()
	MutexRcv.Lock()
	last_block = LastCommitedHeader
	MutexRcv.Unlock()
	//common.Last.Mutex.Unlock()

	common.BchBlockChain.BchBlockIndexAccess.Lock()

	//println("GetHeaders", len(h2get), hashstop.String())
	if len(h2get) > 0 {
		for i := range h2get {
			if bl, ok := common.BchBlockChain.BchBlockIndex[h2get[i].BIdx()]; ok {
				if best_block == nil || bl.Height > best_block.Height {
					//println(" ... bbl", i, bl.Height, bl.BchBlockHash.String())
					best_block = bl
				}
			}
		}
	} else {
		best_block = common.BchBlockChain.BchBlockIndex[hashstop.BIdx()]
	}

	if best_block == nil {
		common.CountSafe("GetHeadersBadBlock")
		best_block = common.BchBlockChain.BchBlockTreeRoot
	}

	var resp []byte
	var cnt uint32

	defer func() {
		// If we get a hash of an old orphaned blocks, FindPathTo() will panic, so...
		if r := recover(); r != nil {
			common.CountSafe("GetHeadersOrphBlk")
		}

		common.BchBlockChain.BchBlockIndexAccess.Unlock()

		// send the response
		out := new(bytes.Buffer)
		bch.WriteVlen(out, uint64(cnt))
		out.Write(resp)
		c.SendRawMsg("headers", out.Bytes())
	}()

	for cnt < 2000 {
		if last_block.Height <= best_block.Height {
			break
		}
		best_block = best_block.FindPathTo(last_block)
		if best_block == nil {
			break
		}
		resp = append(resp, append(best_block.BchBlockHeader[:], 0)...) // 81st byte is always zero
		cnt++
	}

	// Note: the deferred function will be called before exiting

	return
}

func (c *OneConnection) sendGetHeaders() {
	MutexRcv.Lock()
	lb := LastCommitedHeader
	MutexRcv.Unlock()
	min_height := int(lb.Height) - bch_chain.MovingCheckopintDepth
	if min_height < 0 {
		min_height = 0
	}

	blks := new(bytes.Buffer)
	var cnt uint64
	var step int
	step = 1
	for cnt < 50 /*it should never get that far, but just in case...*/ {
		blks.Write(lb.BchBlockHash.Hash[:])
		cnt++
		//println(" geth", cnt, "height", lb.Height, lb.BchBlockHash.String())
		if int(lb.Height) <= min_height {
			break
		}
		for tmp := 0; tmp < step && lb != nil && int(lb.Height) > min_height; tmp++ {
			lb = lb.Parent
		}
		if lb == nil {
			break
		}
		if cnt >= 10 {
			step = step * 2
		}
	}
	var null_stop [32]byte
	blks.Write(null_stop[:])

	bhdr := new(bytes.Buffer)
	binary.Write(bhdr, binary.LittleEndian, common.Version)
	bch.WriteVlen(bhdr, cnt)

	c.SendRawMsg("getheaders", append(bhdr.Bytes(), blks.Bytes()...))
	c.X.LastHeadersHeightAsk = lb.Height
	c.MutexSetBool(&c.X.GetHeadersInProgress, true)
	c.X.GetHeadersTimeout = time.Now().Add(GetHeadersTimeout)
}
