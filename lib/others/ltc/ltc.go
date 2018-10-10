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

// File:        ltc.go
// Description: Bictoin Cash Cash ltc Package

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

package ltc

import (
	"bytes"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/utils"
)

const LTC_ADDR_VERSION = 48
const LTC_ADDR_VERSION_SCRIPT = 50

// LTC signing uses different seed string
func HashFromMessage(msg []byte, out []byte) {
	const MessageMagic = "Litecoin Signed Message:\n"
	b := new(bytes.Buffer)
	bch.WriteVlen(b, uint64(len(MessageMagic)))
	b.Write([]byte(MessageMagic))
	bch.WriteVlen(b, uint64(len(msg)))
	b.Write(msg)
	bch.ShaHash(b.Bytes(), out)
}

func AddrVerPubkey(testnet bool) byte {
	if !testnet {
		return LTC_ADDR_VERSION
	}
	return bch.AddrVerPubkey(testnet)
}

// At some point Litecoin started using addresses with M in front (version 50) - see github issue #41
func AddrVerScript(testnet bool) byte {
	if !testnet {
		return LTC_ADDR_VERSION_SCRIPT
	}
	return bch.AddrVerScript(testnet)
}

func NewAddrFromPkScript(scr []byte, testnet bool) (ad *bch.BtcAddr) {
	ad = bch.NewAddrFromPkScript(scr, testnet)
	if ad != nil && ad.Version == bch.AddrVerPubkey(false) {
		ad.Version = LTC_ADDR_VERSION
	}
	return
}

func GetUnspent(addr *bch.BtcAddr) (res utxo.AllUnspentTx) {
	var er error

	res, er = utils.GetUnspentFromBlockcypher(addr, "ltc")
	if er == nil {
		return
	}
	println("GetUnspentFromBlockcypher:", er.Error())

	return
}

// Download testnet's raw transaction from a web server
func GetTxFromWeb(txid *bch.Uint256) (raw []byte) {
	raw = utils.GetTxFromBlockcypher(txid, "ltc")
	if raw != nil && txid.Equal(bch.NewSha2Hash(raw)) {
		//println("GetTxFromBlockcypher - OK")
		return
	}

	return
}
