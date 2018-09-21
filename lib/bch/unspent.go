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

// File:		unspent.go
// Description:	Bictoin Cash Cash Unspent Transaction Package

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

package bch

import (
	"encoding/binary"
	"fmt"
)

type AllUnspentTx []*OneUnspentTx

// Returned by GetUnspentFromPkScr
type OneUnspentTx struct {
	TxPrevOut
	Value   uint64
	MinedAt uint32
	*BtcAddr
	destAddr string
}

func (x AllUnspentTx) Len() int {
	return len(x)
}

func (x AllUnspentTx) Less(i, j int) bool {
	if x[i].MinedAt == x[j].MinedAt {
		if x[i].TxPrevOut.Hash == x[j].TxPrevOut.Hash {
			return x[i].TxPrevOut.Vout < x[j].TxPrevOut.Vout
		}
		return binary.LittleEndian.Uint64(x[i].TxPrevOut.Hash[24:32]) <
			binary.LittleEndian.Uint64(x[j].TxPrevOut.Hash[24:32])
	}
	return x[i].MinedAt < x[j].MinedAt
}

func (x AllUnspentTx) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (ou *OneUnspentTx) String() (s string) {
	s = fmt.Sprintf("%15.8f  ", float64(ou.Value)/1e8) + ou.TxPrevOut.String()
	if ou.BtcAddr != nil {
		s += " " + ou.DestAddr() + ou.BtcAddr.Label()
	}
	if ou.MinedAt != 0 {
		s += fmt.Sprint("  ", ou.MinedAt)
	}
	return
}

func (ou *OneUnspentTx) UnspentTextLine() (s string) {
	s = fmt.Sprintf("%s # %.8f BTC @ %s%s, block %d", ou.TxPrevOut.String(),
		float64(ou.Value)/1e8, ou.DestAddr(), ou.BtcAddr.Label(), ou.MinedAt)
	return
}

func (ou *OneUnspentTx) DestAddr() string {
	if ou.destAddr == "" {
		return ou.BtcAddr.String()
	}
	return ou.destAddr
}
