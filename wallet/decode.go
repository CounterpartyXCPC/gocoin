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

// File:        decode.go
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

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

// hex dump with max 32 bytes per line
func hex_dump(d []byte) (s string) {
	for {
		le := 32
		if len(d) < le {
			le = len(d)
		}
		s += "       " + hex.EncodeToString(d[:le]) + "\n"
		d = d[le:]
		if len(d) == 0 {
			return
		}
	}
}

func dump_raw_sigscript(d []byte) bool {
	ss, er := bch.ScriptToText(d)
	if er != nil {
		println(er.Error())
		return false
	}

	p2sh := len(ss) >= 2 && d[0] == 0
	if p2sh {
		ms, er := bch.NewMultiSigFromScript(d)
		if er == nil {
			fmt.Println("      Multisig script", ms.SigsNeeded, "of", len(ms.PublicKeys))
			for i := range ms.PublicKeys {
				fmt.Printf("       pkey%d = %s\n", i+1, hex.EncodeToString(ms.PublicKeys[i]))
			}
			for i := range ms.Signatures {
				fmt.Printf("       R%d = %64s\n", i+1, hex.EncodeToString(ms.Signatures[i].R.Bytes()))
				fmt.Printf("       S%d = %64s\n", i+1, hex.EncodeToString(ms.Signatures[i].S.Bytes()))
				fmt.Printf("       HashType%d = %02x\n", i+1, ms.Signatures[i].HashType)
			}
			return len(ms.Signatures) >= int(ms.SigsNeeded)
		} else {
			println(er.Error())
		}
	}

	fmt.Println("      SigScript:")
	for i := range ss {
		if p2sh && i == len(ss)-1 {
			// Print p2sh script
			d, _ = hex.DecodeString(ss[i])
			s2, er := bch.ScriptToText(d)
			if er != nil {
				println(er.Error())
				p2sh = false
				fmt.Println("       ", ss[i])
				continue
				//return
			}
			fmt.Println("        P2SH spend script:")
			for j := range s2 {
				fmt.Println("        ", s2[j])
			}
		} else {
			fmt.Println("       ", ss[i])
		}
	}
	return true
}

func dump_sigscript(d []byte) bool {
	if len(d) == 0 {
		fmt.Println("       WARNING: Empty sigScript")
		return false
	}
	rd := bytes.NewReader(d)

	// ECDSA Signature
	le, _ := rd.ReadByte()
	if le < 0x40 {
		return dump_raw_sigscript(d)
	}
	sd := make([]byte, le)
	_, er := rd.Read(sd)
	if er != nil {
		return dump_raw_sigscript(d)
	}
	sig, er := bch.NewSignature(sd)
	if er != nil {
		return dump_raw_sigscript(d)
	}
	fmt.Printf("       R = %64s\n", hex.EncodeToString(sig.R.Bytes()))
	fmt.Printf("       S = %64s\n", hex.EncodeToString(sig.S.Bytes()))
	fmt.Printf("       HashType = %02x\n", sig.HashType)

	// Key
	le, er = rd.ReadByte()
	if er != nil {
		fmt.Println("       WARNING: PublicKey not present")
		fmt.Print(hex_dump(d))
		return false
	}

	sd = make([]byte, le)
	_, er = rd.Read(sd)
	if er != nil {
		fmt.Println("       WARNING: PublicKey too short", er.Error())
		fmt.Print(hex_dump(d))
		return false
	}

	fmt.Printf("       PublicKeyType = %02x\n", sd[0])
	key, er := bch.NewPublicKey(sd)
	if er != nil {
		fmt.Println("       WARNING: PublicKey broken", er.Error())
		fmt.Print(hex_dump(d))
		return false
	}
	fmt.Printf("       X = %64s\n", key.X.String())
	if le >= 65 {
		fmt.Printf("       Y = %64s\n", key.Y.String())
	}

	if rd.Len() != 0 {
		fmt.Println("       WARNING: Extra bytes at the end of sigScript")
		fmt.Print(hex_dump(d[len(d)-rd.Len():]))
	}
	return true
}

// dump raw transaction
func dump_raw_tx() {
	tx := raw_tx_from_file(*dumptxfn)
	if tx == nil {
		fmt.Println("ERROR: Cannot decode the raw transaction")
		return
	}

	var unsigned, totin, totout, noins uint64

	fmt.Println("ID:", tx.Hash.String())
	fmt.Println("WTxID:", tx.WTxID().String())
	fmt.Println("Tx Version:", tx.Version)
	if tx.SegWit != nil {
		fmt.Println("Segregated Witness transaction", len(tx.SegWit))
	} else {
		fmt.Println("Regular (non-SegWit) transaction", len(tx.SegWit))
	}
	if tx.IsCoinBase() {
		if len(tx.TxIn[0].ScriptSig) >= 4 && tx.TxIn[0].ScriptSig[0] == 3 {
			fmt.Println("Coinbase TX from block height", uint(tx.TxIn[0].ScriptSig[1])|
				uint(tx.TxIn[0].ScriptSig[2])<<8|uint(tx.TxIn[0].ScriptSig[3])<<16)
		} else {
			fmt.Println("Coinbase TX from an unknown block")
		}
		s := hex.EncodeToString(tx.TxIn[0].ScriptSig)
		for len(s) > 0 {
			i := len(s)
			if i > 64 {
				i = 64
			}
			fmt.Println("  ", s[:i])
			s = s[i:]
		}
		for wia := range tx.SegWit {
			for wib, ww := range tx.SegWit[wia] {
				fmt.Println("  Witness", wia, wib, hex.EncodeToString(ww))
			}
		}
		//fmt.Println()
	} else {
		fmt.Println("TX IN cnt:", len(tx.TxIn))
		for i := range tx.TxIn {
			fmt.Printf("%4d) %s sl=%d seq=%08x\n", i, tx.TxIn[i].Input.String(),
				len(tx.TxIn[i].ScriptSig), tx.TxIn[i].Sequence)

			if intx := tx_from_balance(bch.NewUint256(tx.TxIn[i].Input.Hash[:]), false); intx != nil {
				val := intx.TxOut[tx.TxIn[i].Input.Vout].Value
				totin += val
				fmt.Printf("%15s BCH\n", bch.UintToBtc(val))
			} else {
				noins++
			}

			if len(tx.TxIn[i].ScriptSig) > 0 {
				if !dump_sigscript(tx.TxIn[i].ScriptSig) {
					unsigned++
				}
			} else {
				if tx.SegWit == nil || len(tx.SegWit[i]) < 2 {
					unsigned++
				}
			}

			if tx.SegWit != nil {
				fmt.Println("      Witness data:")
				for _, ww := range tx.SegWit[i] {
					if len(ww) == 0 {
						fmt.Println("       ", "OP_0")
					} else {
						fmt.Println("       ", hex.EncodeToString(ww))
					}
				}
			}
		}
	}
	fmt.Println("TX OUT cnt:", len(tx.TxOut))
	for i := range tx.TxOut {
		totout += tx.TxOut[i].Value
		fmt.Printf("%4d) %20s BTC ", i, bch.UintToBtc(tx.TxOut[i].Value))
		addr := addr_from_pkscr(tx.TxOut[i].Pk_script)
		if addr != nil {
			if addr.Version == ver_script() {
				fmt.Println("to scriptH", addr.String())
			} else {
				fmt.Println("to address", addr.String())
			}
		} else {
			if tx.TxOut[i].Value > 0 {
				fmt.Println("WARNING!!! These coins go to non-standard Pk_script:")
			} else {
				fmt.Println("NULL output to Pk_script:")
			}
			ss, er := bch.ScriptToText(tx.TxOut[i].Pk_script)
			if er == nil {
				for i := range ss {
					fmt.Println("       ", ss[i])
				}
			} else {
				fmt.Println(hex.EncodeToString(tx.TxOut[i].Pk_script))
				fmt.Println(er.Error())
			}
		}
	}
	fmt.Println("Lock Time:", tx.Lock_time)

	fmt.Println("Output volume:", bch.UintToBtc(totout), "BCH")
	if noins == 0 {
		fmt.Println("Input volume :", bch.UintToBtc(totin), "BCH")
		fmt.Println("Transact. fee:", bch.UintToBtc(totin-totout), "BCH")
	} else {
		fmt.Println("WARNING: Unable to figure out what the fee is")
	}
	fmt.Println("Transaction Size:", tx.Size, "   NoWitSize:", tx.NoWitSize, "   VSize:", tx.VSize())

	if !tx.IsCoinBase() {
		if unsigned > 0 {
			fmt.Println("WARNING:", unsigned, "out of", len(tx.TxIn), "inputs are not signed or signed only patially")
		} else {
			fmt.Println("All", len(tx.TxIn), "transaction inputs seem to be signed")
		}
	}
}
