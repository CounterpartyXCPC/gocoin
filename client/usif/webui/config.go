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

// File:		config.go
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
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	"github.com/counterpartyxcpc/gocoin-cash/client/usif"
	"github.com/counterpartyxcpc/gocoin-cash/client/wallet"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/peersdb"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

func p_cfg(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	if common.CFG.WebUI.ServerMode {
		return
	}

	if r.Method == "POST" {
		if len(r.Form["configjson"]) > 0 {
			common.LockCfg()
			e := json.Unmarshal([]byte(r.Form["configjson"][0]), &common.CFG)
			if e == nil {
				common.Reset()
			}
			if len(r.Form["save"]) > 0 {
				common.SaveConfig()
			}
			common.UnlockCfg()
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if len(r.Form["friends_file"]) > 0 {
			ioutil.WriteFile(common.GocoinCashHomeDir+"friends.txt", []byte(r.Form["friends_file"][0]), 0600)
			network.Mutex_net.Lock()
			network.NextConnectFriends = time.Now()
			network.Mutex_net.Unlock()
			http.Redirect(w, r, "/net", http.StatusFound)
			return
		}

		if len(r.Form["shutdown"]) > 0 {
			usif.Exit_now.Set()
			w.Write([]byte("Your node should shut down soon"))
			return
		}

		if len(r.Form["wallet"]) > 0 {
			if r.Form["wallet"][0] == "on" {
				wallet.OnOff <- true
			} else if r.Form["wallet"][0] == "off" {
				wallet.OnOff <- false
			}
			if len(r.Form["page"]) > 0 {
				http.Redirect(w, r, r.Form["page"][0], http.StatusFound)
			} else {
				http.Redirect(w, r, "/wal", http.StatusFound)
			}
			return
		}

		return
	}

	// for any other GET we need a matching session-id
	if !checksid(r) {
		new_session_id(w)
		return
	}

	if len(r.Form["getmp"]) > 0 {
		if conid, e := strconv.ParseUint(r.Form["getmp"][0], 10, 32); e == nil {
			network.GetMP(uint32(conid))
		}
		return
	}

	if len(r.Form["freemem"]) > 0 {
		sys.FreeMem()
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if len(r.Form["drop"]) > 0 {
		if conid, e := strconv.ParseUint(r.Form["drop"][0], 10, 32); e == nil {
			network.DropPeer(uint32(conid))
		}
		http.Redirect(w, r, "net", http.StatusFound)
		return
	}

	if len(r.Form["conn"]) > 0 {
		ad, er := peersdb.NewAddrFromString(r.Form["conn"][0], false)
		if er != nil {
			w.Write([]byte(er.Error()))
			return
		}
		w.Write([]byte(fmt.Sprint("Connecting to ", ad.Ip())))
		ad.Manual = true
		network.DoNetwork(ad)
		return
	}

	// All the functions below change modify the config file
	common.LockCfg()
	defer common.UnlockCfg()

	if len(r.Form["txponoff"]) > 0 {
		common.CFG.TXPool.Enabled = !common.CFG.TXPool.Enabled
		http.Redirect(w, r, "txs", http.StatusFound)
		return
	}

	if len(r.Form["txronoff"]) > 0 {
		common.CFG.TXRoute.Enabled = !common.CFG.TXRoute.Enabled
		http.Redirect(w, r, "txs", http.StatusFound)
		return
	}

	if len(r.Form["lonoff"]) > 0 {
		common.CFG.Net.ListenTCP = !common.CFG.Net.ListenTCP
		common.ListenTCP = common.CFG.Net.ListenTCP
		http.Redirect(w, r, "net", http.StatusFound)
		return
	}

	if len(r.Form["savecfg"]) > 0 {
		dat, _ := json.MarshalIndent(&common.CFG, "", "    ")
		if dat != nil {
			ioutil.WriteFile(common.ConfigFile, dat, 0660)
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if len(r.Form["trusthash"]) > 0 {
		if bch.NewUint256FromString(r.Form["trusthash"][0]) != nil {
			common.CFG.LastTrustedBlock = r.Form["trusthash"][0]
			common.ApplyLastTrustedBlock()
		}
		w.Write([]byte(common.CFG.LastTrustedBlock))
		return
	}
}
