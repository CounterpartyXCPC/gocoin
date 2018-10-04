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

// File:		address.go
// Description:	Bictoin Cash rpcapi Package

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

package rpcapi

import (
	"encoding/hex"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	//"github.com/counterpartyxcpc/gocoin-cash/client/common"
)

/*

{"result":
	{"isvalid":true,
	"address":"mqzwxBkSH1UKqEAjGwvkj6aV5Gc6BtBCSs",
	"scriptPubKey":"76a91472fc9e6b1bbbd40a66653989a758098bfbf1b54788ac",
	"ismine":false,
	"iswatchonly":false,
	"isscript":false
}
*/

type ValidAddressResponse struct {
	IsValid      bool   `json:"isvalid"`
	Address      string `json:"address"`
	ScriptPubKey string `json:"scriptPubKey"`
	IsMine       bool   `json:"ismine"`
	IsWatchOnly  bool   `json:"iswatchonly"`
	IsScript     bool   `json:"isscript"`
}

type InvalidAddressResponse struct {
	IsValid bool `json:"isvalid"`
}

func ValidateAddress(addr string) interface{} {
	a, e := bch.NewAddrFromString(addr)
	if e != nil {
		return new(InvalidAddressResponse)
	}
	res := new(ValidAddressResponse)
	res.IsValid = true
	res.Address = addr
	res.ScriptPubKey = hex.EncodeToString(a.OutScript())
	return res
	//res.IsMine = false
	//res.IsWatchOnly = false
	//res.IsScript = false
}
