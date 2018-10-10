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

// File:		network.go
// Description:	Bictoin Cash textui Package

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

package textui

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/peersdb"
)

type SortedKeys []struct {
	Key    uint64
	ConnID uint32
}

func (sk SortedKeys) Len() int {
	return len(sk)
}

func (sk SortedKeys) Less(a, b int) bool {
	return sk[a].ConnID < sk[b].ConnID
}

func (sk SortedKeys) Swap(a, b int) {
	sk[a], sk[b] = sk[b], sk[a]
}

func net_drop(par string) {
	conid, e := strconv.ParseUint(par, 10, 32)
	if e != nil {
		println(e.Error())
		return
	}
	network.DropPeer(uint32(conid))
}

func node_info(par string) {
	conid, e := strconv.ParseUint(par, 10, 32)
	if e != nil {
		return
	}

	var r *network.ConnInfo

	network.Mutex_net.Lock()

	for _, v := range network.OpenCons {
		if uint32(conid) == v.ConnID {
			r = new(network.ConnInfo)
			v.GetStats(r)
			break
		}
	}
	network.Mutex_net.Unlock()

	if r == nil {
		return
	}

	fmt.Printf("Connection ID %d:\n", r.ID)
	if r.Incomming {
		fmt.Println("Comming from", r.PeerIp)
	} else {
		fmt.Println("Going to", r.PeerIp)
	}
	if !r.ConnectedAt.IsZero() {
		fmt.Println("Connected at", r.ConnectedAt.Format("2006-01-02 15:04:05"))
		if r.Version != 0 {
			fmt.Println("Node Version:", r.Version, "/ Services:", fmt.Sprintf("0x%x", r.Services))
			fmt.Println("User Agent:", r.Agent)
			fmt.Println("Chain Height:", r.Height)
			fmt.Printf("Reported IP: %d.%d.%d.%d\n", byte(r.ReportedIp4>>24), byte(r.ReportedIp4>>16),
				byte(r.ReportedIp4>>8), byte(r.ReportedIp4))
			fmt.Println("SendHeaders:", r.SendHeaders)
		}
		fmt.Println("Invs Done:", r.InvsDone)
		fmt.Println("Last data got:", time.Now().Sub(r.LastDataGot).String())
		fmt.Println("Last data sent:", time.Now().Sub(r.LastSent).String())
		fmt.Println("Last command received:", r.LastCmdRcvd, " ", r.LastBtsRcvd, "bytes")
		fmt.Println("Last command sent:", r.LastCmdSent, " ", r.LastBtsSent, "bytes")
		fmt.Print("Invs  Recieved:", r.InvsRecieved, "  Pending:", r.InvsToSend, "\n")
		fmt.Print("Bytes to send:", r.BytesToSend, " (", r.MaxSentBufSize, " max)\n")
		fmt.Print("BlockInProgress:", r.BchBlocksInProgress, "  GetHeadersInProgress:", r.GetHeadersInProgress, "\n")
		fmt.Println("GetBlocksDataNow:", r.GetBlocksDataNow)
		fmt.Println("AllHeadersReceived:", r.AllHeadersReceived)
		fmt.Println("Total Received:", r.BytesReceived, " /  Sent:", r.BytesSent)
		for k, v := range r.Counters {
			fmt.Println(k, ":", v)
		}
	} else {
		fmt.Println("Not yet connected")
	}
}

func net_conn(par string) {
	ad, er := peersdb.NewAddrFromString(par, false)
	if er != nil {
		fmt.Println(par, er.Error())
		return
	}
	fmt.Println("Connecting to", ad.Ip())
	ad.Manual = true
	network.DoNetwork(ad)
}

func net_stats(par string) {
	if par == "bw" {
		common.PrintBWStats()
		return
	} else if par != "" {
		node_info(par)
		return
	}

	network.Mutex_net.Lock()
	fmt.Printf("%d active net connections, %d outgoing\n", len(network.OpenCons), network.OutConsActive)
	srt := make(SortedKeys, len(network.OpenCons))
	cnt := 0
	for k, v := range network.OpenCons {
		srt[cnt].Key = k
		srt[cnt].ConnID = v.ConnID
		cnt++
	}
	sort.Sort(srt)
	for idx := range srt {
		v := network.OpenCons[srt[idx].Key]
		v.Mutex.Lock()
		fmt.Printf("%8d) ", v.ConnID)

		if v.X.Incomming {
			fmt.Print("<- ")
		} else {
			fmt.Print(" ->")
		}
		fmt.Printf(" %21s %5dms %7d : %-16s %7d : %-16s", v.PeerAddr.Ip(),
			v.GetAveragePing(), v.X.LastBtsRcvd, v.X.LastCmdRcvd, v.X.LastBtsSent, v.X.LastCmdSent)
		fmt.Printf("%9s %9s", common.BytesToString(v.X.Counters["BytesReceived"]), common.BytesToString(v.X.Counters["BytesSent"]))
		fmt.Print("  ", v.Node.Agent)

		if b2s := v.BytesToSent(); b2s > 0 {
			fmt.Print("  ", b2s)
		}
		v.Mutex.Unlock()
		fmt.Println()
	}

	if network.ExternalAddrLen() > 0 {
		fmt.Print("External addresses:")
		network.ExternalIpMutex.Lock()
		for ip, cnt := range network.ExternalIp4 {
			fmt.Printf(" %d.%d.%d.%d(%d)", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip), cnt)
		}
		network.ExternalIpMutex.Unlock()
		fmt.Println()
	} else {
		fmt.Println("No known external address")
	}

	network.Mutex_net.Unlock()

	fmt.Print("RecentlyDisconencted:")
	network.HammeringMutex.Lock()
	for ip, ti := range network.RecentlyDisconencted {
		fmt.Printf(" %d.%d.%d.%d-%s", ip[0], ip[1], ip[2], ip[3], time.Now().Sub(ti).String())
	}
	network.HammeringMutex.Unlock()
	fmt.Println()

	fmt.Println("GetMPInProgress:", len(network.GetMPInProgressTicket) != 0)

	common.PrintBWStats()
}

func init() {
	newUi("net n", false, net_stats, "Show network statistics. Specify ID to see its details.")
	newUi("drop", false, net_drop, "Disconenct from node with a given IP")
	newUi("conn", false, net_conn, "Connect to the given node (specify IP and optionally a port)")
}
