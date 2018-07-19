package script

import (
	//"os"
	//"fmt"
	"errors"
	"testing"
	"strings"
	"io/ioutil"
	"encoding/hex"
	"encoding/json"
	"github.com/piotrnar/gocoin/lib/btc"
)

type one_test_vector struct {
	sigscr, pkscr []byte
	flags uint32
	exp_res bool
	desc string

	value uint64
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
		}

		credit_tx := mk_credit_tx(v.pkscr, v.value)
		spend_tx := mk_spend_tx(credit_tx, v.sigscr)

		if DBG_SCR {
			println("desc:", v, tot, v.desc)
			println("pkscr:", hex.EncodeToString(v.pkscr))
			println("sigscr:", hex.EncodeToString(v.sigscr))
			println("credit:", hex.EncodeToString(credit_tx.Serialize()))
			println("spend:", hex.EncodeToString(spend_tx.Serialize()))
			println("------------------------------ testing vector", tot, v.value)
		}
		res := VerifyTxScript(v.pkscr, v.value, 0, spend_tx, flags)

		if res!=v.exp_res {
			t.Error(tot, "TestScritps failed. Got:", res, "   exp:", v.exp_res, v.desc)
			return
		} else {
			if DBG_SCR {
				println(tot, "ok:", res, v.desc)
			}
		}

		if tot==114400 {
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
			case "MINIMALIF":
				fl |= VER_MINIMALIF
			case "NULLFAIL":
				fl |= VER_NULLFAIL
			default:
				e = errors.New("Unsupported flag "+ss[i])
				return
		}
	}
	return
}


func mk_credit_tx(pk_scr []byte, value uint64) (input_tx *btc.Tx) {
	// We build input_tx only to calculate it's hash for output_tx
	input_tx = new(btc.Tx)
	input_tx.Version = 1
	input_tx.TxIn = []*btc.TxIn{ &btc.TxIn{Input:btc.TxPrevOut{Vout:0xffffffff},
		ScriptSig:[]byte{0,0}, Sequence:0xffffffff} }
	input_tx.TxOut = []*btc.TxOut{ &btc.TxOut{Pk_script:pk_scr, Value:value} }
	// Lock_time = 0
	input_tx.SetHash(input_tx.Serialize())
	return
}

func mk_spend_tx(input_tx *btc.Tx, sig_scr []byte) (output_tx *btc.Tx) {
	output_tx = new(btc.Tx)
	output_tx.Version = 1
	output_tx.TxIn = []*btc.TxIn{ &btc.TxIn{Input:btc.TxPrevOut{Hash:btc.Sha2Sum(input_tx.Serialize()), Vout:0},
		ScriptSig:sig_scr, Sequence:0xffffffff} }
	output_tx.TxOut = []*btc.TxOut{ &btc.TxOut{Value:input_tx.TxOut[0].Value} }
	// Lock_time = 0

	output_tx.SetHash(output_tx.Serialize())
	return
}