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

// File:        sipasec_windows.go
// Description: Bictoin Cash Cash sipasec Package

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

// +build windows

package sipasec

/*
1. MSYS2 + MinGW64
See the following web pages for info on installing MSYS2 and mingw64 for your Windows OS.
Please note that you will need the 64-bit compiler.
 * http://www.msys2.org/
 * https://stackoverflow.com/questions/30069830/how-to-install-mingw-w64-and-msys2#30071634


2. Dependencies
After having MSYS2 and Mingw64 installed, you have to install dependency packages.
Just execute the following command from within the "MSYS2 MSYS" shell:

 > pacman -S make autoconf automake libtoolm lzip


3. gmplib + secp256k1
Now use "MSYS2 MinGW 64-bit" shell and execute:

 > cd ~
 > wget https://gmplib.org/download/gmp/gmp-6.1.2.tar.lz
 > tar vxf gmp-6.1.2.tar.lz
 > cd gmp-6.1.2
 > ./configure
 > make
 > make install

 > cd ~
 > git clone https://github.com/bitcoin/bitcoin.git
 > cd bitcoin/src/secp256k1/
 > ./autogen.sh
 > ./configure
 > make
 > make install



If everything went well, you should see "PASS" executing "go test" in this folder.
Then copy "gocoin/client/speedups/sipasec.go" to "gocoin/client/" to boost your client.
*/

// #cgo LDFLAGS: -lsecp256k1 -lgmp
import "C"
