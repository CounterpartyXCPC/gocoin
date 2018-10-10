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

// File:        unspent.go
// Description: Bictoin Cash Cash utils Package

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

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
)

// https://api.blockchair.com/bitcoin/outputs?q=is_spent(0),recipient(bc1qdvpxmyvyu9urhadl6sk69gcjsfqsvrjsqfk5aq)

func GetUnspentFromExplorer(addr *bch.BtcAddr, testnet bool) (res utxo.AllUnspentTx, er error) {
	var r *http.Response
	if testnet {
		r, er = http.Get("https://testnet.blockexplorer.com/api/addr/" + addr.String() + "/utxo")
	} else {
		r, er = http.Get("https://blockexplorer.com/api/addr/" + addr.String() + "/utxo")
	}
	if er != nil {
		return
	}
	if r.StatusCode != 200 {
		er = errors.New(fmt.Sprint("HTTP StatusCode ", r.StatusCode))
		return
	}

	c, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var result []struct {
		Addr   string `json:"address"`
		TxID   string `json:"txid"`
		Vout   uint32 `json:"vout"`
		Value  uint64 `json:"satoshis"`
		Height uint32 `json:"height"`
	}

	er = json.Unmarshal(c, &result)
	if er != nil {
		return
	}

	for _, r := range result {
		ur := new(utxo.OneUnspentTx)
		id := bch.NewUint256FromString(r.TxID)
		if id == nil {
			er = errors.New(fmt.Sprint("Bad TXID:", r.TxID))
			return
		}
		copy(ur.TxPrevOut.Hash[:], id.Hash[:])
		ur.TxPrevOut.Vout = r.Vout
		ur.Value = r.Value
		ur.MinedAt = r.Height
		ur.BtcAddr = addr
		res = append(res, ur)
	}

	return
}

func GetUnspentFromBlockchainInfo(addr *bch.BtcAddr) (res utxo.AllUnspentTx, er error) {
	var r *http.Response
	r, er = http.Get("https://blockchain.info/unspent?active=" + addr.String())
	if er != nil {
		return
	}
	if r.StatusCode != 200 {
		er = errors.New(fmt.Sprint("HTTP StatusCode ", r.StatusCode))
		return
	}

	c, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var result struct {
		U []struct {
			TxID  string `json:"tx_hash_big_endian"`
			Vout  uint32 `json:"tx_output_n"`
			Value uint64 `json:"value"`
		} `json:"unspent_outputs"`
	}

	er = json.Unmarshal(c, &result)
	if er != nil {
		return
	}

	for _, r := range result.U {
		ur := new(utxo.OneUnspentTx)
		id := bch.NewUint256FromString(r.TxID)
		if id == nil {
			er = errors.New(fmt.Sprint("Bad TXID:", r.TxID))
			return
		}
		copy(ur.TxPrevOut.Hash[:], id.Hash[:])
		ur.TxPrevOut.Vout = r.Vout
		ur.Value = r.Value
		//ur.MinedAt = r.Height
		ur.BtcAddr = addr
		res = append(res, ur)
	}

	return
}

func GetUnspentFromBlockcypher(addr *bch.BtcAddr, currency string) (res utxo.AllUnspentTx, er error) {
	var r *http.Response

	r, er = http.Get("https://api.blockcypher.com/v1/" + currency + "/main/addrs/" + addr.String() + "?unspentOnly=true")

	if er != nil {
		return
	}
	if r.StatusCode != 200 {
		er = errors.New(fmt.Sprint("HTTP StatusCode ", r.StatusCode))
		return
	}

	c, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var result struct {
		Addr string `json:"address"`
		Outs []struct {
			TxID   string `json:"tx_hash"`
			Vout   uint32 `json:"tx_output_n"`
			Value  uint64 `json:"value"`
			Height uint32 `json:"block_height"`
		} `json:"txrefs"`
	}

	er = json.Unmarshal(c, &result)
	if er != nil {
		return
	}

	for _, r := range result.Outs {
		ur := new(utxo.OneUnspentTx)
		id := bch.NewUint256FromString(r.TxID)
		if id == nil {
			er = errors.New(fmt.Sprint("Bad TXID:", r.TxID))
			return
		}
		copy(ur.TxPrevOut.Hash[:], id.Hash[:])
		ur.TxPrevOut.Vout = r.Vout
		ur.Value = r.Value
		ur.MinedAt = r.Height
		ur.BtcAddr = addr
		res = append(res, ur)
	}

	return
}

// currency is either "bitcoin" or "bitcoin-cash"
func GetUnspentFromBlockchair(addr *bch.BtcAddr, currency string) (res utxo.AllUnspentTx, er error) {
	var r *http.Response

	r, er = http.Get("https://api.blockchair.com/" + currency + "/outputs?q=is_spent(0),recipient(" + addr.String() + ")")

	if er != nil {
		return
	}
	if r.StatusCode != 200 {
		er = errors.New(fmt.Sprint("HTTP StatusCode ", r.StatusCode))
		return
	}

	c, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var result struct {
		Outs []struct {
			TxID   string `json:"transaction_hash"`
			Vout   uint32 `json:"index"`
			Value  string `json:"value"`
			Height uint32 `json:"block_id"`
		} `json:"data"`
	}

	er = json.Unmarshal(c, &result)
	if er != nil {
		return
	}

	for _, r := range result.Outs {
		ur := new(utxo.OneUnspentTx)
		id := bch.NewUint256FromString(r.TxID)
		if id == nil {
			er = errors.New(fmt.Sprint("Bad TXID:", r.TxID))
			return
		}
		copy(ur.TxPrevOut.Hash[:], id.Hash[:])
		ur.TxPrevOut.Vout = r.Vout
		if ur.Value, er = strconv.ParseUint(r.Value, 10, 64); er != nil {
			er = errors.New(fmt.Sprint("Bad Value:", r.Value))
			return
		}
		ur.MinedAt = r.Height
		ur.BtcAddr = addr
		res = append(res, ur)
	}

	return
}

func GetUnspent(addr *bch.BtcAddr) (res utxo.AllUnspentTx) {
	var er error

	res, er = GetUnspentFromBlockchair(addr, "bitcoin")
	if er == nil {
		return
	}
	println("GetUnspentFromBlockchair:", er.Error())

	res, er = GetUnspentFromExplorer(addr, false)
	if er == nil {
		return
	}
	println("GetUnspentFromExplorer:", er.Error())

	res, er = GetUnspentFromBlockcypher(addr, "btc")
	if er == nil {
		return
	}
	println("GetUnspentFromBlockcypher:", er.Error())

	return
}

func GetUnspentTestnet(addr *bch.BtcAddr) (res utxo.AllUnspentTx) {
	var er error

	res, er = GetUnspentFromExplorer(addr, true)
	if er == nil {
		return
	}
	println("GetUnspentFromExplorer:", er.Error())

	res, er = GetUnspentFromBlockcypher(addr, "btc-testnet")
	if er == nil {
		return
	}
	println("GetUnspentFromBlockcypher:", er.Error())

	return
}
