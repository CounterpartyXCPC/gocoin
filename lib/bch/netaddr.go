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

// File:		netaddr.go
// Description:	Bictoin Cash Address Package

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

package bch

import (
	"encoding/binary"
	"fmt"
)

type NetAddr struct {
	Services uint64
	Ip6      [12]byte
	Ip4      [4]byte
	Port     uint16
}

func NewNetAddr(b []byte) (na *NetAddr) {
	if len(b) != 26 {
		println("Incorrect input data length", len(b))
		return
	}
	na = new(NetAddr)
	na.Services = binary.LittleEndian.Uint64(b[0:8])
	copy(na.Ip6[:], b[8:20])
	copy(na.Ip4[:], b[20:24])
	na.Port = binary.BigEndian.Uint16(b[24:26])
	return
}

func (a *NetAddr) Bytes() (res []byte) {
	res = make([]byte, 26)
	binary.LittleEndian.PutUint64(res[0:8], a.Services)
	copy(res[8:20], a.Ip6[:])
	copy(res[20:24], a.Ip4[:])
	binary.BigEndian.PutUint16(res[24:26], a.Port)
	return
}

func (a *NetAddr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", a.Ip4[0], a.Ip4[1], a.Ip4[2], a.Ip4[3], a.Port)
}
