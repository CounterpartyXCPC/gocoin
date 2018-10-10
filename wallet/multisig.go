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

// File:        meltisig.go
// Description: Bictoin Cash Cash main Package

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

package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

const MultiToSignOut = "multi2sign.txt"

// add P2SH pre-signing data into a raw tx
func make_p2sh() {
	tx := raw_tx_from_file(*rawtx)
	if tx == nil {
		fmt.Println("ERROR: Cannot decode the raw transaction")
		return
	}

	d, er := hex.DecodeString(*p2sh)
	if er != nil {
		println("P2SH hex data:", er.Error())
		return
	}

	ms, er := bch.NewMultiSigFromP2SH(d)
	if er != nil {
		println("Decode P2SH:", er.Error())
		return
	}

	fmt.Println("The P2SH data points to address", ms.BtcAddr(testnet).String())

	sd := ms.Bytes()

	for i := range tx.TxIn {
		if *input < 0 || i == *input {
			tx.TxIn[i].ScriptSig = sd
			fmt.Println("Input number", i, " - hash to sign:", hex.EncodeToString(tx.SignatureHash(d, i, bch.SIGHASH_ALL)))
		}
	}
	ioutil.WriteFile(MultiToSignOut, []byte(hex.EncodeToString(tx.Serialize())), 0666)
	fmt.Println("Transaction with", len(tx.TxIn), "inputs ready for multi-signing, stored in", MultiToSignOut)
}

// reorder signatures to meet order of the keys
// remove signatuers made by the same keys
// remove exessive signatures (keeps transaction size down)
func multisig_reorder(tx *bch.Tx) (all_signed bool) {
	all_signed = true
	for i := range tx.TxIn {
		ms, _ := bch.NewMultiSigFromScript(tx.TxIn[i].ScriptSig)
		if ms == nil {
			continue
		}
		hash := tx.SignatureHash(ms.P2SH(), i, bch.SIGHASH_ALL)

		var sigs []*bch.Signature
		for ki := range ms.PublicKeys {
			var sig *bch.Signature
			for si := range ms.Signatures {
				if bch.EcdsaVerify(ms.PublicKeys[ki], ms.Signatures[si].Bytes(), hash) {
					//fmt.Println("Key number", ki, "has signature number", si)
					sig = ms.Signatures[si]
					break
				}
			}
			if sig != nil {
				sigs = append(sigs, sig)
			} else if *verbose {
				fmt.Println("WARNING: Key number", ki, "has no matching signature")
			}

			if !*allowextramsigns && uint(len(sigs)) >= ms.SigsNeeded {
				break
			}
		}

		if *verbose {
			if len(ms.Signatures) > len(sigs) {
				fmt.Println("WARNING: Some signatures are obsolete and will be removed", len(ms.Signatures), "=>", len(sigs))
			} else if len(ms.Signatures) < len(sigs) {
				fmt.Println("It appears that same key is re-used.", len(sigs)-len(ms.Signatures), "more signatures were added")
			}
		}

		ms.Signatures = sigs
		tx.TxIn[i].ScriptSig = ms.Bytes()

		if len(sigs) < int(ms.SigsNeeded) {
			all_signed = false
		}
	}
	return
}

// sign a multisig transaction with a specific key
func multisig_sign() {
	tx := raw_tx_from_file(*rawtx)
	if tx == nil {
		println("ERROR: Cannot decode the raw multisig transaction")
		println("Always use -msign <addr> along with -raw multi2sign.txt")
		return
	}

	k := address_to_key(*multisign)
	if k == nil {
		println("You do not know a key for address", *multisign)
		return
	}

	for i := range tx.TxIn {
		ms, er := bch.NewMultiSigFromScript(tx.TxIn[i].ScriptSig)
		if er != nil {
			println("WARNING: Input", i, "- not multisig:", er.Error())
			continue
		}
		hash := tx.SignatureHash(ms.P2SH(), i, bch.SIGHASH_ALL)
		//fmt.Println("Input number", i, len(ms.Signatures), " - hash to sign:", hex.EncodeToString(hash))

		r, s, e := bch.EcdsaSign(k.Key, hash)
		if e != nil {
			println(e.Error())
			return
		}
		btcsig := &bch.Signature{HashType: 0x01}
		btcsig.R.Set(r)
		btcsig.S.Set(s)

		ms.Signatures = append(ms.Signatures, btcsig)
		tx.TxIn[i].ScriptSig = ms.Bytes()
	}

	// Now re-order the signatures as they shall be:
	multisig_reorder(tx)

	write_tx_file(tx)
}
