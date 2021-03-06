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
// Description:	Bictoin Cash textui Package

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

package textui

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

func do_mining(s string) {
	var totbtc, hrs, segwit_cnt uint64
	if s != "" {
		hrs, _ = strconv.ParseUint(s, 10, 64)
	}
	if hrs == 0 {
		hrs = uint64(common.CFG.Stat.MiningHrs)
	}
	fmt.Println("Looking back", hrs, "hours...")
	lim := uint32(time.Now().Add(-time.Hour * time.Duration(hrs)).Unix())
	common.Last.Mutex.Lock()
	bte := common.Last.BchBlock
	end := bte
	common.Last.Mutex.Unlock()
	cnt, diff := 0, float64(0)
	tot_blocks, tot_blocks_len := 0, 0

	bip100_voting := make(map[string]uint)
	bip100x := regexp.MustCompile("/BV{0,1}[0-9]+[M]{0,1}/")

	eb_ad_voting := make(map[string]uint)
	eb_ad_x := regexp.MustCompile("/EB[0-9]+/AD[0-9]+/")

	for end.Timestamp() >= lim {
		bl, _, e := common.BchBlockChain.BchBlocks.BchBlockGet(end.BchBlockHash)
		if e != nil {
			println(cnt, e.Error())
			return
		}
		block, e := bch.NewBchBlock(bl)
		if e != nil {
			println("bch.NewBchBlock failed", e.Error())
			return
		}

		bt, _ := bch.NewBchBlock(bl)
		cbasetx, _ := bch.NewTx(bl[bt.TxOffset:])

		tot_blocks++
		tot_blocks_len += len(bl)
		diff += bch.GetDifficulty(block.Bits())

		if (block.Version() & 0x20000002) == 0x20000002 {
			segwit_cnt++
		}

		res := bip100x.Find(cbasetx.TxIn[0].ScriptSig)
		if res != nil {
			bip100_voting[string(res)]++
			nimer, _ := common.TxMiner(cbasetx)
			fmt.Println("      block", end.Height, "by", nimer, "BIP100 voting", string(res), " total:", bip100_voting[string(res)])
		}

		res = eb_ad_x.Find(cbasetx.TxIn[0].ScriptSig)
		if res != nil {
			eb_ad_voting[string(res)]++
		}

		end = end.Parent
	}
	if tot_blocks == 0 {
		fmt.Println("There are no blocks from the last", hrs, "hour(s)")
		return
	}
	diff /= float64(tot_blocks)
	if cnt > 0 {
		fmt.Printf("Projected weekly income : %.0f BCH,  estimated hashrate : %s\n",
			7*24*float64(totbtc)/float64(hrs)/1e8,
			common.HashrateToString(float64(cnt)/float64(6*hrs)*diff*7158278.826667))
	}
	bph := float64(tot_blocks) / float64(hrs)
	fmt.Printf("Total network hashrate : %s @ average diff %.0f  (%.2f bph)\n",
		common.HashrateToString(bph/6*diff*7158278.826667), diff, bph)
	fmt.Printf("%d blocks in %d hours. Average size %.1f KB,  next diff in %d blocks\n",
		tot_blocks, hrs, float64(tot_blocks_len/tot_blocks)/1e3, 2016-bte.Height%2016)

	fmt.Printf("\nSegWit Voting: %d (%.1f%%)\n", segwit_cnt, float64(segwit_cnt)*100/float64(tot_blocks))
	fmt.Println()
	fmt.Println("BU Voting")
	for k, v := range eb_ad_voting {
		fmt.Printf(" %s \t %d \t %.1f%%\n", k, v, float64(v)*100/float64(tot_blocks))
	}
}

func init() {
	newUi("minerstat m", false, do_mining, "Look for the miner ID in recent blocks (optionally specify number of hours)")
}
