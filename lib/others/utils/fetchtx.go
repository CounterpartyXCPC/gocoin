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

// File:        fetchtx.go
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

// Download (and re-assemble) raw transaction from blockexplorer.com
func GetTxFromExplorer(txid *bch.Uint256, testnet bool) (rawtx []byte) {
	var url string
	if testnet {
		url = "http://testnet.blockexplorer.com/api/rawtx/" + txid.String()
	} else {
		url = "http://blockexplorer.com/api/rawtx/" + txid.String()
	}
	r, er := http.Get(url)
	if er == nil {
		if r.StatusCode == 200 {
			defer r.Body.Close()
			c, _ := ioutil.ReadAll(r.Body)
			var txx struct {
				Raw string `json:"rawtx"`
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

// Download raw transaction from webbch.com
func GetTxFromWebBTC(txid *bch.Uint256) (raw []byte) {
	url := "https://webbch.com/tx/" + txid.String() + ".bin"
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

// Download (and re-assemble) raw transaction from blockexplorer.com
func GetTxFromBlockchainInfo(txid *bch.Uint256) (rawtx []byte) {
	url := "https://blockchain.info/tx/" + txid.String() + "?format=hex"
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
		fmt.Println("blockchain.info:", er.Error())
	}
	return
}

// Download (and re-assemble) raw transaction from blockcypher.com
func GetTxFromBlockcypher(txid *bch.Uint256, currency string) (rawtx []byte) {
	var url string
	url = "https://api.blockcypher.com/v1/" + currency + "/main/txs/" + txid.String() + "?limit=1000&instart=1000&outstart=1000&includeHex=true"
	r, er := http.Get(url)
	if er == nil {
		if r.StatusCode == 200 {
			defer r.Body.Close()
			c, _ := ioutil.ReadAll(r.Body)
			var txx struct {
				Raw string `json:"hex"`
			}
			er = json.Unmarshal(c[:], &txx)
			if er == nil {
				rawtx, er = hex.DecodeString(txx.Raw)
			}
		} else {
			fmt.Println("blockcypher.com StatusCode=", r.StatusCode)
		}
	}
	if er != nil {
		fmt.Println("blockcypher.com:", er.Error())
	}
	return
}

func verify_txid(txid *bch.Uint256, rawtx []byte) bool {
	tx, _ := bch.NewTx(rawtx)
	if tx == nil {
		return false
	}
	tx.SetHash(rawtx)
	return txid.Equal(&tx.Hash)
}

// Download raw transaction from a web server (try one after another)
func GetTxFromWeb(txid *bch.Uint256) (raw []byte) {
	raw = GetTxFromExplorer(txid, false)
	if raw != nil && verify_txid(txid, raw) {
		//println("GetTxFromExplorer - OK")
		return
	}

	raw = GetTxFromWebBTC(txid)
	if raw != nil && verify_txid(txid, raw) {
		//println("GetTxFromWebBTC - OK")
		return
	}

	raw = GetTxFromBlockchainInfo(txid)
	if raw != nil && verify_txid(txid, raw) {
		//println("GetTxFromBlockchainInfo - OK")
		return
	}

	raw = GetTxFromBlockcypher(txid, "btc")
	if raw != nil && verify_txid(txid, raw) {
		//println("GetTxFromBlockcypher - OK")
		return
	}

	return
}

// Download testnet's raw transaction from a web server
func GetTestnetTxFromWeb(txid *bch.Uint256) (raw []byte) {
	raw = GetTxFromExplorer(txid, true)
	if raw != nil && verify_txid(txid, raw) {
		//println("GetTxFromExplorer - OK")
		return
	}

	raw = GetTxFromBlockcypher(txid, "btc-testnet")
	if raw != nil && verify_txid(txid, raw) {
		//println("GetTxFromBlockcypher - OK")
		return
	}

	return
}
