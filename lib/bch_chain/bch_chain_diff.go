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

// File:		bch_chain_diff.go
// Description:	Bictoin Cash Chain Package

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

package chain

import (
	"math/big"

	btc "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

const (
	POWRetargetSpam = 14 * 24 * 60 * 60 // two weeks
	TargetSpacing   = 10 * 60
	targetInterval  = POWRetargetSpam / TargetSpacing
)

func (ch *Chain) GetNextWorkRequired(lst *BlockTreeNode, ts uint32) (res uint32) {
	// Genesis block
	if lst.Parent == nil {
		return ch.Consensus.MaxPOWBits
	}

	if ((lst.Height + 1) % targetInterval) != 0 {
		// Special difficulty rule for testnet:
		if ch.testnet() {
			// If the new block's timestamp is more than 2* 10 minutes
			// then allow mining of a min-difficulty block.
			if ts > lst.Timestamp()+TargetSpacing*2 {
				return ch.Consensus.MaxPOWBits
			} else {
				// Return the last non-special-min-difficulty-rules-block
				prv := lst
				for prv.Parent != nil && (prv.Height%targetInterval) != 0 && prv.Bits() == ch.Consensus.MaxPOWBits {
					prv = prv.Parent
				}
				return prv.Bits()
			}
		}
		return lst.Bits()
	}

	prv := lst
	for i := 0; i < targetInterval-1; i++ {
		prv = prv.Parent
	}

	actualTimespan := int64(lst.Timestamp() - prv.Timestamp())

	if actualTimespan < POWRetargetSpam/4 {
		actualTimespan = POWRetargetSpam / 4
	}
	if actualTimespan > POWRetargetSpam*4 {
		actualTimespan = POWRetargetSpam * 4
	}

	// Retarget
	bnewbn := btc.SetCompact(lst.Bits())
	bnewbn.Mul(bnewbn, big.NewInt(actualTimespan))
	bnewbn.Div(bnewbn, big.NewInt(POWRetargetSpam))

	if bnewbn.Cmp(ch.Consensus.MaxPOWValue) > 0 {
		bnewbn = ch.Consensus.MaxPOWValue
	}

	res = btc.GetCompact(bnewbn)

	return
}

// Returns true if b1 has more POW than b2
func (b1 *BlockTreeNode) MorePOW(b2 *BlockTreeNode) bool {
	var b1sum, b2sum float64
	for b1.Height > b2.Height {
		b1sum += btc.GetDifficulty(b1.Bits())
		b1 = b1.Parent
	}
	for b2.Height > b1.Height {
		b2sum += btc.GetDifficulty(b2.Bits())
		b2 = b2.Parent
	}
	for b1 != b2 {
		b1sum += btc.GetDifficulty(b1.Bits())
		b2sum += btc.GetDifficulty(b2.Bits())
		b1 = b1.Parent
		b2 = b2.Parent
	}
	return b1sum > b2sum
}
