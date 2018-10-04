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

// File:        script_test.go
// Description: Bictoin Cash Cash script Package

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

package script

import (
	//"os"
	//"fmt"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

type one_test_vector struct {
	sigscr, pkscr []byte
	flags         uint32
	exp_res       bool
	desc          string

	witness [][]byte
	value   uint64
}

func TestScritps(t *testing.T) {
	var str interface{}
	var vecs []*one_test_vector

	DBG_ERR = false
	dat, er := ioutil.ReadFile("../test/script_tests.json")
	if er != nil {
		t.Error(er.Error())
		return
	}
	er = json.Unmarshal(dat, &str)
	if er != nil {
		t.Error(er.Error())
		return
	}

	m := str.([]interface{})
	for i := range m {
		switch mm := m[i].(type) {
		case []interface{}:
			if len(mm) < 4 {
				continue
			}

			var skip bool
			var bfield int
			var e error
			var all_good bool

			vec := new(one_test_vector)
			for ii := range mm {
				switch segwitdata := mm[ii].(type) {
				case []interface{}:
					for iii := range segwitdata {
						switch segwitdata[iii].(type) {
						case string:
							var by []byte
							s := segwitdata[iii].(string)
							by, e = hex.DecodeString(s)
							if e != nil {
								t.Error("error parsing serwit script", s)
								skip = true
								break
							}
							vec.witness = append(vec.witness, by)

						case float64:
							vec.value = uint64(1e8 * segwitdata[iii].(float64))
						}
					}

				case string:
					s := mm[ii].(string)
					if bfield == 0 {
						vec.sigscr, e = bch.DecodeScript(s)
						if e != nil {
							t.Error("error parsing script", s)
							skip = true
							break
						}
					} else if bfield == 1 {
						vec.pkscr, e = bch.DecodeScript(s)
						if e != nil {
							skip = true
							break
						}
					} else if bfield == 2 {
						vec.flags, e = decode_flags(s)
						if e != nil {
							println("error parsing flag", e.Error())
							skip = true
							break
						}
					} else if bfield == 3 {
						vec.exp_res = s == "OK"
						all_good = true
					} else if bfield == 4 {
						vec.desc = s
						skip = true
						break
					}
					bfield++

				default:
					panic("Enexpected test vector")
					skip = true
				}
				if skip {
					break
				}
			}
			if all_good {
				vecs = append(vecs, vec)
			}
		}
	}

	tot := 0
	for _, v := range vecs {
		tot++

		/*
			if tot==114400 {
				DBG_SCR = true
				DBG_ERR = true
			}*/

		flags := v.flags
		if (flags & VER_CLEANSTACK) != 0 {
			flags |= VER_P2SH
			flags |= VER_WITNESS
		}

		credit_tx := mk_credit_tx(v.pkscr, v.value)
		spend_tx := mk_spend_tx(credit_tx, v.sigscr, v.witness)

		if DBG_SCR {
			println("desc:", v, tot, v.desc)
			println("pkscr:", hex.EncodeToString(v.pkscr))
			println("sigscr:", hex.EncodeToString(v.sigscr))
			println("credit:", hex.EncodeToString(credit_tx.Serialize()))
			println("spend:", hex.EncodeToString(spend_tx.Serialize()))
			println("------------------------------ testing vector", tot, len(v.witness), v.value)
		}
		res := VerifyTxScript(v.pkscr, v.value, 0, spend_tx, flags)

		if res != v.exp_res {
			t.Error(tot, "TestScritps failed. Got:", res, "   exp:", v.exp_res, v.desc)
			return
		} else {
			if DBG_SCR {
				println(tot, "ok:", res, v.desc)
			}
		}

		if tot == 114400 {
			return
		}
	}
}

func decode_flags(s string) (fl uint32, e error) {
	ss := strings.Split(s, ",")
	for i := range ss {
		switch ss[i] {
		case "": // ignore
		case "NONE": // ignore
			break
		case "P2SH":
			fl |= VER_P2SH
		case "STRICTENC":
			fl |= VER_STRICTENC
		case "DERSIG":
			fl |= VER_DERSIG
		case "LOW_S":
			fl |= VER_LOW_S
		case "NULLDUMMY":
			fl |= VER_NULLDUMMY
		case "SIGPUSHONLY":
			fl |= VER_SIGPUSHONLY
		case "MINIMALDATA":
			fl |= VER_MINDATA
		case "DISCOURAGE_UPGRADABLE_NOPS":
			fl |= VER_BLOCK_OPS
		case "CLEANSTACK":
			fl |= VER_CLEANSTACK
		case "CHECKLOCKTIMEVERIFY":
			fl |= VER_CLTV
		case "CHECKSEQUENCEVERIFY":
			fl |= VER_CSV
		case "WITNESS":
			fl |= VER_WITNESS
		case "DISCOURAGE_UPGRADABLE_WITNESS_PROGRAM":
			fl |= VER_WITNESS_PROG
		case "MINIMALIF":
			fl |= VER_MINIMALIF
		case "NULLFAIL":
			fl |= VER_NULLFAIL
		case "WITNESS_PUBKEYTYPE":
			fl |= VER_WITNESS_PUBKEY
		default:
			e = errors.New("Unsupported flag " + ss[i])
			return
		}
	}
	return
}

func mk_credit_tx(pk_scr []byte, value uint64) (input_tx *bch.Tx) {
	// We build input_tx only to calculate it's hash for output_tx
	input_tx = new(bch.Tx)
	input_tx.Version = 1
	input_tx.TxIn = []*bch.TxIn{{Input: bch.TxPrevOut{Vout: 0xffffffff},
		ScriptSig: []byte{0, 0}, Sequence: 0xffffffff}}
	input_tx.TxOut = []*bch.TxOut{{Pk_script: pk_scr, Value: value}}
	// Lock_time = 0
	input_tx.SetHash(input_tx.Serialize())
	return
}

func mk_spend_tx(input_tx *bch.Tx, sig_scr []byte, witness [][]byte) (output_tx *bch.Tx) {
	output_tx = new(bch.Tx)
	output_tx.Version = 1
	output_tx.TxIn = []*bch.TxIn{{Input: bch.TxPrevOut{Hash: bch.Sha2Sum(input_tx.Serialize()), Vout: 0},
		ScriptSig: sig_scr, Sequence: 0xffffffff}}
	output_tx.TxOut = []*bch.TxOut{{Value: input_tx.TxOut[0].Value}}
	// Lock_time = 0

	if len(witness) > 0 {
		output_tx.SegWit = make([][][]byte, 1)
		output_tx.SegWit[0] = witness
		if DBG_SCR {
			println("tx has", len(witness), "ws")
			for xx := range witness {
				println("", xx, hex.EncodeToString(witness[xx]))
			}
		}
	}
	output_tx.SetHash(output_tx.Serialize())
	return
}
