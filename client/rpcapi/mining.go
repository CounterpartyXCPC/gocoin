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
// Description:	Bictoin Cash rpcapi Package

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

package rpcapi

import (
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

const MAX_TXS_LEN = 999e3 // 999KB, with 1KB margin to not exceed 1MB with conibase

type OneTransaction struct {
	Data    string `json:"data"`
	Hash    string `json:"hash"`
	Depends []uint `json:"depends"`
	Fee     uint64 `json:"fee"`
	Sigops  uint64 `json:"sigops"`
}

type GetBlockTemplateResp struct {
	Capabilities      []string         `json:"capabilities"`
	Version           uint32           `json:"version"`
	PreviousBlockHash string           `json:"previousblockhash"`
	Transactions      []OneTransaction `json:"transactions"`
	Coinbaseaux       struct {
		Flags string `json:"flags"`
	} `json:"coinbaseaux"`
	Coinbasevalue uint64   `json:"coinbasevalue"`
	Longpollid    string   `json:"longpollid"`
	Target        string   `json:"target"`
	Mintime       uint     `json:"mintime"`
	Mutable       []string `json:"mutable"`
	Noncerange    string   `json:"noncerange"`
	Sigoplimit    uint     `json:"sigoplimit"`
	Sizelimit     uint     `json:"sizelimit"`
	Curtime       uint     `json:"curtime"`
	Bits          string   `json:"bits"`
	Height        uint     `json:"height"`
}

type RpcGetBlockTemplateResp struct {
	Id     interface{}          `json:"id"`
	Result GetBlockTemplateResp `json:"result"`
	Error  interface{}          `json:"error"`
}

func GetNextBlockTemplate(r *GetBlockTemplateResp) {
	var zer [32]byte

	common.Last.Mutex.Lock()

	r.Curtime = uint(time.Now().Unix())
	r.Mintime = uint(common.Last.BchBlock.GetMedianTimePast()) + 1
	if r.Curtime < r.Mintime {
		r.Curtime = r.Mintime
	}
	height := common.Last.BchBlock.Height + 1
	bits := common.BchBlockChain.GetNextWorkRequired(common.Last.BchBlock, uint32(r.Curtime))
	target := bch.SetCompact(bits).Bytes()

	r.Capabilities = []string{"proposal"}
	r.Version = 4
	r.PreviousBlockHash = common.Last.BchBlock.BchBlockHash.String()
	r.Transactions, r.Coinbasevalue = GetTransactions(height, uint32(r.Mintime))
	r.Coinbasevalue += bch.GetBlockReward(height)
	r.Coinbaseaux.Flags = ""
	r.Longpollid = r.PreviousBlockHash
	r.Target = hex.EncodeToString(append(zer[:32-len(target)], target...))
	r.Mutable = []string{"time", "transactions", "prevblock"}
	r.Noncerange = "00000000ffffffff"
	r.Sigoplimit = bch.MAX_BLOCK_SIGOPS_COST / bch.WITNESS_SCALE_FACTOR
	r.Sizelimit = 1e6
	r.Bits = fmt.Sprintf("%08x", bits)
	r.Height = uint(height)

	last_given_time = uint32(r.Curtime)
	last_given_mintime = uint32(r.Mintime)

	common.Last.Mutex.Unlock()
}

/* memory pool transaction sorting stuff */
type one_mining_tx struct {
	*network.OneTxToSend
	depends []uint
	startat int
}

type sortedTxList []*one_mining_tx

func (tl sortedTxList) Len() int           { return len(tl) }
func (tl sortedTxList) Swap(i, j int)      { tl[i], tl[j] = tl[j], tl[i] }
func (tl sortedTxList) Less(i, j int) bool { return tl[j].Fee < tl[i].Fee }

var txs_so_far map[[32]byte]uint
var totlen int
var sigops uint64

func get_next_tranche_of_txs(height, timestamp uint32) (res sortedTxList) {
	var unsp *bch.TxOut
	var all_inputs_found bool
	for _, v := range network.TransactionsToSend {
		tx := v.Tx

		if _, ok := txs_so_far[tx.Hash.Hash]; ok {
			continue
		}

		if !tx.IsFinal(height, timestamp) {
			continue
		}

		if totlen+len(v.Raw) > 1e6 {
			//println("Too many txs - limit to 999000 bytes")
			return
		}
		totlen += len(v.Raw)

		if sigops+v.SigopsCost > bch.MAX_BLOCK_SIGOPS_COST {
			//println("Too many sigops - limit to 999000 bytes")
			return
		}
		sigops += v.SigopsCost

		all_inputs_found = true
		var depends []uint
		for i := range tx.TxIn {
			unsp = common.BchBlockChain.Unspent.UnspentGet(&tx.TxIn[i].Input)
			if unsp == nil {
				// not found in the confirmed blocks
				// check if txid is in txs_so_far
				if idx, ok := txs_so_far[tx.TxIn[i].Input.Hash]; !ok {
					// also not in txs_so_far
					all_inputs_found = false
					break
				} else {
					depends = append(depends, idx)
				}
			}
		}

		if all_inputs_found {
			res = append(res, &one_mining_tx{OneTxToSend: v, depends: depends, startat: 1 + len(txs_so_far)})
		}
	}
	return
}

func GetTransactions(height, timestamp uint32) (res []OneTransaction, totfees uint64) {

	network.TxMutex.Lock()
	defer network.TxMutex.Unlock()

	var cnt int
	var sorted sortedTxList
	txs_so_far = make(map[[32]byte]uint)
	totlen = 0
	sigops = 0
	//println("\ngetting txs from the pool of", len(network.TransactionsToSend), "...")
	for {
		new_piece := get_next_tranche_of_txs(height, timestamp)
		if new_piece.Len() == 0 {
			break
		}
		//println("adding another", len(new_piece))
		sort.Sort(new_piece)

		for i := 0; i < len(new_piece); i++ {
			txs_so_far[new_piece[i].Tx.Hash.Hash] = uint(1 + len(sorted) + i)
		}

		sorted = append(sorted, new_piece...)
	}
	/*if len(txs_so_far)!=len(network.TransactionsToSend) {
		println("ERROR: txs_so_far len", len(txs_so_far), " - please report!")
	}*/
	txs_so_far = nil // leave it for the garbage collector

	res = make([]OneTransaction, len(sorted))
	for cnt = 0; cnt < len(sorted); cnt++ {
		v := sorted[cnt]
		res[cnt].Data = hex.EncodeToString(v.Raw)
		res[cnt].Hash = v.Tx.Hash.String()
		res[cnt].Fee = v.Fee
		res[cnt].Sigops = v.SigopsCost
		res[cnt].Depends = v.depends
		totfees += v.Fee
		//println("", cnt+1, v.Tx.Hash.String(), "  turn:", v.startat, "  spb:", int(v.Fee)/len(v.Data), "  depend:", fmt.Sprint(v.depends))
	}

	//println("returning transacitons:", totlen, len(res))
	return
}
