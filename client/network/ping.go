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

// File:		ping.go
// Description:	Bictoin Cash network Package

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

package network

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
)

const (
	PingHistoryLength        = 20
	PingAssumedIfUnsupported = 4999 // ms
)

func (c *OneConnection) HandlePong(pl []byte) {
	if pl != nil {
		if !bytes.Equal(pl, c.PingInProgress) {
			common.CountSafe("PongMismatch")
			return
		}
		common.CountSafe("PongOK")
		c.ExpireBlocksToGet(nil, c.X.PingSentCnt)
	} else {
		common.CountSafe("PongTimeout")
	}
	ms := time.Now().Sub(c.LastPingSent) / time.Millisecond
	if ms == 0 {
		//println(c.ConnID, "Ping returned after 0ms")
		ms = 1
	}
	c.Mutex.Lock()
	c.X.PingHistory[c.X.PingHistoryIdx] = int(ms)
	c.X.PingHistoryIdx = (c.X.PingHistoryIdx + 1) % PingHistoryLength
	c.PingInProgress = nil
	c.Mutex.Unlock()
}

// Returns (median) average ping
// Make sure to called it within c.Mutex.Lock()
func (c *OneConnection) GetAveragePing() int {
	if !c.X.VersionReceived {
		return 0
	}
	if c.Node.Version > 60000 {
		var pgs [PingHistoryLength]int
		var act_len int
		for _, p := range c.X.PingHistory {
			if p != 0 {
				pgs[act_len] = p
				act_len++
			}
		}
		if act_len == 0 {
			return 0
		}
		sort.Ints(pgs[:act_len])
		return pgs[act_len/2]
	} else {
		return PingAssumedIfUnsupported
	}
}

type SortedConnections []struct {
	Conn          *OneConnection
	Ping          int
	BchBlockCount int
	TxsCount      int
	MinutesOnline int
	Special       bool
}

// Returns the slowest peers first
// Make suure to call it with locked Mutex_net
func GetSortedConnections() (list SortedConnections, any_ping bool, segwit_cnt int) {
	var cnt int
	var now time.Time
	var tlist SortedConnections
	now = time.Now()
	tlist = make(SortedConnections, len(OpenCons))
	for _, v := range OpenCons {
		v.Mutex.Lock()
		tlist[cnt].Conn = v
		tlist[cnt].Ping = v.GetAveragePing()
		tlist[cnt].BchBlockCount = len(v.blocksreceived)
		tlist[cnt].TxsCount = v.X.TxsReceived
		tlist[cnt].Special = v.X.IsSpecial
		if v.X.VersionReceived == false || v.X.ConnectedAt.IsZero() {
			tlist[cnt].MinutesOnline = 0
		} else {
			tlist[cnt].MinutesOnline = int(now.Sub(v.X.ConnectedAt) / time.Minute)
		}
		v.Mutex.Unlock()

		if tlist[cnt].Ping > 0 {
			any_ping = true
		}
		if (v.Node.Services & SERVICE_SEGWIT) != 0 {
			segwit_cnt++
		}

		cnt++
	}
	if cnt > 0 {
		list = make(SortedConnections, len(tlist))
		var ignore_bcnt bool // otherwise count blocks
		var idx, best_idx, bcnt, best_bcnt, best_tcnt, best_ping int

		for idx = len(list) - 1; idx >= 0; idx-- {
			best_idx = -1
			for i, v := range tlist {
				if v.Conn == nil {
					continue
				}
				if best_idx < 0 {
					best_idx = i
					best_tcnt = v.TxsCount
					best_bcnt = v.BchBlockCount
					best_ping = v.Ping
				} else {
					if ignore_bcnt {
						bcnt = best_bcnt
					} else {
						bcnt = v.BchBlockCount
					}
					if best_bcnt < bcnt ||
						best_bcnt == bcnt && best_tcnt < v.TxsCount ||
						best_bcnt == bcnt && best_tcnt == v.TxsCount && best_ping > v.Ping {
						best_bcnt = v.BchBlockCount
						best_tcnt = v.TxsCount
						best_ping = v.Ping
						best_idx = i
					}
				}
			}
			list[idx] = tlist[best_idx]
			tlist[best_idx].Conn = nil
			ignore_bcnt = !ignore_bcnt
		}
	}
	return
}

// This function should be called only when OutConsActive >= MaxOutCons
func drop_worst_peer() bool {
	var list SortedConnections
	var any_ping bool
	var segwit_cnt int

	Mutex_net.Lock()
	defer Mutex_net.Unlock()

	list, any_ping, segwit_cnt = GetSortedConnections()
	if !any_ping { // if "list" is empty "any_ping" will also be false
		return false
	}

	for _, v := range list {
		if v.MinutesOnline < OnlineImmunityMinutes {
			continue
		}
		if v.Special {
			continue
		}
		if common.CFG.Net.MinSegwitCons > 0 && segwit_cnt <= int(common.CFG.Net.MinSegwitCons) &&
			(v.Conn.Node.Services&SERVICE_SEGWIT) != 0 {
			continue
		}
		if v.Conn.X.Incomming {
			if InConsActive+2 > common.GetUint32(&common.CFG.Net.MaxInCons) {
				common.CountSafe("PeerInDropped")
				if common.FLAG.Log {
					f, _ := os.OpenFile("drop_log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
					if f != nil {
						fmt.Fprintf(f, "%s: Drop incomming id:%d  blks:%d  txs:%d  ping:%d  mins:%d\n",
							time.Now().Format("2006-01-02 15:04:05"),
							v.Conn.ConnID, v.BchBlockCount, v.TxsCount, v.Ping, v.MinutesOnline)
						f.Close()
					}
				}
				v.Conn.Disconnect("PeerInDropped")
				return true
			}
		} else {
			if OutConsActive+2 > common.GetUint32(&common.CFG.Net.MaxOutCons) {
				common.CountSafe("PeerOutDropped")
				if common.FLAG.Log {
					f, _ := os.OpenFile("drop_log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
					if f != nil {
						fmt.Fprintf(f, "%s: Drop outgoing id:%d  blks:%d  txs:%d  ping:%d  mins:%d\n",
							time.Now().Format("2006-01-02 15:04:05"),
							v.Conn.ConnID, v.BchBlockCount, v.TxsCount, v.Ping, v.MinutesOnline)
						f.Close()
					}
				}
				v.Conn.Disconnect("PeerOutDropped")
				return true
			}
		}
	}
	return false
}

func (c *OneConnection) TryPing() bool {
	if common.GetDuration(&common.PingPeerEvery) == 0 {
		return false // pinging disabled in global config
	}

	if c.Node.Version <= 60000 {
		return false // insufficient protocol version
	}

	if time.Now().Before(c.LastPingSent.Add(common.GetDuration(&common.PingPeerEvery))) {
		return false // not yet...
	}

	if c.PingInProgress != nil {
		c.HandlePong(nil) // this will set PingInProgress to nil
	}

	c.X.PingSentCnt++
	c.PingInProgress = make([]byte, 8)
	rand.Read(c.PingInProgress[:])
	c.SendRawMsg("ping", c.PingInProgress)
	c.LastPingSent = time.Now()
	//println(c.PeerAddr.Ip(), "ping...")
	return true
}
