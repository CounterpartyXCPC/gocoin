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

// File:		wallethd.go
// Description:	Bictoin Cash Cash HD Wallet Package

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

/*
This code originates from:
 * https://github.com/WeMeetAgain/go-hdwallet
*/

package bch

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"

	"github.com/counterpartyxcpc/gocoin-cash/lib/secp256k1"
)

const (
	Public      = uint32(0x0488B21E)
	Private     = uint32(0x0488ADE4)
	TestPublic  = uint32(0x043587CF)
	TestPrivate = uint32(0x04358394)
)

// HDWallet defines the components of a hierarchical deterministic wallet
type HDWallet struct {
	Prefix   uint32
	Depth    byte
	Checksum [4]byte
	I        uint32
	ChCode   []byte //32 bytes
	Key      []byte //33 bytes
}

// Child returns the ith child of wallet w. Values of i >= 2^31
// signify private key derivation. Attempting private key derivation
// with a public key will throw an error.
func (w *HDWallet) Child(i uint32) (res *HDWallet) {
	var ha, newkey []byte
	var chksum [20]byte

	if w.Prefix == Private || w.Prefix == TestPrivate {
		pub := PublicFromPrivate(w.Key[1:], true)
		mac := hmac.New(sha512.New, w.ChCode)
		if i >= uint32(0x80000000) {
			mac.Write(w.Key)
		} else {
			mac.Write(pub)
		}
		binary.Write(mac, binary.BigEndian, i)
		ha = mac.Sum(nil)
		newkey = append([]byte{0}, DeriveNextPrivate(ha[:32], w.Key[1:])...)
		RimpHash(pub, chksum[:])
	} else if w.Prefix == Public || w.Prefix == TestPublic {
		mac := hmac.New(sha512.New, w.ChCode)
		if i >= uint32(0x80000000) {
			panic("HDWallet.Child(): Private derivation on Public key")
		}
		mac.Write(w.Key)
		binary.Write(mac, binary.BigEndian, i)
		ha = mac.Sum(nil)
		newkey = DeriveNextPublic(w.Key, ha[:32])
		RimpHash(w.Key, chksum[:])
	} else {
		panic("HDWallet.Child(): Unexpected Prefix")
	}
	res = new(HDWallet)
	res.Prefix = w.Prefix
	res.Depth = w.Depth + 1
	copy(res.Checksum[:], chksum[:4])
	res.I = i
	res.ChCode = ha[32:]
	res.Key = newkey
	return
}

// Serialize returns the serialized form of the wallet.
// vbytes || depth || fingerprint || i || chaincode || key
func (w *HDWallet) Serialize() []byte {
	var tmp [32]byte
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, w.Prefix)
	b.WriteByte(w.Depth)
	b.Write(w.Checksum[:])
	binary.Write(b, binary.BigEndian, w.I)
	b.Write(w.ChCode)
	b.Write(w.Key)
	ShaHash(b.Bytes(), tmp[:])
	return append(b.Bytes(), tmp[:4]...)
}

// String returns the base58-encoded string form of the wallet.
func (w *HDWallet) String() string {
	return Encodeb58(w.Serialize())
}

// StringWallet returns a wallet given a base58-encoded extended key
func StringWallet(data string) (*HDWallet, error) {
	dbin := Decodeb58(data)
	if err := ByteCheck(dbin); err != nil {
		return &HDWallet{}, err
	}
	var res [32]byte
	ShaHash(dbin[:(len(dbin)-4)], res[:])
	if !bytes.Equal(res[:4], dbin[(len(dbin)-4):]) {
		return &HDWallet{}, errors.New("StringWallet: Invalid checksum")
	}
	r := new(HDWallet)
	r.Prefix = binary.BigEndian.Uint32(dbin[0:4])
	r.Depth = dbin[4]
	copy(r.Checksum[:], dbin[5:9])
	r.I = binary.BigEndian.Uint32(dbin[9:13])
	r.ChCode = dbin[13:45]
	r.Key = dbin[45:78]
	return r, nil
}

// Pub returns a new wallet which is the public key version of w.
// If w is a public key, Pub returns a copy of w
func (w *HDWallet) Pub() *HDWallet {
	if w.Prefix == Public || w.Prefix == TestPublic {
		r := new(HDWallet)
		*r = *w
		return r
	} else {
		return &HDWallet{Prefix: Public, Depth: w.Depth, Checksum: w.Checksum,
			I: w.I, ChCode: w.ChCode, Key: PublicFromPrivate(w.Key[1:], true)}
	}
}

// StringChild returns the ith base58-encoded extended key of a base58-encoded extended key.
func StringChild(data string, i uint32) string {
	w, err := StringWallet(data)
	if err != nil {
		return ""
	} else {
		w = w.Child(i)
		return w.String()
	}
}

//StringToAddress returns the Bitcoin address of a base58-encoded extended key.
func StringAddress(data string) (string, error) {
	w, err := StringWallet(data)
	if err != nil {
		return "", err
	}

	return NewAddrFromPubkey(w.Key, AddrVerPubkey(w.Prefix == TestPublic || w.Prefix == TestPrivate)).String(), nil
}

// PublicAddress returns base58 encoded public address of the given HD key
func (w *HDWallet) PubAddr() *BtcAddr {
	var pub []byte
	if w.Prefix == Private || w.Prefix == TestPrivate {
		pub = PublicFromPrivate(w.Key[1:], true)
	} else {
		pub = w.Key
	}
	return NewAddrFromPubkey(pub, AddrVerPubkey(w.Prefix == TestPublic || w.Prefix == TestPrivate))
}

// MasterKey returns a new wallet given a random seed.
func MasterKey(seed []byte, testnet bool) *HDWallet {
	key := []byte("Bitcoin seed")
	mac := hmac.New(sha512.New, key)
	mac.Write(seed)
	I := mac.Sum(nil)
	res := &HDWallet{ChCode: I[len(I)/2:], Key: append([]byte{0}, I[:len(I)/2]...)}
	if testnet {
		res.Prefix = TestPrivate
	} else {
		res.Prefix = Private
	}
	return res
}

// StringCheck is a validation check of a base58-encoded extended key.
func StringCheck(key string) error {
	return ByteCheck(Decodeb58(key))
}

// Verifies consistency of a serialized HD address
func ByteCheck(dbin []byte) error {
	// check proper length
	if len(dbin) != 82 {
		return errors.New("ByteCheck: Unexpected length")
	}

	// check for correct Public or Private Prefix
	vb := binary.BigEndian.Uint32(dbin[:4])
	if vb != Public && vb != Private && vb != TestPublic && vb != TestPrivate {
		return errors.New("ByteCheck: Unexpected Prefix")
	}

	// if Public, check x coord is on curve
	if vb == Public || vb == TestPublic {
		var xy secp256k1.XY
		xy.ParsePubkey(dbin[45:78])
		if !xy.IsValid() {
			return errors.New("ByteCheck: Invalid public key")
		}
	}
	return nil
}

// Returns first 32 bits, as expected for sepcific HD address
func HDKeyPrefix(private, testnet bool) uint32 {
	if private {
		if testnet {
			return TestPrivate
		} else {
			return Private
		}
	} else {
		if testnet {
			return TestPublic
		} else {
			return Public
		}
	}
}
