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

// File:        signtx.go
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
	"bytes"
	"encoding/hex"
	"fmt"
	"os"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

// prepare a signed transaction
func sign_tx(tx *bch.Tx) (all_signed bool) {
	var multisig_done bool
	all_signed = true

	// go through each input
	for in := range tx.TxIn {
		if ms, _ := bch.NewMultiSigFromScript(tx.TxIn[in].ScriptSig); ms != nil {
			hash := tx.SignatureHash(ms.P2SH(), in, bch.SIGHASH_ALL)
			for ki := range ms.PublicKeys {
				k := public_to_key(ms.PublicKeys[ki])
				if k != nil {
					r, s, e := bch.EcdsaSign(k.Key, hash)
					if e != nil {
						println("ERROR in sign_tx:", e.Error())
						all_signed = false
					} else {
						btcsig := &bch.Signature{HashType: 0x01}
						btcsig.R.Set(r)
						btcsig.S.Set(s)

						ms.Signatures = append(ms.Signatures, btcsig)
						tx.TxIn[in].ScriptSig = ms.Bytes()
						multisig_done = true
					}
				}
			}
		} else {
			uo := getUO(&tx.TxIn[in].Input)
			if uo == nil {
				println("ERROR: Unkown input:", tx.TxIn[in].Input.String(), "- missing balance folder?")
				all_signed = false
				continue
			}
			adr := addr_from_pkscr(uo.Pk_script)
			if adr == nil {
				fmt.Println("WARNING: Don't know how to sign input number", in)
				fmt.Println(" Pk_script:", hex.EncodeToString(uo.Pk_script))
				all_signed = false
				continue
			}

			ver, segwit_prog := bch.IsWitnessProgram(uo.Pk_script)
			if len(segwit_prog) == 20 && ver == 0 {
				copy(adr.Hash160[:], segwit_prog) // native segwith P2WPKH output
			}

			k_idx := hash_to_key_idx(adr.Hash160[:])
			if k_idx < 0 {
				fmt.Println("WARNING: You do not have key for", adr.String(), "at input", in)
				all_signed = false
				continue
			}
			var er error
			k := keys[k_idx]
			if segwit_prog != nil {
				er = tx.SignWitness(in, k.BtcAddr.OutScript(), uo.Value, bch.SIGHASH_ALL, k.BtcAddr.Pubkey, k.Key)
			} else if adr.String() == segwit[k_idx].String() {
				tx.TxIn[in].ScriptSig = append([]byte{22, 0, 20}, k.BtcAddr.Hash160[:]...)
				er = tx.SignWitness(in, k.BtcAddr.OutScript(), uo.Value, bch.SIGHASH_ALL, k.BtcAddr.Pubkey, k.Key)
			} else {
				er = tx.Sign(in, uo.Pk_script, bch.SIGHASH_ALL, k.BtcAddr.Pubkey, k.Key)
			}
			if er != nil {
				fmt.Println("ERROR: Sign failed for input number", in, er.Error())
				all_signed = false
			}
		}
	}

	// reorder signatures if we signed any multisig inputs
	if multisig_done && !multisig_reorder(tx) {
		all_signed = false
	}

	if !all_signed {
		fmt.Println("WARNING: Not all the inputs have been signed")
	}

	return
}

func write_tx_file(tx *bch.Tx) {
	var signedrawtx []byte
	if tx.SegWit != nil {
		signedrawtx = tx.SerializeNew()
	} else {
		signedrawtx = tx.Serialize()
	}
	tx.SetHash(signedrawtx)

	hs := tx.Hash.String()
	fmt.Println("TxID", hs)

	var fn string

	if txfilename == "" {
		fn = hs[:8] + ".txt"
	} else {
		fn = txfilename
	}

	f, _ := os.Create(fn)
	if f != nil {
		f.Write([]byte(hex.EncodeToString(signedrawtx)))
		f.Close()
		fmt.Println("Transaction data stored in", fn)
	}
}

// prepare a signed transaction
func make_signed_tx() {
	// Make an empty transaction
	tx := new(bch.Tx)
	tx.Version = 1
	tx.Lock_time = 0

	// Select as many inputs as we need to pay the full amount (with the fee)
	var btcsofar uint64
	for i := range unspentOuts {
		if unspentOuts[i].key == nil {
			continue
		}
		uo := getUO(&unspentOuts[i].TxPrevOut)
		// add the input to our transaction:
		tin := new(bch.TxIn)
		tin.Input = unspentOuts[i].TxPrevOut
		tin.Sequence = uint32(*sequence)
		tx.TxIn = append(tx.TxIn, tin)

		btcsofar += uo.Value
		unspentOuts[i].spent = true
		if !*useallinputs && (btcsofar >= spendBtc+feeBtc) {
			break
		}
	}
	if btcsofar < (spendBtc + feeBtc) {
		fmt.Println("ERROR: You have", bch.UintToBtc(btcsofar), "BCH, but you need",
			bch.UintToBtc(spendBtc+feeBtc), "BCH for the transaction")
		cleanExit(1)
	}
	changeBtc = btcsofar - (spendBtc + feeBtc)
	if *verbose {
		fmt.Printf("Spending %d out of %d outputs...\n", len(tx.TxIn), len(unspentOuts))
	}

	// Build transaction outputs:
	for o := range sendTo {
		outs, er := bch.NewSpendOutputs(sendTo[o].addr, sendTo[o].amount, testnet)
		if er != nil {
			fmt.Println("ERROR:", er.Error())
			cleanExit(1)
		}
		tx.TxOut = append(tx.TxOut, outs...)
	}

	if changeBtc > 0 {
		// Add one more output (with the change)
		chad := get_change_addr()
		if *verbose {
			fmt.Println("Sending change", changeBtc, "to", chad.String())
		}
		outs, er := bch.NewSpendOutputs(chad, changeBtc, testnet)
		if er != nil {
			fmt.Println("ERROR:", er.Error())
			cleanExit(1)
		}
		tx.TxOut = append(tx.TxOut, outs...)
	}

	if *message != "" {
		// Add NULL output with an arbitrary message
		scr := new(bytes.Buffer)
		scr.WriteByte(0x6a) // OP_RETURN
		bch.WritePutLen(scr, uint32(len(*message)))
		scr.Write([]byte(*message))
		tx.TxOut = append(tx.TxOut, &bch.TxOut{Value: 0, Pk_script: scr.Bytes()})
	}

	signed := sign_tx(tx)
	write_tx_file(tx)

	if apply2bal && signed {
		apply_to_balance(tx)
	}
}

// sign raw transaction with all the keys we have
func process_raw_tx() {
	tx := raw_tx_from_file(*rawtx)
	if tx == nil {
		fmt.Println("ERROR: Cannot decode the raw transaction")
		return
	}

	sign_tx(tx)
	write_tx_file(tx)
}
