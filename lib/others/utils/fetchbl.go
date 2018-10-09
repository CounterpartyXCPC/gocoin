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

// File:        fetchbl.go
// Description: Bictoin Cash Cash utils Package

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

package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

// @todo oct 9, 2018 -- Update 'GetBlockFromExplorer' to BCH sources

// https://blockchain.info/block/000000000000000000871f4f01a389bda59e568ead8d0fd45fc7cc1919d2666e?format=hex
// https://webbch.com/block/0000000000000000000cdc0d2a9b33c2d4b34b4d4fa8920f074338d0dc1164dc.bin
// https://blockexplorer.com/api/rawblock/0000000000000000000cdc0d2a9b33c2d4b34b4d4fa8920f074338d0dc1164dc

// Download (and re-assemble) raw block from blockexplorer.com
func GetBlockFromExplorer(hash *bch.Uint256) (rawtx []byte) {
	url := "http://blockexplorer.com/api/rawblock/" + hash.String()
	r, er := http.Get(url)
	if er == nil {
		if r.StatusCode == 200 {
			defer r.Body.Close()
			c, _ := ioutil.ReadAll(r.Body)
			var txx struct {
				Raw string `json:"rawblock"`
			}
			er = json.Unmarshal(c[:], &txx)
			if er == nil {
				rawtx, er = hex.DecodeString(txx.Raw)
			}
		} else {
			fmt.Println("blockexplorer.com StatusCode=", r.StatusCode)
		}
	}
	if er != nil {
		fmt.Println("blockexplorer.com:", er.Error())
	}
	return
}

// Download raw block from webbch.com
func GetBlockFromWebBTC(hash *bch.Uint256) (raw []byte) {
	url := "https://webbch.com/block/" + hash.String() + ".bin"
	r, er := http.Get(url)
	if er == nil {
		if r.StatusCode == 200 {
			raw, _ = ioutil.ReadAll(r.Body)
			r.Body.Close()
		} else {
			fmt.Println("webbch.com StatusCode=", r.StatusCode)
		}
	}
	if er != nil {
		fmt.Println("webbch.com:", er.Error())
	}
	return
}

// Download raw block from blockexplorer.com
func GetBlockFromBlockchainInfo(hash *bch.Uint256) (rawtx []byte) {
	url := "https://blockchain.info/block/" + hash.String() + "?format=hex"
	r, er := http.Get(url)
	if er == nil {
		if r.StatusCode == 200 {
			defer r.Body.Close()
			rawhex, _ := ioutil.ReadAll(r.Body)
			rawtx, er = hex.DecodeString(string(rawhex))
		} else {
			fmt.Println("blockchain.info StatusCode=", r.StatusCode)
		}
	}
	if er != nil {
		fmt.Println("blockexplorer.com:", er.Error())
	}
	return
}

func IsBlockOK(raw []byte, hash *bch.Uint256) (bl *bch.BchBlock) {
	var er error
	bl, er = bch.NewBchBlock(raw)
	if er != nil {
		return
	}
	if !bl.Hash.Equal(hash) {
		return nil
	}
	er = bl.BuildTxList()
	if er != nil {
		return nil
	}
	if !bl.MerkleRootMatch() {
		return nil
	}
	return
}

// Download raw block from a web server (try one after another)
func GetBlockFromWeb(hash *bch.Uint256) (bl *bch.BchBlock) {
	var raw []byte

	raw = GetBlockFromBlockchainInfo(hash)
	if bl = IsBlockOK(raw, hash); bl != nil {
		//println("GetTxFromBlockchainInfo - OK")
		return
	}

	raw = GetBlockFromWebBTC(hash)
	if bl = IsBlockOK(raw, hash); bl != nil {
		//println("GetTxFromWebBTC - OK")
		return
	}

	raw = GetBlockFromExplorer(hash)
	if bl = IsBlockOK(raw, hash); bl != nil {
		//println("GetTxFromExplorer - OK")
		return
	}

	return
}
