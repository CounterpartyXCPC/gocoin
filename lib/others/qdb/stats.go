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

// File:        stats.go
// Description: Bictoin Cash Cash qdb Package

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

package qdb

import (
	"fmt"
	"sort"
	"sync"
)

var (
	counter       map[string]uint64 = make(map[string]uint64)
	counter_mutex sync.Mutex
)

func cnt(k string) {
	cntadd(k, 1)
}

func cntadd(k string, val uint64) {
	counter_mutex.Lock()
	counter[k] += val
	counter_mutex.Unlock()
}

func GetStats() (s string) {
	counter_mutex.Lock()
	ck := make([]string, len(counter))
	idx := 0
	for k := range counter {
		ck[idx] = k
		idx++
	}
	sort.Strings(ck)

	for i := range ck {
		k := ck[i]
		v := counter[k]
		s += fmt.Sprintln(k, ": ", v)
	}
	counter_mutex.Unlock()
	return s
}
