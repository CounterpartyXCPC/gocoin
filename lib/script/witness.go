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

// File:        witness.go
// Description: Bictoin Cash Cash script Package

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

package script

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

type witness_ctx struct {
	stack scrStack
}

func (w *witness_ctx) IsNull() bool {
	return w.stack.size() == 0
}

func VerifyWitnessProgram(witness *witness_ctx, amount uint64, tx *bch.Tx, inp int, witversion int, program []byte, flags uint32) bool {
	var stack scrStack
	var scriptPubKey []byte

	if DBG_SCR {
		fmt.Println("*****************VerifyWitnessProgram", len(tx.SegWit), witversion, flags, witness.stack.size(), len(program))
	}

	if witversion == 0 {
		if len(program) == 32 {
			// Version 0 segregated witness program: SHA256(CScript) inside the program, CScript + inputs in witness
			if witness.stack.size() == 0 {
				if DBG_ERR {
					fmt.Println("SCRIPT_ERR_WITNESS_PROGRAM_WITNESS_EMPTY")
				}
				return false
			}
			scriptPubKey = witness.stack.pop()
			sha := sha256.New()
			sha.Write(scriptPubKey)
			sum := sha.Sum(nil)
			if !bytes.Equal(program, sum) {
				if DBG_ERR {
					fmt.Println("32-SCRIPT_ERR_WITNESS_PROGRAM_MISMATCH")
					fmt.Println(hex.EncodeToString(program))
					fmt.Println(hex.EncodeToString(sum))
					fmt.Println(hex.EncodeToString(scriptPubKey))
				}
				return false
			}
			stack.copy_from(&witness.stack)
			witness.stack.push(scriptPubKey)
		} else if len(program) == 20 {
			// Special case for pay-to-pubkeyhash; signature + pubkey in witness
			if witness.stack.size() != 2 {
				if DBG_ERR {
					fmt.Println("20-SCRIPT_ERR_WITNESS_PROGRAM_MISMATCH", tx.Hash.String())
				}
				return false
			}

			scriptPubKey = make([]byte, 25)
			scriptPubKey[0] = 0x76
			scriptPubKey[1] = 0xa9
			scriptPubKey[2] = 0x14
			copy(scriptPubKey[3:23], program)
			scriptPubKey[23] = 0x88
			scriptPubKey[24] = 0xac
			stack.copy_from(&witness.stack)
		} else {
			if DBG_ERR {
				fmt.Println("SCRIPT_ERR_WITNESS_PROGRAM_WRONG_LENGTH")
			}
			return false
		}
	} else if (flags & VER_WITNESS_PROG) != 0 {
		if DBG_ERR {
			fmt.Println("SCRIPT_ERR_DISCOURAGE_UPGRADABLE_WITNESS_PROGRAM")
		}
		return false
	} else {
		// Higher version witness scripts return true for future softfork compatibility
		return true
	}

	if DBG_SCR {
		fmt.Println("*****************", stack.size())
	}
	// Disallow stack item size > MAX_SCRIPT_ELEMENT_SIZE in witness stack
	for i := 0; i < stack.size(); i++ {
		if len(stack.at(i)) > bch.MAX_SCRIPT_ELEMENT_SIZE {
			if DBG_ERR {
				fmt.Println("SCRIPT_ERR_PUSH_SIZE")
			}
			return false
		}
	}

	if !evalScript(scriptPubKey, amount, &stack, tx, inp, flags, SIGVERSION_WITNESS_V0) {
		return false
	}

	// Scripts inside witness implicitly require cleanstack behaviour
	if stack.size() != 1 {
		if DBG_ERR {
			fmt.Println("SCRIPT_ERR_EVAL_FALSE")
		}
		return false
	}

	if !stack.topBool(-1) {
		if DBG_ERR {
			fmt.Println("SCRIPT_ERR_EVAL_FALSE")
		}
		return false
	}
	return true
}
