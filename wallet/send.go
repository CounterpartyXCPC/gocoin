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

// File:        send.go
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
	"os"
	"strings"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

// Resolved while parsing "-send" parameter
type oneSendTo struct {
	addr   *bch.BtcAddr
	amount uint64
}

var (
	// set in parse_spend():
	spendBtc, feeBtc, changeBtc uint64
	sendTo                      []oneSendTo
)

// parse the "-send ..." parameter
func parse_spend() {
	outs := strings.Split(*send, ",")

	for i := range outs {
		tmp := strings.Split(strings.Trim(outs[i], " "), "=")
		if len(tmp) != 2 {
			println("The outputs must be in a format address1=amount1[,addressN=amountN]")
			cleanExit(1)
		}

		a, e := bch.NewAddrFromString(tmp[0])
		if e != nil {
			println("NewAddrFromString:", e.Error())
			cleanExit(1)
		}
		assert_address_version(a)

		am, er := bch.StringToSatoshis(tmp[1])
		if er != nil {
			println("Incorrect amount: ", tmp[1], er.Error())
			cleanExit(1)
		}
		if *subfee && i == 0 {
			am -= curFee
		}

		sendTo = append(sendTo, oneSendTo{addr: a, amount: am})
		spendBtc += am
	}
}

// parse the "-batch ..." parameter
func parse_batch() {
	f, e := os.Open(*batch)
	if e == nil {
		defer f.Close()
		td := bufio.NewReader(f)
		var lcnt int
		for {
			li, _, _ := td.ReadLine()
			if li == nil {
				break
			}
			lcnt++
			tmp := strings.SplitN(strings.Trim(string(li), " "), "=", 2)
			if len(tmp) < 2 {
				println("Error in the batch file line", lcnt)
				cleanExit(1)
			}
			if tmp[0][0] == '#' {
				continue // Just a comment-line
			}

			a, e := bch.NewAddrFromString(tmp[0])
			if e != nil {
				println("NewAddrFromString:", e.Error())
				cleanExit(1)
			}
			assert_address_version(a)

			am, e := bch.StringToSatoshis(tmp[1])
			if e != nil {
				println("StringToSatoshis:", e.Error())
				cleanExit(1)
			}

			sendTo = append(sendTo, oneSendTo{addr: a, amount: am})
			spendBtc += am
		}
	} else {
		println(e.Error())
		cleanExit(1)
	}
}

// returns true if spend operation has been requested
func send_request() bool {
	feeBtc = curFee
	if *send != "" {
		parse_spend()
	}
	if *batch != "" {
		parse_batch()
	}
	return len(sendTo) > 0
}
