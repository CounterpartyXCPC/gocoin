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

// File:		counts.go
// Description:	Bictoin Cash webui Package

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

package webui

import (
	//	"os"
	"fmt"
	"sort"

	//	"strings"
	"net/http"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
)

type many_counters []one_counter

type one_counter struct {
	key string
	cnt uint64
}

func (c many_counters) Len() int {
	return len(c)
}

func (c many_counters) Less(i, j int) bool {
	return c[i].key < c[j].key
}

func (c many_counters) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func p_counts(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}
	s := load_template("counts.html")
	write_html_head(w, r)
	w.Write([]byte(s))
	write_html_tail(w)
}

func json_counts(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}
	var net []string
	var gen, txs many_counters
	common.CounterMutex.Lock()
	for k, v := range common.Counter {
		if k[4] == '_' {
			var i int
			for i = 0; i < len(net); i++ {
				if net[i] == k[5:] {
					break
				}
			}
			if i == len(net) {
				net = append(net, k[5:])
			}
		} else if k[:2] == "Tx" {
			txs = append(txs, one_counter{key: k[2:], cnt: v})
		} else {
			gen = append(gen, one_counter{key: k, cnt: v})
		}
	}
	common.CounterMutex.Unlock()
	sort.Sort(gen)
	sort.Sort(txs)
	sort.Strings(net)

	w.Header()["Content-Type"] = []string{"application/json"}
	w.Write([]byte("{\n"))

	w.Write([]byte(" \"gen\":["))
	for i := range gen {
		w.Write([]byte(fmt.Sprint("{\"var\":\"", gen[i].key, "\",\"cnt\":", gen[i].cnt, "}")))
		if i < len(gen)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("],\n \"txs\":["))

	for i := range txs {
		w.Write([]byte(fmt.Sprint("{\"var\":\"", txs[i].key, "\",\"cnt\":", txs[i].cnt, "}")))
		if i < len(txs)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("],\n \"net\":["))

	for i := range net {
		fin := "_" + net[i]
		w.Write([]byte("{\"var\":\"" + net[i] + "\","))
		common.CounterMutex.Lock()
		w.Write([]byte(fmt.Sprint("\"rcvd\":", common.Counter["rcvd"+fin], ",")))
		w.Write([]byte(fmt.Sprint("\"rbts\":", common.Counter["rbts"+fin], ",")))
		w.Write([]byte(fmt.Sprint("\"sent\":", common.Counter["sent"+fin], ",")))
		w.Write([]byte(fmt.Sprint("\"sbts\":", common.Counter["sbts"+fin], "}")))
		common.CounterMutex.Unlock()
		if i < len(net)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]\n}\n"))
}
