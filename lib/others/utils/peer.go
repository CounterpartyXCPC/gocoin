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

// File:        peer.go
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
	"encoding/binary"
	"hash/crc64"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

type OnePeer struct {
	bch.NetAddr
	Time   uint32 // When seen last time
	Banned uint32 // time when this address baned or zero if never
}

var crctab = crc64.MakeTable(crc64.ISO)

/*
Serialized peer record (all values are LSB unless specified otherwise):
 [0:4] - Unix timestamp of when last the peer was seen
 [4:12] - Services
 [12:24] - IPv6 (network order)
 [24:28] - IPv4 (network order)
 [28:30] - TCP port (big endian)
 [30:34] - OPTIONAL: if present, unix timestamp of when the peer was banned
*/

func NewPeer(v []byte) (p *OnePeer) {
	if len(v) < 30 {
		println("NewPeer: unexpected length", len(v))
		return
	}
	p = new(OnePeer)
	p.Time = binary.LittleEndian.Uint32(v[0:4])
	p.Services = binary.LittleEndian.Uint64(v[4:12])
	copy(p.Ip6[:], v[12:24])
	copy(p.Ip4[:], v[24:28])
	p.Port = binary.BigEndian.Uint16(v[28:30])
	if len(v) >= 34 {
		p.Banned = binary.LittleEndian.Uint32(v[30:34])
	}
	return
}

func (p *OnePeer) Bytes() (res []byte) {
	if p.Banned != 0 {
		res = make([]byte, 34)
		binary.LittleEndian.PutUint32(res[30:34], p.Banned)
	} else {
		res = make([]byte, 30)
	}
	binary.LittleEndian.PutUint32(res[0:4], p.Time)
	binary.LittleEndian.PutUint64(res[4:12], p.Services)
	copy(res[12:24], p.Ip6[:])
	copy(res[24:28], p.Ip4[:])
	binary.BigEndian.PutUint16(res[28:30], p.Port)
	return
}

func (p *OnePeer) UniqID() uint64 {
	h := crc64.New(crctab)
	h.Write(p.Ip6[:])
	h.Write(p.Ip4[:])
	h.Write([]byte{byte(p.Port >> 8), byte(p.Port)})
	return h.Sum64()
}
