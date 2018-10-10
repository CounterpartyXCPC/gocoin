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
// Description:	Bictoin Cash webui Package

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

package webui

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

func p_net(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	net_page := load_template("net.html")

	network.Mutex_net.Lock()
	net_page = strings.Replace(net_page, "{LISTEN_TCP}", fmt.Sprint(common.IsListenTCP(), network.TCPServerStarted), 1)
	net_page = strings.Replace(net_page, "{EXTERNAL_ADDR}", bch.NewNetAddr(network.BestExternalAddr()).String(), 1)

	network.Mutex_net.Unlock()

	d, _ := ioutil.ReadFile(common.GocoinCashHomeDir + "friends.txt")
	net_page = strings.Replace(net_page, "{FRIENDS_TXT}", html.EscapeString(string(d)), 1)

	write_html_head(w, r)
	w.Write([]byte(net_page))
	write_html_tail(w)
}

func json_netcon(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			fmt.Println("json_netcon recovered:", err.Error())
			fmt.Println(string(debug.Stack()))
		}
	}()

	network.Mutex_net.Lock()
	defer network.Mutex_net.Unlock()

	net_cons := make([]network.ConnInfo, len(network.OpenCons))
	tmp, _, _ := network.GetSortedConnections()
	i := len(net_cons)
	for _, v := range tmp {
		i--
		v.Conn.GetStats(&net_cons[i])
		net_cons[i].HasImmunity = v.MinutesOnline < network.OnlineImmunityMinutes
	}

	bx, er := json.Marshal(net_cons)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}

}

func json_peerst(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	if len(r.Form["id"]) == 0 {
		return
	}

	conid, e := strconv.ParseUint(r.Form["id"][0], 10, 32)
	if e != nil {
		return
	}

	var res *network.ConnInfo

	network.Mutex_net.Lock()
	for _, v := range network.OpenCons {
		if uint32(conid) == v.ConnID {
			res = new(network.ConnInfo)
			v.GetStats(res)
			break
		}
	}
	network.Mutex_net.Unlock()

	if res != nil {
		bx, er := json.Marshal(&res)
		if er == nil {
			w.Header()["Content-Type"] = []string{"application/json"}
			w.Write(bx)
		} else {
			println(er.Error())
		}
	}
}

func json_bwidth(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	type one_ext_ip struct {
		Ip               string
		Count, Timestamp uint
	}

	var out struct {
		Open_conns_total int
		Open_conns_out   uint32
		Open_conns_in    uint32
		Dl_speed_now     uint64
		Dl_speed_max     uint64
		Dl_total         uint64
		Ul_speed_now     uint64
		Ul_speed_max     uint64
		Ul_total         uint64
		ExternalIP       []one_ext_ip
		GetMPInProgress  bool
	}

	common.LockBw()
	common.TickRecv()
	common.TickSent()
	out.Dl_speed_now = common.GetAvgBW(common.DlBytesPrevSec[:], common.DlBytesPrevSecIdx, 5)
	out.Dl_speed_max = common.DownloadLimit()
	out.Dl_total = common.DlBytesTotal
	out.Ul_speed_now = common.GetAvgBW(common.UlBytesPrevSec[:], common.UlBytesPrevSecIdx, 5)
	out.Ul_speed_max = common.UploadLimit()
	out.Ul_total = common.UlBytesTotal
	common.UnlockBw()

	network.Mutex_net.Lock()
	out.Open_conns_total = len(network.OpenCons)
	out.Open_conns_out = network.OutConsActive
	out.Open_conns_in = network.InConsActive
	network.Mutex_net.Unlock()

	arr := network.GetExternalIPs()
	for _, rec := range arr {
		out.ExternalIP = append(out.ExternalIP, one_ext_ip{
			Ip:    fmt.Sprintf("%d.%d.%d.%d", byte(rec.IP>>24), byte(rec.IP>>16), byte(rec.IP>>8), byte(rec.IP)),
			Count: rec.Cnt, Timestamp: rec.Tim})
	}

	out.GetMPInProgress = len(network.GetMPInProgressTicket) != 0

	bx, er := json.Marshal(out)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}
}

func json_bwchar(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	var cnt uint64

	if len(r.Form["seconds"]) > 0 {
		cnt, _ = strconv.ParseUint(r.Form["seconds"][0], 10, 32)
	}
	if cnt < 1 {
		cnt = 1
	} else if cnt > 300 {
		cnt = 300
	}

	var out struct {
		DL           [200]uint64 // max 200 records (from 200 seconds to ~16.7 hours)
		UL           [200]uint64
		MaxDL, MaxUL uint64
	}

	common.LockBw()
	common.TickRecv()
	common.TickSent()

	idx := uint16(common.DlBytesPrevSecIdx)
	for i := range out.DL {
		var sum uint64
		for c := 0; c < int(cnt); c++ {
			idx--
			sum += common.DlBytesPrevSec[idx]
			if common.DlBytesPrevSec[idx] > out.MaxDL {
				out.MaxDL = common.DlBytesPrevSec[idx]
			}
		}
		out.DL[i] = sum / cnt
	}

	idx = uint16(common.UlBytesPrevSecIdx)
	for i := range out.UL {
		var sum uint64
		for c := 0; c < int(cnt); c++ {
			idx--
			sum += common.UlBytesPrevSec[idx]
			if common.UlBytesPrevSec[idx] > out.MaxUL {
				out.MaxUL = common.UlBytesPrevSec[idx]
			}
		}
		out.UL[i] = sum / cnt
	}

	common.UnlockBw()

	bx, er := json.Marshal(out)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}
}
