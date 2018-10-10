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

// File:		multisig_test.go
// Description:	Bictoin Cash Multisig Package Testing

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

package bch

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestMultisigFromScript(t *testing.T) {
	txt := "004730440220485ef45dd67e7e3ffee699d42cf56ec88b4162d9f373770c30efec075468281702204929343ea97b007c1fc2ed49b306355ebf6bc5fb1613f0ed51ebca44fcc2003a014c69512103af88375d5fc9230446365b7d33540a73397ab3cc1a9f3e306a26833d1bfc260f21030677e0dd58025a5404747fbc64083040083acf3b390515f71a8ede95dc9c4d8a2103af88375d5fc9230446365b7d33540a73397ab3cc1a9f3e306a26833d1bfc260f53ae"
	d, _ := hex.DecodeString(txt)
	s, e := NewMultiSigFromScript(d)
	if e != nil {
		t.Error(e.Error())
	}

	b := s.Bytes()
	if !bytes.Equal(b, d) {
		t.Error("Multisig script does not match the input\n", hex.EncodeToString(b), "\n", hex.EncodeToString(d))
	}
}
