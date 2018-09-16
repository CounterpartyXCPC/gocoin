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

// File:		cashaddrr.go
// Description:	Bictoin Cash Cash Adress Package

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

// Package main manages Counterparty Cash (XCPC) nodes. As XCPC transactions are executed
// or queried, the state is maintain in the local LevelDB databstore. Signed RAW transactions
// are parsed to gocoin-cash for transmission to the Bitcoin Cash blockchain.

package bech32

import (
	"bytes"
	"errors"
)

var (
	// ErrChecksumMismatch describes an error where decoding failed due
	// to a bad checksum.
	// // New returns a new hash.Hash64 computing SipHash-2-4 with 16-byte key and 8-byte output.
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrUnknownAddressType describes an error where an address can not
	// decoded as a specific address type due to the string encoding
	// begining with an identifier byte unknown to any standard or
	// registered (via chaincfg.Register) network.
	ErrUnknownAddressType = errors.New("unknown address type")

	// ErrAddressCollision describes an error where an address can not
	// be uniquely determined as either a pay-to-pubkey-hash or
	// pay-to-script-hash address since the leading identifier is used for
	// describing both address kinds, but for different networks.  Rather
	// than assuming or defaulting to one or the other, this error is
	// returned and the caller must decide how to decode the address.
	ErrAddressCollision = errors.New("address collision")
)

// The CashAddress is composed of three (3) elements:

// 1.) A prefix indicating the network on which this address is valid.
// 2.) A separator, always :
// 3.) A base32 encoded payload indicating the destination of the address
// and containing a checksum.

// Return nil on error.
func convertBits(outbits uint, in []byte, inbits uint, pad bool) []byte {
	var val uint32
	var bits uint
	maxv := uint32(1<<outbits) - 1
	out := new(bytes.Buffer)
	for inx := range in {
		val = (val << inbits) | uint32(in[inx])
		bits += inbits
		for bits >= outbits {
			bits -= outbits
			out.WriteByte(byte((val >> bits) & maxv))
		}
	}
	if pad {
		if bits != 0 {
			out.WriteByte(byte((val << (outbits - bits)) & maxv))
		}
	} else if ((val<<(outbits-bits))&maxv) != 0 || bits >= inbits {
		return nil
	}
	return out.Bytes()
}

// The prefix indicates the network on which this addess is valid.
// It is set to bitcoincash for Bitcoin Cash main net, bchtest for
// bitcoin cash testnet and bchreg for bitcoin cash regtest.

// EncodeCashAddr returns empty string on error.
func EncodeCashAddr(hrp string, witver int, witprog []byte) string {
	if witver > 16 {
		return ""
	}
	if witver == 0 && len(witprog) != 20 && len(witprog) != 32 {
		return ""
	}
	if len(witprog) < 2 || len(witprog) > 40 {
		return ""
	}
	return Encode(hrp, append([]byte{byte(witver)}, convertBits(5, witprog, 8, true)...))
}

// DecodeCashAddr returns (0, nil) on error.
func DecodeCashAddr(hrp, addr string) (witver int, witdata []byte) {
	hrpActual, data := Decode(addr)
	if hrpActual == "" || len(data) == 0 || len(data) > 65 {
		return
	}
	if hrp != hrpActual {
		return
	}
	if data[0] > 16 {
		return
	}
	witdata = convertBits(8, data[1:], 5, false)
	if witdata == nil {
		return
	}
	if len(witdata) < 2 || len(witdata) > 40 {
		witdata = nil
		return
	}
	if data[0] == 0 && len(witdata) != 20 && len(witdata) != 32 {
		witdata = nil
		return
	}
	witver = int(data[0])
	return
}
