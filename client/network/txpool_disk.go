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

// File:		txpool_disk.go
// Description:	Bictoin Cash network Package

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

package network

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

var (
	END_MARKER = []byte("END_OF_FILE")
)

const (
	MEMPOOL_FILE_NAME2 = "mempool.dmp"
)

func bool2byte(v bool) byte {
	if v {
		return 1
	} else {
		return 0
	}
}

func (t2s *OneTxToSend) WriteBytes(wr io.Writer) {
	bch.WriteVlen(wr, uint64(len(t2s.Raw)))
	wr.Write(t2s.Raw)

	bch.WriteVlen(wr, uint64(len(t2s.Spent)))
	binary.Write(wr, binary.LittleEndian, t2s.Spent[:])

	binary.Write(wr, binary.LittleEndian, t2s.Invsentcnt)
	binary.Write(wr, binary.LittleEndian, t2s.SentCnt)
	binary.Write(wr, binary.LittleEndian, uint32(t2s.Firstseen.Unix()))
	binary.Write(wr, binary.LittleEndian, uint32(t2s.Lastsent.Unix()))
	binary.Write(wr, binary.LittleEndian, t2s.Volume)
	binary.Write(wr, binary.LittleEndian, t2s.Fee)
	binary.Write(wr, binary.LittleEndian, t2s.SigopsCost)
	binary.Write(wr, binary.LittleEndian, t2s.VerifyTime)
	wr.Write([]byte{bool2byte(t2s.Local), t2s.BchBlocked, bool2byte(t2s.MemInputs != nil), bool2byte(t2s.Final)})
}

func MempoolSave(force bool) {
	if !force && !common.CFG.TXPool.SaveOnDisk {
		os.Remove(common.GocoinCashHomeDir + MEMPOOL_FILE_NAME2)
		return
	}

	f, er := os.Create(common.GocoinCashHomeDir + MEMPOOL_FILE_NAME2)
	if er != nil {
		println(er.Error())
		return
	}

	fmt.Println("Saving", MEMPOOL_FILE_NAME2)
	wr := bufio.NewWriter(f)

	wr.Write(common.Last.BchBlock.BchBlockHash.Hash[:])

	bch.WriteVlen(wr, uint64(len(TransactionsToSend)))
	for _, t2s := range TransactionsToSend {
		t2s.WriteBytes(wr)
	}

	bch.WriteVlen(wr, uint64(len(SpentOutputs)))
	for k, v := range SpentOutputs {
		binary.Write(wr, binary.LittleEndian, k)
		binary.Write(wr, binary.LittleEndian, v)
	}

	wr.Write(END_MARKER[:])
	wr.Flush()
	f.Close()
}

func MempoolLoad2() bool {
	var t2s *OneTxToSend
	var totcnt, le uint64
	var tmp [32]byte
	var bi BIDX
	var tina uint32
	var i int
	var cnt1, cnt2 uint

	f, er := os.Open(common.GocoinCashHomeDir + MEMPOOL_FILE_NAME2)
	if er != nil {
		fmt.Println("MempoolLoad:", er.Error())
		return false
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	if er = bch.ReadAll(rd, tmp[:32]); er != nil {
		goto fatal_error
	}
	if !bytes.Equal(tmp[:32], common.Last.BchBlock.BchBlockHash.Hash[:]) {
		er = errors.New(MEMPOOL_FILE_NAME2 + " is for different last block hash (try to load it with 'mpl' command)")
		goto fatal_error
	}

	if totcnt, er = bch.ReadVLen(rd); er != nil {
		goto fatal_error
	}

	TransactionsToSend = make(map[BIDX]*OneTxToSend, int(totcnt))
	for ; totcnt > 0; totcnt-- {
		le, er = bch.ReadVLen(rd)
		if er != nil {
			goto fatal_error
		}

		t2s = new(OneTxToSend)
		raw := make([]byte, int(le))

		er = bch.ReadAll(rd, raw)
		if er != nil {
			goto fatal_error
		}

		t2s.Tx, i = bch.NewTx(raw)
		if t2s.Tx == nil || i != len(raw) {
			er = errors.New(fmt.Sprint("Error parsing tx from ", MEMPOOL_FILE_NAME2, " at idx", len(TransactionsToSend)))
			goto fatal_error
		}
		t2s.Tx.SetHash(raw)

		le, er = bch.ReadVLen(rd)
		if er != nil {
			goto fatal_error
		}
		t2s.Spent = make([]uint64, int(le))
		if er = binary.Read(rd, binary.LittleEndian, t2s.Spent[:]); er != nil {
			goto fatal_error
		}

		if er = binary.Read(rd, binary.LittleEndian, &t2s.Invsentcnt); er != nil {
			goto fatal_error
		}

		if er = binary.Read(rd, binary.LittleEndian, &t2s.SentCnt); er != nil {
			goto fatal_error
		}

		if er = binary.Read(rd, binary.LittleEndian, &tina); er != nil {
			goto fatal_error
		}
		t2s.Firstseen = time.Unix(int64(tina), 0)

		if er = binary.Read(rd, binary.LittleEndian, &tina); er != nil {
			goto fatal_error
		}
		t2s.Lastsent = time.Unix(int64(tina), 0)

		if er = binary.Read(rd, binary.LittleEndian, &t2s.Volume); er != nil {
			goto fatal_error
		}

		if er = binary.Read(rd, binary.LittleEndian, &t2s.Fee); er != nil {
			goto fatal_error
		}

		if er = binary.Read(rd, binary.LittleEndian, &t2s.SigopsCost); er != nil {
			goto fatal_error
		}

		if er = binary.Read(rd, binary.LittleEndian, &t2s.VerifyTime); er != nil {
			goto fatal_error
		}

		if er = bch.ReadAll(rd, tmp[:4]); er != nil {
			goto fatal_error
		}
		t2s.Local = tmp[0] != 0
		t2s.BchBlocked = tmp[1]
		if tmp[2] != 0 {
			t2s.MemInputs = make([]bool, len(t2s.TxIn))
		}
		t2s.Final = tmp[3] != 0

		t2s.Tx.Fee = t2s.Fee

		TransactionsToSend[t2s.Hash.BIdx()] = t2s
		TransactionsToSendSize += uint64(len(t2s.Raw))
		TransactionsToSendWeight += uint64(t2s.Weight())
	}

	if totcnt, er = bch.ReadVLen(rd); er != nil {
		goto fatal_error
	}

	SpentOutputs = make(map[uint64]BIDX, int(totcnt))
	for ; totcnt > 0; totcnt-- {
		if er = binary.Read(rd, binary.LittleEndian, &le); er != nil {
			goto fatal_error
		}

		if er = binary.Read(rd, binary.LittleEndian, &bi); er != nil {
			goto fatal_error
		}

		SpentOutputs[le] = bi
	}

	if er = bch.ReadAll(rd, tmp[:len(END_MARKER)]); er != nil {
		goto fatal_error
	}
	if !bytes.Equal(tmp[:len(END_MARKER)], END_MARKER) {
		er = errors.New(MEMPOOL_FILE_NAME2 + " has marker missing")
		goto fatal_error
	}

	// recover MemInputs
	for _, t2s := range TransactionsToSend {
		if t2s.MemInputs != nil {
			cnt1++
			for i := range t2s.TxIn {
				if _, inmem := TransactionsToSend[bch.BIdx(t2s.TxIn[i].Input.Hash[:])]; inmem {
					t2s.MemInputs[i] = true
					t2s.MemInputCnt++
					cnt2++
				}
			}
			if t2s.MemInputCnt == 0 {
				println("ERROR: MemInputs not nil but nothing found")
				t2s.MemInputs = nil
			}
		}
	}

	fmt.Println(len(TransactionsToSend), "transactions taking", TransactionsToSendSize, "Bytes loaded from", MEMPOOL_FILE_NAME2)
	fmt.Println(cnt1, "transactions use", cnt2, "memory inputs")

	return true

fatal_error:
	fmt.Println("Error loading", MEMPOOL_FILE_NAME2, ":", er.Error())
	TransactionsToSend = make(map[BIDX]*OneTxToSend)
	TransactionsToSendSize = 0
	TransactionsToSendWeight = 0
	SpentOutputs = make(map[uint64]BIDX)
	return false
}

// this one is only called from TextUI
func MempoolLoadNew(fname string, abort *bool) bool {
	var ntx *TxRcvd
	var idx, totcnt, le, tmp64, oneperc, cntdwn, perc uint64
	var tmp [32]byte
	var tina uint32
	var i int
	var cnt1, cnt2 uint
	var t2s OneTxToSend

	f, er := os.Open(fname)
	if er != nil {
		fmt.Println("MempoolLoad:", er.Error())
		return false
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	if er = bch.ReadAll(rd, tmp[:32]); er != nil {
		goto fatal_error
	}

	if totcnt, er = bch.ReadVLen(rd); er != nil {
		goto fatal_error
	}
	fmt.Println("Loading", totcnt, "transactions from", fname)

	oneperc = totcnt / 100

	for idx = 0; idx < totcnt; idx++ {
		if cntdwn == 0 {
			fmt.Print("\r", perc, "% complete...")
			perc++
			cntdwn = oneperc
		}
		cntdwn--
		if abort != nil && *abort {
			break
		}
		le, er = bch.ReadVLen(rd)
		if er != nil {
			goto fatal_error
		}

		ntx = new(TxRcvd)
		raw := make([]byte, int(le))

		er = bch.ReadAll(rd, raw)
		if er != nil {
			goto fatal_error
		}

		ntx.Tx, i = bch.NewTx(raw)
		if ntx.Tx == nil || i != len(raw) {
			er = errors.New(fmt.Sprint("Error parsing tx from ", fname, " at idx", idx))
			goto fatal_error
		}
		ntx.SetHash(raw)

		le, er = bch.ReadVLen(rd)
		if er != nil {
			goto fatal_error
		}

		for le > 0 {
			if er = binary.Read(rd, binary.LittleEndian, &tmp64); er != nil {
				goto fatal_error
			}
			le--
		}

		// discard all the rest...
		binary.Read(rd, binary.LittleEndian, &t2s.Invsentcnt)
		binary.Read(rd, binary.LittleEndian, &t2s.SentCnt)
		binary.Read(rd, binary.LittleEndian, &tina)
		binary.Read(rd, binary.LittleEndian, &tina)
		binary.Read(rd, binary.LittleEndian, &t2s.Volume)
		binary.Read(rd, binary.LittleEndian, &t2s.Fee)
		binary.Read(rd, binary.LittleEndian, &t2s.SigopsCost)
		binary.Read(rd, binary.LittleEndian, &t2s.VerifyTime)
		if er = bch.ReadAll(rd, tmp[:4]); er != nil {
			goto fatal_error
		}

		// submit tx if we dont have it yet...
		if NeedThisTx(&ntx.Hash, nil) {
			cnt2++
			if HandleNetTx(ntx, true) {
				cnt1++
			}
		}
	}

	fmt.Print("\r                                    \r")
	fmt.Println(cnt1, "out of", cnt2, "new transactions accepted into memory pool")

	return true

fatal_error:
	fmt.Println("Error loading", fname, ":", er.Error())
	return false
}
