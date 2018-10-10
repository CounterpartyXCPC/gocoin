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

// File:        z_init.go
// Description: Bictoin Cash Cash secp256k1 Package

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

package secp256k1

/*
import (
	"os"
	"fmt"
	"time"
)


var (
	pre_g, pre_g_128 []XY
	prec [64][16]XY
	fin XY
)


func ecmult_start() {
	return
	sta := time.Now()

	g := TheCurve.G

	// calculate 2^128*generator
	var g_128j XYZ
	g_128j.SetXY(&g)

	for i := 0; i < 128; i++ {
		g_128j.Double(&g_128j)
	}

	var g_128 XY
	g_128.SetXYZ(&g_128j)

    // precompute the tables with odd multiples
	pre_g = g.precomp(WINDOW_G)
	pre_g_128 = g_128.precomp(WINDOW_G)

	// compute prec and fin
	var gg XYZ
	gg.SetXY(&g)
	ad := g
	var fn XYZ
	fn.Infinity = true
	for j:=0; j<64; j++ {
		prec[j][0].SetXYZ(&gg)
		fn.Add(&fn, &gg)
		for i:=1; i<16; i++ {
			gg.AddXY(&gg, &ad)
			prec[j][i].SetXYZ(&gg)
		}
		ad = prec[j][15]
	}
	fin.SetXYZ(&fn)
	fin.Neg(&fin)

	if false {
		f, _ := os.Create("z_prec.go")
		fmt.Fprintln(f, "package secp256k1\n\nvar prec = [64][16]XY {")
		for j:=0; j<64; j++ {
			fmt.Fprintln(f, " {")
			for i:=0; i<16; i++ {
				fmt.Fprintln(f, "{X:" + fe2str(&prec[j][i].X) + ", Y:" + fe2str(&prec[j][i].Y) + "},")
			}
			fmt.Fprintln(f, "},")
		}
		fmt.Fprintln(f, "}")
		f.Close()
	}

	if false {
		f, _ := os.Create("z_pre_g.go")
		fmt.Fprintln(f, "package secp256k1\n\nvar pre_g = []XY {")
		for i := range pre_g {
			fmt.Fprintln(f, "{X:" + fe2str(&pre_g[i].X) + ", Y:" + fe2str(&pre_g[i].Y) + "},")
		}
		fmt.Fprintln(f, "}")
		f.Close()
	}

	if false {
		f, _ := os.Create("z_pre_g_128.go")
		fmt.Fprintln(f, "package secp256k1\n\nvar pre_g_128 = []XY {")
		for i := range pre_g_128 {
			fmt.Fprintln(f, "{X:" + fe2str(&pre_g_128[i].X) + ", Y:" + fe2str(&pre_g_128[i].Y) + "},")
		}
		fmt.Fprintln(f, "}")
		f.Close()
	}

	if false {
		f, _ := os.Create("z_fin.go")
		fmt.Fprintln(f, "package secp256k1\n\nvar fim = XY {")
		fmt.Fprintln(f, "X:" + fe2str(&fin.X) + ", Y:" + fe2str(&fin.Y) + ",")
		fmt.Fprintln(f, "}")
		f.Close()
	}

	println("start done in", time.Now().Sub(sta).String())
}


func fe2str(f *Field) (s string) {
	s = fmt.Sprintf("Field{[10]uint32{0x%08x", f.n[0])
	for i:=1; i<len(f.n); i++ {
		s += fmt.Sprintf(", 0x%08x", f.n[i])
	}
	s += "}}"
	return
}


*/
