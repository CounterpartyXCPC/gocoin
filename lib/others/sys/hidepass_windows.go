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

// File:        hidepass_windows.go
// Description: Bictoin Cash Cash sys Package

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

package sys

import (
	"fmt"
	"os"
)

// New method (requires msvcrt.dll):
import (
	"syscall"
)

var (
	msvcrt = syscall.NewLazyDLL("msvcrt.dll")
	_getch = msvcrt.NewProc("_getch")
)

func getch() int {
	res, _, _ := syscall.Syscall(_getch.Addr(), 0, 0, 0, 0)
	return int(res)
}

func enterpassext(b []byte) (n int) {
	for {
		chr := byte(getch())
		if chr == 3 {
			// Ctrl+C
			ClearBuffer(b)
			os.Exit(0)
		}
		if chr == 13 || chr == 10 {
			fmt.Println()
			break // Enter
		}
		if chr == '\b' {
			if n > 0 {
				n--
				b[n] = 0
				fmt.Print("\b \b")
			} else {
				fmt.Print("\007")
			}
			continue
		}
		if chr < ' ' {
			fmt.Print("\007")
			fmt.Println("\n", chr)
			continue
		}
		if n == len(b) {
			fmt.Print("\007")
			continue
		}
		fmt.Print("*")
		b[n] = chr
		n++
	}
	return
}

func init() {
	er := _getch.Find()
	if er != nil {
		println(er.Error())
		println("WARNING: Characters will be visible during password input.")
		return
	}

	secrespass = enterpassext
}

/*
Old method (requires mingw):

#include <conio.h>
// end the comment here when enablign this method
import "C"

func getch() int {
	return int(C._getch())
}

*/
