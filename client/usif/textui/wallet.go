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

// File:		wallet.go
// Description:	Bictoin Cash textui Package

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

package textui

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	"github.com/counterpartyxcpc/gocoin-cash/client/wallet"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

type OneWalletAddrs struct {
	Typ int // 0-p2kh, 1-p2sh, 2-segwit_prog
	Key []byte
	rec *wallet.OneAllAddrBal
}

type SortedWalletAddrs []OneWalletAddrs

var sort_by_cnt bool

func (sk SortedWalletAddrs) Len() int {
	return len(sk)
}

func (sk SortedWalletAddrs) Less(a, b int) bool {
	if sort_by_cnt {
		return sk[a].rec.Count() > sk[b].rec.Count()
	}
	return sk[a].rec.Value > sk[b].rec.Value
}

func (sk SortedWalletAddrs) Swap(a, b int) {
	sk[a], sk[b] = sk[b], sk[a]
}

func max_outs(par string) {
	sort_by_cnt = true
	all_addrs(par)
}

func best_val(par string) {
	sort_by_cnt = false
	all_addrs(par)
}

func new_slice(in []byte) (kk []byte) {
	kk = make([]byte, len(in))
	copy(kk, in)
	return
}

func all_addrs(par string) {
	var ptkh_outs, ptkh_vals, ptsh_outs, ptsh_vals uint64
	var ptwkh_outs, ptwkh_vals, ptwsh_outs, ptwsh_vals uint64
	var best SortedWalletAddrs
	var cnt int = 15
	var mode int

	if par != "" {
		if c, e := strconv.ParseUint(par, 10, 32); e == nil {
			if c > 3 {
				cnt = int(c)
			} else {
				mode = int(c + 1)
				fmt.Println("Counting only addr type", ([]string{"P2KH", "P2SH", "P2WKH", "P2WSH"})[int(c)])
			}
		}
	}

	var MIN_BCH uint64 = 100e8
	var MIN_OUTS int = 1000

	if mode != 0 {
		MIN_BCH = 0
		MIN_OUTS = 0
	}

	if mode == 0 || mode == 1 {
		for k, rec := range wallet.AllBalancesP2KH {
			ptkh_vals += rec.Value
			ptkh_outs += uint64(rec.Count())
			if sort_by_cnt && rec.Count() >= MIN_OUTS || !sort_by_cnt && rec.Value >= MIN_BCH {
				best = append(best, OneWalletAddrs{Typ: 0, Key: new_slice(k[:]), rec: rec})
			}
		}
		fmt.Println(bch.UintToBtc(ptkh_vals), "BCH in", ptkh_outs, "unspent recs from", len(wallet.AllBalancesP2KH), "P2KH addresses")
	}

	if mode == 0 || mode == 2 {
		for k, rec := range wallet.AllBalancesP2SH {
			ptsh_vals += rec.Value
			ptsh_outs += uint64(rec.Count())
			if sort_by_cnt && rec.Count() >= MIN_OUTS || !sort_by_cnt && rec.Value >= MIN_BCH {
				best = append(best, OneWalletAddrs{Typ: 1, Key: new_slice(k[:]), rec: rec})
			}
		}
		fmt.Println(bch.UintToBtc(ptsh_vals), "BCH in", ptsh_outs, "unspent recs from", len(wallet.AllBalancesP2SH), "P2SH addresses")
	}

	if mode == 0 || mode == 3 {
		for k, rec := range wallet.AllBalancesP2WKH {
			ptwkh_vals += rec.Value
			ptwkh_outs += uint64(rec.Count())
			if sort_by_cnt && rec.Count() >= MIN_OUTS || !sort_by_cnt && rec.Value >= MIN_BCH {
				best = append(best, OneWalletAddrs{Typ: 2, Key: new_slice(k[:]), rec: rec})
			}
		}
		fmt.Println(bch.UintToBtc(ptwkh_vals), "BCH in", ptwkh_outs, "unspent recs from", len(wallet.AllBalancesP2WKH), "P2WKH addresses")
	}

	if mode == 0 || mode == 4 {
		for k, rec := range wallet.AllBalancesP2WSH {
			ptwsh_vals += rec.Value
			ptwsh_outs += uint64(rec.Count())
			if sort_by_cnt && rec.Count() >= MIN_OUTS || !sort_by_cnt && rec.Value >= MIN_BCH {
				best = append(best, OneWalletAddrs{Typ: 2, Key: new_slice(k[:]), rec: rec})
			}
		}
		fmt.Println(bch.UintToBtc(ptwsh_vals), "BCH in", ptwsh_outs, "unspent recs from", len(wallet.AllBalancesP2WSH), "P2WSH addresses")
	}

	if sort_by_cnt {
		fmt.Println("Top addresses with at least", MIN_OUTS, "unspent outputs:", len(best))
	} else {
		fmt.Println("Top addresses with at least", bch.UintToBtc(MIN_BCH), "BCH:", len(best))
	}

	sort.Sort(best)

	var pkscr_p2sk [23]byte
	var pkscr_p2kh [25]byte
	var ad *bch.BtcAddr

	pkscr_p2sk[0] = 0xa9
	pkscr_p2sk[1] = 20
	pkscr_p2sk[22] = 0x87

	pkscr_p2kh[0] = 0x76
	pkscr_p2kh[1] = 0xa9
	pkscr_p2kh[2] = 20
	pkscr_p2kh[23] = 0x88
	pkscr_p2kh[24] = 0xac

	for i := 0; i < len(best) && i < cnt; i++ {
		switch best[i].Typ {
		case 0:
			copy(pkscr_p2kh[3:23], best[i].Key)
			ad = bch.NewAddrFromPkScript(pkscr_p2kh[:], common.CFG.Testnet)
		case 1:
			copy(pkscr_p2sk[2:22], best[i].Key)
			ad = bch.NewAddrFromPkScript(pkscr_p2sk[:], common.CFG.Testnet)
		case 2:
			ad = new(bch.BtcAddr)
			ad.SegwitProg = new(bch.SegwitProg)
			ad.SegwitProg.HRP = bch.GetSegwitHRP(common.CFG.Testnet)
			ad.SegwitProg.Program = best[i].Key
		}
		fmt.Println(i+1, ad.String(), bch.UintToBtc(best[i].rec.Value), "BCH in", best[i].rec.Count(), "inputs")
	}
}

func list_unspent(addr string) {
	fmt.Println("Checking unspent coins for addr", addr)

	ad, e := bch.NewAddrFromString(addr)
	if e != nil {
		println(e.Error())
		return
	}

	outscr := ad.OutScript()

	unsp := wallet.GetAllUnspent(ad)
	if len(unsp) == 0 {
		fmt.Println(ad.String(), "has no coins")
	} else {
		var tot uint64
		sort.Sort(unsp)
		for i := range unsp {
			unsp[i].BtcAddr = nil // no need to print the address here
			tot += unsp[i].Value
		}
		fmt.Println(ad.String(), "has", bch.UintToBtc(tot), "BCH in", len(unsp), "records:")
		for i := range unsp {
			fmt.Println(unsp[i].String())
			network.TxMutex.Lock()
			bidx, spending := network.SpentOutputs[unsp[i].TxPrevOut.UIdx()]
			var t2s *network.OneTxToSend
			if spending {
				t2s, spending = network.TransactionsToSend[bidx]
			}
			network.TxMutex.Unlock()
			if spending {
				fmt.Println("\t- being spent by TxID", t2s.Hash.String())
			}
		}
	}

	network.TxMutex.Lock()
	for _, t2s := range network.TransactionsToSend {
		for vo, to := range t2s.TxOut {
			if bytes.Equal(to.Pk_script, outscr) {
				fmt.Println(fmt.Sprintf("Mempool Tx: %15s BCH comming with %s-%03d",
					bch.UintToBtc(to.Value), t2s.Hash.String(), vo))
			}
		}
	}
	network.TxMutex.Unlock()
}

func all_val_stats(s string) {
	wallet.PrintStat()
}

func wallet_on_off(s string) {
	if s == "on" {
		select {
		case wallet.OnOff <- true:
		default:
		}
		return
	} else if s == "off" {
		select {
		case wallet.OnOff <- false:
		default:
		}
		return
	}

	if common.GetBool(&common.WalletON) {
		fmt.Println("Wallet functionality is currently ENABLED. Execute 'wallet off' to disable it.")
		fmt.Println("")
	} else {
		if perc := common.GetUint32(&common.WalletProgress); perc != 0 {
			fmt.Println("Enabling wallet functionality -", (perc-1)/10, "percent complete. Execute 'wallet off' to abort it.")
		} else {
			fmt.Println("Wallet functionality is currently DISABLED. Execute 'wallet on' to enable it.")
		}
	}

	if pend := common.GetUint32(&common.WalletOnIn); pend > 0 {
		fmt.Println("Wallet functionality will auto enable in", pend, "seconds")
	}
}

func init() {
	newUi("richest r", true, best_val, "Show addresses with most coins [0,1,2,3 or count]")
	newUi("maxouts o", true, max_outs, "Show addresses with highest number of outputs [0,1,2,3 or count]")
	newUi("balance a", true, list_unspent, "List balance of given bitcoin address")
	newUi("allbal ab", true, all_val_stats, "Show Allbalance statistics")
	newUi("wallet w", false, wallet_on_off, "Enable (on) or disable (off) wallet functionality")
}
