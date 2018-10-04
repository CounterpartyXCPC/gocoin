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

// File:		blocks.go
// Description:	Bictoin Cash webui Package

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

package webui

import (
	"encoding/json"
	"net/http"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	"github.com/counterpartyxcpc/gocoin-cash/client/usif"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"

	//	"regexp"
	"strconv"
	"time"
)

func p_blocks(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	write_html_head(w, r)
	w.Write([]byte(load_template("blocks.html")))
	write_html_tail(w)
}

func json_blocks(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	type one_block struct {
		Height    uint32
		Timestamp uint32
		Hash      string
		TxCnt     int
		Size      int
		Weight    uint
		Version   uint32
		Reward    uint64
		Miner     string
		FeeSPB    float64

		Received                          uint32
		TimePre, TimeDl, TimeVer, TimeQue int
		WasteCnt                          uint
		MissedCnt                         int
		FromConID                         uint32
		Sigops                            int

		NonWitnessSize int
		//EBAD           string

		HaveFeeStats bool
	}

	var blks []*one_block

	common.Last.Mutex.Lock()
	end := common.Last.BchBlock
	common.Last.Mutex.Unlock()

	//eb_ad_x := regexp.MustCompile("/EB[0-9]+/AD[0-9]+/")

	for cnt := uint32(0); end != nil && cnt < common.GetUint32(&common.CFG.WebUI.ShowBlocks); cnt++ {
		bl, _, e := common.BchBlockChain.BchBlocks.BchBlockGet(end.BchBlockHash)
		if e != nil {
			break
		}
		block, e := bch.NewBchBlock(bl)
		if e != nil {
			break
		}

		b := new(one_block)
		b.Height = end.Height
		b.Timestamp = block.BchBlockTime()
		b.Hash = end.BchBlockHash.String()
		b.TxCnt = block.TxCount
		b.Size = len(bl)
		b.Weight = block.BchBlockWeight
		b.Version = block.Version()

		cbasetx, cbaselen := bch.NewTx(bl[block.TxOffset:])
		for o := range cbasetx.TxOut {
			b.Reward += cbasetx.TxOut[o].Value
		}

		b.Miner, _ = common.TxMiner(cbasetx)
		if len(bl)-block.TxOffset-cbaselen != 0 {
			b.FeeSPB = float64(b.Reward-bch.GetBlockReward(end.Height)) / float64(len(bl)-block.TxOffset-cbaselen)
		}

		common.BchBlockChain.BchBlockIndexAccess.Lock()
		node := common.BchBlockChain.BchBlockIndex[end.BchBlockHash.BIdx()]
		common.BchBlockChain.BchBlockIndexAccess.Unlock()

		network.MutexRcv.Lock()
		rb := network.ReceivedBlocks[end.BchBlockHash.BIdx()]
		network.MutexRcv.Unlock()

		b.Received = uint32(rb.TmStart.Unix())
		b.Sigops = int(node.SigopsCost)

		if rb.TmPreproc.IsZero() {
			b.TimePre = -1
		} else {
			b.TimePre = int(rb.TmPreproc.Sub(rb.TmStart) / time.Millisecond)
		}

		if rb.TmDownload.IsZero() {
			b.TimeDl = -1
		} else {
			b.TimeDl = int(rb.TmDownload.Sub(rb.TmStart) / time.Millisecond)
		}

		if rb.TmQueue.IsZero() {
			b.TimeQue = -1
		} else {
			b.TimeQue = int(rb.TmQueue.Sub(rb.TmStart) / time.Millisecond)
		}

		if rb.TmAccepted.IsZero() {
			b.TimeVer = -1
		} else {
			b.TimeVer = int(rb.TmAccepted.Sub(rb.TmStart) / time.Millisecond)
		}

		b.WasteCnt = rb.Cnt
		b.MissedCnt = rb.TxMissing
		b.FromConID = rb.FromConID

		b.NonWitnessSize = rb.NonWitnessSize

		/*if res := eb_ad_x.Find(cbasetx.TxIn[0].ScriptSig); res != nil {
			b.EBAD = string(res)
		}*/

		usif.BchBlockFeesMutex.Lock()
		_, b.HaveFeeStats = usif.BchBlockFees[end.Height]
		usif.BchBlockFeesMutex.Unlock()

		blks = append(blks, b)
		end = end.Parent
	}

	bx, er := json.Marshal(blks)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}

}

func json_blfees(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	if len(r.Form["height"]) == 0 {
		w.Write([]byte("No hash given"))
		return
	}

	height, e := strconv.ParseUint(r.Form["height"][0], 10, 32)
	if e != nil {
		w.Write([]byte(e.Error()))
		return
	}

	usif.BchBlockFeesMutex.Lock()
	fees, ok := usif.BchBlockFees[uint32(height)]
	usif.BchBlockFeesMutex.Unlock()

	if !ok {
		w.Write([]byte("File not found"))
		return
	}

	bx, er := json.Marshal(fees)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}
}
