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

// File:		common.go
// Description:	Bictoin Cash common Package

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

package common

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/utils"
)

const (
	ConfigFile = "gocoin-cash.conf"
	Version    = uint32(70015)
	// Services   = uint64(0x00000009)
	Services = uint64(0x1) // Oct 11.
)

var (
	LogBuffer             = new(bytes.Buffer)
	Log       *log.Logger = log.New(LogBuffer, "", 0)

	BchBlockChain *bch_chain.Chain
	GenesisBlock  *bch.Uint256
	Magic         [4]byte
	Testnet       bool

	Last TheLastBlock

	GocoinCashHomeDir string
	StartTime         time.Time
	MaxPeersNeeded    int

	CounterMutex sync.Mutex
	Counter      map[string]uint64 = make(map[string]uint64)

	busyLine int32

	NetworkClosed sys.SyncBool

	AverageBlockSize sys.SyncInt

	allBalMinVal uint64

	DropSlowestEvery    time.Duration
	BchBlockExpireEvery time.Duration
	PingPeerEvery       time.Duration

	UserAgent string

	ListenTCP bool

	minFeePerKB, routeMinFeePerKB, minminFeePerKB uint64
	maxMempoolSizeBytes, maxRejectedSizeBytes     uint64

	KillChan chan os.Signal = make(chan os.Signal)

	SecretKey []byte // 32 bytes of secret key
	PublicKey string

	WalletON       bool
	WalletProgress uint32 // 0 for not / 1000 for max
	WalletOnIn     uint32

	BchBlockChainSynchronized bool

	lastTrustedBlock       *bch.Uint256
	LastTrustedBlockHeight uint32
)

type TheLastBlock struct {
	sync.Mutex // use it for writing and reading from non-chain thread
	BchBlock   *bch_chain.BchBlockTreeNode
	time.Time
}

func (b *TheLastBlock) BchBlockHeight() (res uint32) {
	b.Mutex.Lock()
	res = b.BchBlock.Height
	b.Mutex.Unlock()
	return
}

func CountSafe(k string) {
	CounterMutex.Lock()
	Counter[k]++
	CounterMutex.Unlock()
}

func CountSafeAdd(k string, val uint64) {
	CounterMutex.Lock()
	Counter[k] += val
	CounterMutex.Unlock()
}

func Busy() {
	var line int
	_, _, line, _ = runtime.Caller(1)
	atomic.StoreInt32(&busyLine, int32(line))
}

func BusyIn() int {
	return int(atomic.LoadInt32(&busyLine))
}

func BytesToString(val uint64) string {
	if val < 1e6 {
		return fmt.Sprintf("%.1f KB", float64(val)/1e3)
	} else if val < 1e9 {
		return fmt.Sprintf("%.2f MB", float64(val)/1e6)
	}
	return fmt.Sprintf("%.2f GB", float64(val)/1e9)
}

func NumberToString(num float64) string {
	if num > 1e24 {
		return fmt.Sprintf("%.2f Y", num/1e24)
	}
	if num > 1e21 {
		return fmt.Sprintf("%.2f Z", num/1e21)
	}
	if num > 1e18 {
		return fmt.Sprintf("%.2f E", num/1e18)
	}
	if num > 1e15 {
		return fmt.Sprintf("%.2f P", num/1e15)
	}
	if num > 1e12 {
		return fmt.Sprintf("%.2f T", num/1e12)
	}
	if num > 1e9 {
		return fmt.Sprintf("%.2f G", num/1e9)
	}
	if num > 1e6 {
		return fmt.Sprintf("%.2f M", num/1e6)
	}
	if num > 1e3 {
		return fmt.Sprintf("%.2f K", num/1e3)
	}
	return fmt.Sprintf("%.2f", num)
}

func HashrateToString(hr float64) string {
	return NumberToString(hr) + "H/s"
}

// Calculates average blocks size over the last "CFG.Stat.BSizeBlks" blocks
// Only call from blockchain thread.
func RecalcAverageBlockSize() {
	n := BchBlockChain.LastBlock()
	var sum, cnt uint
	for maxcnt := CFG.Stat.BSizeBlks; maxcnt > 0 && n != nil; maxcnt-- {
		sum += uint(n.BchBlockSize)
		cnt++
		n = n.Parent
	}
	if sum > 0 && cnt > 0 {
		AverageBlockSize.Store(int(sum / cnt))
	} else {
		AverageBlockSize.Store(204)
	}
}

func GetRawTx(BchBlockHeight uint32, txid *bch.Uint256) (data []byte, er error) {
	data, er = BchBlockChain.GetRawTx(BchBlockHeight, txid)
	if er != nil {
		if Testnet {
			data = utils.GetTestnetTxFromWeb(txid)
		} else {
			data = utils.GetTxFromWeb(txid)
		}
		if data != nil {
			er = nil
		} else {
			er = errors.New("GetRawTx and GetTxFromWeb failed for " + txid.String())
		}
	}
	return
}

func WalletPendingTick() (res bool) {
	mutex_cfg.Lock()
	if WalletOnIn > 0 && BchBlockChainSynchronized {
		WalletOnIn--
		res = WalletOnIn == 0
	}
	mutex_cfg.Unlock()
	return
}

// Make sure to call it with mutex_cfg locked
func ApplyLastTrustedBlock() {
	hash := bch.NewUint256FromString(CFG.LastTrustedBlock)
	lastTrustedBlock = hash
	LastTrustedBlockHeight = 0

	if BchBlockChain != nil {
		BchBlockChain.BchBlockIndexAccess.Lock()
		node := BchBlockChain.BchBlockIndex[hash.BIdx()]
		BchBlockChain.BchBlockIndexAccess.Unlock()
		if node != nil {
			LastTrustedBlockHeight = node.Height
			for node != nil {
				node.Trusted = true
				node = node.Parent
			}
		}
	}
}

func LastTrustedBlockMatch(h *bch.Uint256) (res bool) {
	mutex_cfg.Lock()
	res = lastTrustedBlock != nil && lastTrustedBlock.Equal(h)
	mutex_cfg.Unlock()
	return
}

func AcceptTx() (res bool) {
	mutex_cfg.Lock()
	res = CFG.TXPool.Enabled && BchBlockChainSynchronized
	mutex_cfg.Unlock()
	return
}
