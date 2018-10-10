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

// File:        hidepass_unix.go
// Description: Bictoin Cash Cash sys Package

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

// +build !windows

package sys

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var wsta syscall.WaitStatus = 0

func enterpassext(b []byte) (n int) {
	si := make(chan os.Signal, 10)
	br := make(chan bool)
	fd := []uintptr{os.Stdout.Fd()}

	signal.Notify(si, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM)
	go sighndl(fd, si, br)

	pid, er := syscall.ForkExec("/bin/stty", []string{"stty", "-echo"}, &syscall.ProcAttr{Dir: "", Files: fd})
	if er == nil {
		syscall.Wait4(pid, &wsta, 0, nil)
		n = getline(b)
		close(br)
		echo(fd)
		fmt.Println()
	} else {
		n = getline(b)
	}

	return
}

func echo(fd []uintptr) {
	pid, e := syscall.ForkExec("/bin/stty", []string{"stty", "echo"}, &syscall.ProcAttr{Dir: "", Files: fd})
	if e == nil {
		syscall.Wait4(pid, &wsta, 0, nil)
	}
}

func sighndl(fd []uintptr, signal chan os.Signal, br chan bool) {
	select {
	case <-signal:
		echo(fd)
		os.Exit(-1)
	case <-br:
	}
}

func init() {
	secrespass = enterpassext
}
