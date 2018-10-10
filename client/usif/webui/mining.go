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

// File:		mining.go
// Description:	Bictoin Cash webui Package

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

package webui

import (
	"fmt"
	"sort"
	"time"

	//	"bytes"
	//	"regexp"
	"encoding/binary"
	"encoding/json"
	"net/http"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

type omv struct {
	unknown_miner bool
	cnt           int
	bts           uint64
	fees          uint64
	//ebad_cnt int
	//nya_cnt int
}

type onemiernstat []struct {
	name string
	omv
}

func (x onemiernstat) Len() int {
	return len(x)
}

func (x onemiernstat) Less(i, j int) bool {
	if x[i].cnt == x[j].cnt {
		return x[i].name < x[j].name // Same numbers: sort by name
	}
	return x[i].cnt > x[j].cnt
}

func (x onemiernstat) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func p_miners(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	write_html_head(w, r)
	w.Write([]byte(load_template("miners.html")))
	write_html_tail(w)
}

func json_blkver(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	w.Header()["Content-Type"] = []string{"application/json"}

	common.Last.Mutex.Lock()
	end := common.Last.BchBlock
	common.Last.Mutex.Unlock()

	w.Write([]byte("["))
	if end != nil {
		max_cnt := 2 * 2016 //common.BchBlockChain.Consensus.Window
		for {
			w.Write([]byte(fmt.Sprint("[", end.Height, ",", binary.LittleEndian.Uint32(end.BchBlockHeader[0:4]), "]")))
			end = end.Parent
			if end == nil || max_cnt <= 1 {
				break
			}
			max_cnt--
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]"))
}

func json_miners(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	type one_miner_row struct {
		Unknown               bool
		Name                  string
		BchBlocks             int
		TotalFees, TotalBytes uint64
		//BUcnt, NYAcnt int
	}

	type the_mining_stats struct {
		MiningStatHours  uint
		BchBlockCount    uint
		FirstBlockTime   int64
		AvgBlocksPerHour float64
		AvgDifficulty    float64
		AvgHashrate      float64
		NextDiffChange   uint32
		Miners           []one_miner_row
	}

	common.ReloadMiners()

	m := make(map[string]omv, 20)
	var om omv
	cnt := uint(0)
	common.Last.Mutex.Lock()
	end := common.Last.BchBlock
	common.Last.Mutex.Unlock()
	var lastts int64
	var diff float64
	now := time.Now().Unix()

	next_diff_change := 2016 - end.Height%2016

	//eb_ad_x := regexp.MustCompile("/EB[0-9]+/AD[0-9]+/")

	for ; end != nil; cnt++ {
		if now-int64(end.Timestamp()) > int64(common.CFG.Stat.MiningHrs)*3600 {
			break
		}
		lastts = int64(end.Timestamp())
		bl, _, e := common.BchBlockChain.BchBlocks.BchBlockGet(end.BchBlockHash)
		if e != nil {
			break
		}

		block, e := bch.NewBchBlock(bl)
		if e != nil {
			break
		}

		cbasetx, _ := bch.NewTx(bl[block.TxOffset:])

		diff += bch.GetDifficulty(end.Bits())
		miner, mid := common.TxMiner(cbasetx)
		om = m[miner]
		om.cnt++
		om.bts += uint64(len(bl))
		om.unknown_miner = (mid == -1)

		// Blocks reward
		var rew uint64
		for o := range cbasetx.TxOut {
			rew += cbasetx.TxOut[o].Value
		}
		fees := rew - bch.GetBlockReward(end.Height)
		if int64(fees) > 0 { // solution for a possibility of a miner not claiming the reward (see block #501726)
			om.fees += fees
		}

		/*if eb_ad_x.Find(cbasetx.TxIn[0].ScriptSig) != nil {
			om.ebad_cnt++
		}

		if bytes.Index(cbasetx.TxIn[0].ScriptSig, []byte("/NYA/")) != -1 {
			om.nya_cnt++
		}*/

		m[miner] = om

		end = end.Parent
	}

	if cnt == 0 {
		w.Write([]byte("{}"))
		return
	}

	srt := make(onemiernstat, len(m))
	i := 0
	for k, v := range m {
		srt[i].name = k
		srt[i].omv = v
		i++
	}
	sort.Sort(srt)

	var stats the_mining_stats

	diff /= float64(cnt)
	bph := float64(cnt) / float64(common.CFG.Stat.MiningHrs)
	hrate := bph / 6 * diff * 7158278.826667

	stats.MiningStatHours = common.CFG.Stat.MiningHrs
	stats.BchBlockCount = cnt
	stats.FirstBlockTime = lastts
	stats.AvgBlocksPerHour = bph
	stats.AvgDifficulty = diff
	stats.AvgHashrate = hrate
	stats.NextDiffChange = next_diff_change

	stats.Miners = make([]one_miner_row, len(srt))
	for i := range srt {
		stats.Miners[i].Unknown = srt[i].unknown_miner
		stats.Miners[i].Name = srt[i].name
		stats.Miners[i].BchBlocks = srt[i].cnt
		stats.Miners[i].TotalFees = srt[i].fees
		stats.Miners[i].TotalBytes = srt[i].bts
		//stats.Miners[i].BUcnt = srt[i].ebad_cnt
		//stats.Miners[i].NYAcnt = srt[i].nya_cnt
	}

	bx, er := json.Marshal(stats)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}

}
