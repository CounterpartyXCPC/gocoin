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

// File:		block.go
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

type BchBlockSubmited struct {
	*bch.BchBlock
	Error string
	Done  sync.WaitGroup
}

var RpcBlocks chan *BchBlockSubmited = make(chan *BchBlockSubmited, 1)

func SubmitBlock(cmd *RpcCommand, resp *RpcResponse, b []byte) {
	var bd []byte
	var er error

	switch uu := cmd.Params.(type) {
	case []interface{}:
		if len(uu) < 1 {
			resp.Error = RpcError{Code: -1, Message: "empty params array"}
			return
		}
		str := uu[0].(string)
		if str[0] == '@' {
			/*
				gocoin special case: if the string starts with @, it's a name of the file with block's binary data
					curl --user gocoinrpc:gocoinpwd --data-binary \
						'{"jsonrpc": "1.0", "id":"curltest", "method": "submitblock", "params": \
							["@450529_000000000000000000cf208f521de0424677f7a87f2f278a1042f38d159565f5.bin"] }' \
						-H 'content-type: text/plain;' http://127.0.0.1:8332/
			*/
			//println("jade z koksem", str[1:])
			bd, er = ioutil.ReadFile(str[1:])
		} else {
			bd, er = hex.DecodeString(str)
		}
		if er != nil {
			resp.Error = RpcError{Code: -3, Message: er.Error()}
			return
		}

	default:
		resp.Error = RpcError{Code: -2, Message: "incorrect params type"}
		return
	}

	bs := new(BchBlockSubmited)

	bs.BchBlock, er = bch.NewBchBlock(bd)
	if er != nil {
		resp.Error = RpcError{Code: -4, Message: er.Error()}
		return
	}

	network.MutexRcv.Lock()
	network.ReceivedBlocks[bs.BchBlock.Hash.BIdx()] = &network.OneReceivedBlock{TmStart: time.Now()}
	network.MutexRcv.Unlock()

	println("new block", bs.BchBlock.Hash.String(), "len", len(bd), "- submitting...")
	bs.Done.Add(1)
	RpcBlocks <- bs
	bs.Done.Wait()
	if bs.Error != "" {
		//resp.Error = RpcError{Code: -10, Message: bs.Error}
		idx := strings.Index(bs.Error, "- RPC_Result:")
		if idx == -1 {
			resp.Result = "inconclusive"
		} else {
			resp.Result = bs.Error[idx+13:]
		}
		println("submiting block error:", bs.Error)
		println("submiting block result:", resp.Result.(string))

		print("time_now:", time.Now().Unix())
		print("  cur_block_ts:", bs.BchBlock.BchBlockTime())
		print("  last_given_now:", last_given_time)
		print("  last_given_min:", last_given_mintime)
		common.Last.Mutex.Lock()
		print("  prev_block_ts:", common.Last.BchBlock.Timestamp())
		common.Last.Mutex.Unlock()
		println()

		return
	}

	// cress check with bitcoind...
	if false {
		bitcoind_result := process_rpc(b)
		json.Unmarshal(bitcoind_result, &resp)
		switch cmd.Params.(type) {
		case string:
			println("\007Block rejected by bitcoind:", resp.Result.(string))
			ioutil.WriteFile(fmt.Sprint(bs.BchBlock.Height, "-", bs.BchBlock.Hash.String()), bd, 0777)
		default:
			println("submiting block verified OK", bs.Error)
		}
	}
}

var last_given_time, last_given_mintime uint32
