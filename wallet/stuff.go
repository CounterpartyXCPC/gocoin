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

// File:        stuff.go
// Description: Bictoin Cash Cash main Package

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

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/ltc"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

// Cache for txs from already loaded from balance/ folder
var loadedTxs map[[32]byte]*bch.Tx = make(map[[32]byte]*bch.Tx)

// Read a line from stdin
func getline() string {
	li, _, _ := bufio.NewReader(os.Stdin).ReadLine()
	return string(li)
}

func ask_yes_no(msg string) bool {
	for {
		fmt.Print(msg, " (y/n) : ")
		l := strings.ToLower(getline())
		if l == "y" {
			return true
		} else if l == "n" {
			return false
		}
	}
	return false
}

func getpass() []byte {
	var pass [1024]byte
	var n int
	var e error
	var f *os.File

	if stdin {
		if *ask4pass {
			fmt.Println("ERROR: Both -p and -stdin switches are not allowed at the same time")
			return nil
		}
		d, er := ioutil.ReadAll(os.Stdin)
		if er != nil {
			fmt.Println("Reading from stdin:", e.Error())
			return nil
		}
		n = len(d)
		if n <= 0 {
			return nil
		}
		copy(pass[:n], d)
		sys.ClearBuffer(d)
		goto check_pass
	}

	if !*ask4pass {
		f, e = os.Open(PassSeedFilename)
		if e == nil {
			n, e = f.Read(pass[:])
			f.Close()
			if n <= 0 {
				return nil
			}
			goto check_pass
		}

		fmt.Println("Seed file", PassSeedFilename, "not found")
	}

	fmt.Print("Enter your wallet's seed password: ")
	n = sys.ReadPassword(pass[:])
	if n <= 0 {
		return nil
	}

	if *list {
		if !*singleask {
			fmt.Print("Re-enter the seed password (to be sure): ")
			var pass2 [1024]byte
			p2len := sys.ReadPassword(pass2[:])
			if p2len != n || !bytes.Equal(pass[:n], pass2[:p2len]) {
				sys.ClearBuffer(pass[:n])
				sys.ClearBuffer(pass2[:p2len])
				println("The two passwords you entered do not match")
				return nil
			}
			sys.ClearBuffer(pass2[:p2len])
		}
		if *list {
			// Maybe he wants to save the password?
			if ask_yes_no("Save the password on disk, so you won't be asked for it later?") {
				e = ioutil.WriteFile(PassSeedFilename, pass[:n], 0600)
				if e != nil {
					fmt.Println("WARNING: Could not save the password", e.Error())
				} else {
					fmt.Println("The seed password has been stored in", PassSeedFilename)
				}
			}
		}
	}
check_pass:
	for i := 0; i < n; i++ {
		if pass[i] < ' ' || pass[i] > 126 {
			fmt.Println("WARNING: Your secret contains non-printable characters")
			break
		}
	}
	outpass := make([]byte, n+len(secret_seed))
	if len(secret_seed) > 0 {
		copy(outpass, secret_seed)
	}
	copy(outpass[len(secret_seed):], pass[:n])
	sys.ClearBuffer(pass[:n])
	return outpass
}

// return the change addrress or nil if there will be no change
func get_change_addr() (chng *bch.BtcAddr) {
	if *change != "" {
		var e error
		chng, e = bch.NewAddrFromString(*change)
		if e != nil {
			println("Change address:", e.Error())
			cleanExit(1)
		}
		assert_address_version(chng)
		return
	}

	// If change address not specified, send it back to the first input
	for idx := range unspentOuts {
		uo := getUO(&unspentOuts[idx].TxPrevOut)
		if k := pkscr_to_key(uo.Pk_script); k != nil {
			chng = k.BtcAddr
			return
		}
	}

	fmt.Println("ERROR: Could not determine address where to send change. Add -change switch")
	cleanExit(1)
	return
}

func raw_tx_from_file(fn string) *bch.Tx {
	dat := sys.GetRawData(fn)
	if dat == nil {
		fmt.Println("Cannot fetch raw transaction data")
		return nil
	}
	tx, txle := bch.NewTx(dat)
	if tx != nil {
		tx.SetHash(dat)
		if txle != len(dat) {
			fmt.Println("WARNING: Raw transaction length mismatch", txle, len(dat))
		}
	}
	return tx
}

// Get tx with given id from the balance folder, of from cache
func tx_from_balance(txid *bch.Uint256, error_is_fatal bool) (tx *bch.Tx) {
	if tx = loadedTxs[txid.Hash]; tx != nil {
		return // we have it in cache already
	}
	fn := "balance/" + txid.String() + ".tx"
	buf, er := ioutil.ReadFile(fn)
	if er == nil && buf != nil {
		var th [32]byte
		bch.ShaHash(buf, th[:])
		if txid.Hash == th {
			tx, _ = bch.NewTx(buf)
			if error_is_fatal && tx == nil {
				println("Transaction is corrupt:", txid.String())
				cleanExit(1)
			}
		} else if error_is_fatal {
			println("Transaction file is corrupt:", txid.String())
			cleanExit(1)
		}
	} else if error_is_fatal {
		println("Error reading transaction file:", fn)
		if er != nil {
			println(er.Error())
		}
		cleanExit(1)
	}
	loadedTxs[txid.Hash] = tx // store it in the cache
	return
}

// Look for specific TxPrevOut in the balance folder
func getUO(pto *bch.TxPrevOut) *bch.TxOut {
	if _, ok := loadedTxs[pto.Hash]; !ok {
		loadedTxs[pto.Hash] = tx_from_balance(bch.NewUint256(pto.Hash[:]), true)
	}
	return loadedTxs[pto.Hash].TxOut[pto.Vout]
}

// version byte for P2KH addresses
func ver_pubkey() byte {
	if litecoin {
		return ltc.AddrVerPubkey(testnet)
	} else {
		return bch.AddrVerPubkey(testnet)
	}
}

// version byte for P2SH addresses
func ver_script() byte {
	if litecoin {
		return ltc.AddrVerScript(testnet)
	} else {
		return bch.AddrVerScript(testnet)
	}
}

// version byte for private key addresses
func ver_secret() byte {
	return ver_pubkey() + 0x80
}

// get BtcAddr from pk_script
func addr_from_pkscr(scr []byte) *bch.BtcAddr {
	if litecoin {
		return ltc.NewAddrFromPkScript(scr, testnet)
	} else {
		return bch.NewAddrFromPkScript(scr, testnet)
	}
}

// make sure the version byte in the given address is what we expect
func assert_address_version(a *bch.BtcAddr) {
	if a.SegwitProg != nil {
		if a.SegwitProg.HRP != bch.GetSegwitHRP(testnet) {
			println("Sending address", a.String(), "has an incorrect HRP string", a.SegwitProg.HRP)
			cleanExit(1)
		}
	} else if a.Version != ver_pubkey() && a.Version != ver_script() {
		println("Sending address", a.String(), "has an incorrect version", a.Version)
		cleanExit(1)
	}
}
