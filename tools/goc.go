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

// File:		goc.go
// Description:	Bictoin Cash Cash main Package

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

package main

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

var (
	HOST string
	SID  string
)

func http_get(url string) (res []byte) {
	req, _ := http.NewRequest("GET", url, nil)
	if SID != "" {
		req.AddCookie(&http.Cookie{Name: "sid", Value: SID})
	}
	r, er := new(http.Client).Do(req)
	if er != nil {
		println(url, er.Error())
		os.Exit(1)
	}
	if SID == "" {
		for i := range r.Cookies() {
			if r.Cookies()[i].Name == "sid" {
				SID = r.Cookies()[i].Value
				//fmt.Println("sid", SID)
			}
		}
	}
	if r.StatusCode == 200 {
		defer r.Body.Close()
		res, _ = ioutil.ReadAll(r.Body)
	} else {
		println(url, "http.Get returned code", r.StatusCode)
		os.Exit(1)
	}
	return
}

func fetch_balance() {
	os.RemoveAll("balance/")

	d := http_get(HOST + "balance.zip")
	r, er := zip.NewReader(bytes.NewReader(d), int64(len(d)))
	if er != nil {
		println(er.Error())
		os.Exit(1)
	}

	os.Mkdir("balance/", 0777)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		dat, _ := ioutil.ReadAll(rc)
		rc.Close()
		ioutil.WriteFile(f.Name, dat, 0666)
	}
}

func list_wallets() {
	d := http_get(HOST + "wallets.xml")
	var wls struct {
		Wallet []struct {
			Name     string
			Selected bool
		}
	}
	er := xml.Unmarshal(d, &wls)
	if er != nil {
		println(er.Error())
		os.Exit(1)
	}
	for i := range wls.Wallet {
		fmt.Print(wls.Wallet[i].Name)
		if wls.Wallet[i].Selected {
			fmt.Print(" (selected)")
		}
		fmt.Println()
	}
}

func switch_to_wallet(s string) {
	http_get(HOST + "cfg") // get SID
	u, _ := url.Parse(HOST + "cfg")
	ps := url.Values{}
	ps.Add("sid", SID)
	ps.Add("qwalsel", s)
	u.RawQuery = ps.Encode()
	http_get(u.String())
}

func push_tx(rawtx string) {
	dat := sys.GetRawData(rawtx)
	if dat == nil {
		println("Cannot fetch the raw transaction data (specify hexdump or filename)")
		return
	}

	val := make(url.Values)
	val["rawtx"] = []string{hex.EncodeToString(dat)}

	r, er := http.PostForm(HOST+"txs", val)
	if er != nil {
		println(er.Error())
		os.Exit(1)
	}
	if r.StatusCode == 200 {
		defer r.Body.Close()
		res, _ := ioutil.ReadAll(r.Body)
		if len(res) > 100 {
			txid := bch.NewSha2Hash(dat)
			fmt.Println("TxID", txid.String(), "loaded")

			http_get(HOST + "cfg") // get SID
			//fmt.Println("sid", SID)

			u, _ := url.Parse(HOST + "txs2s.xml")
			ps := url.Values{}
			ps.Add("sid", SID)
			ps.Add("send", txid.String())
			u.RawQuery = ps.Encode()
			http_get(u.String())
		}
	} else {
		println("http.Post returned code", r.StatusCode)
		os.Exit(1)
	}
}

func show_help() {
	fmt.Println("Specify the command and (optionally) its arguments:")
	fmt.Println("  wal [wallet_name] - switch to a given wallet (or list them)")
	fmt.Println("  bal - creates balance/ folder for current wallet")
	fmt.Println("  ptx <rawtx> - pushes raw tx into the network")
}

func main() {
	if len(os.Args) < 2 {
		show_help()
		return
	}

	HOST = os.Getenv("GOCOIN_WEBUI")
	if HOST == "" {
		HOST = "http://127.0.0.1:8833/"
	} else {
		if !strings.HasPrefix(HOST, "http://") {
			HOST = "http://" + HOST
		}
		if !strings.HasSuffix(HOST, "/") {
			HOST = HOST + "/"
		}
	}
	fmt.Println("Gocoin WebUI at", HOST, "(you can overwrite it via env variable GOCOIN_WEBUI)")

	switch os.Args[1] {
	case "wal":
		if len(os.Args) > 2 {
			switch_to_wallet(os.Args[2])
		} else {
			list_wallets()
		}

	case "bal":
		fetch_balance()

	case "ptx":
		push_tx(os.Args[2])

	default:
		show_help()
	}
}
