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

// File:		peers.go
// Description:	Bictoin Cash Cash main Package

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
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/lib/others/qdb"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/utils"
)

type manyPeers []*utils.OnePeer

func (mp manyPeers) Len() int {
	return len(mp)
}

func (mp manyPeers) Less(i, j int) bool {
	return mp[i].Time > mp[j].Time
}

func (mp manyPeers) Swap(i, j int) {
	mp[i], mp[j] = mp[j], mp[i]
}

func main() {
	var dir string

	if len(os.Args) > 1 {
		dir = os.Args[1]
	} else {
		dir = sys.BitcoinHome() + "gocoin" + string(os.PathSeparator) + "btcnet" + string(os.PathSeparator) + "peers3"
	}

	db, er := qdb.NewDB(dir, true)

	if er != nil {
		println(er.Error())
		os.Exit(1)
	}

	println(db.Count(), "peers in databse", dir)
	if db.Count() == 0 {
		return
	}

	tmp := make(manyPeers, db.Count())
	cnt := 0
	db.Browse(func(k qdb.KeyType, v []byte) uint32 {
		np := utils.NewPeer(v)
		if !sys.ValidIp4(np.Ip4[:]) {
			return 0
		}
		if cnt < len(tmp) {
			tmp[cnt] = np
			cnt++
		}
		return 0
	})

	sort.Sort(tmp[:cnt])
	for cnt = 0; cnt < len(tmp) && cnt < 2500; cnt++ {
		ad := tmp[cnt]
		fmt.Printf("%3d) %16s   %5d  - seen %5d min ago\n", cnt+1,
			fmt.Sprintf("%d.%d.%d.%d", ad.Ip4[0], ad.Ip4[1], ad.Ip4[2], ad.Ip4[3]),
			ad.Port, (time.Now().Unix()-int64(ad.Time))/60)
	}
}
