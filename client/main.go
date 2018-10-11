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

// File:		main.go
// Description:	Bictoin Cash main Package

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
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"
	"unsafe"

	"github.com/counterpartyxcpc/gocoin-cash"
	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	"github.com/counterpartyxcpc/gocoin-cash/client/rpcapi"
	"github.com/counterpartyxcpc/gocoin-cash/client/usif"
	"github.com/counterpartyxcpc/gocoin-cash/client/usif/textui"
	"github.com/counterpartyxcpc/gocoin-cash/client/usif/webui"
	"github.com/counterpartyxcpc/gocoin-cash/client/wallet"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/peersdb"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/qdb"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

var (
	retryCachedBlocks bool
	SaveBlockChain    *time.Timer = time.NewTimer(24 * time.Hour)
)

const (
	SaveBlockChainAfter       = 2 * time.Second
	SaveBlockChainAfterNoSync = 10 * time.Minute
)

func reset_save_timer() {
	SaveBlockChain.Stop()
	for len(SaveBlockChain.C) > 0 {
		<-SaveBlockChain.C
	}
	if common.BchBlockChainSynchronized {
		SaveBlockChain.Reset(SaveBlockChainAfter)
	} else {
		SaveBlockChain.Reset(SaveBlockChainAfterNoSync)
	}
}

func blockMined(bl *bch.BchBlock) {
	network.BchBlockMined(bl)
	if int(bl.LastKnownHeight)-int(bl.Height) < 144 { // do not run it when syncing chain
		usif.ProcessBlockFees(bl.Height, bl)
	}
}

func LocalAcceptBlock(newbl *network.BchBlockRcvd) (e error) {
	print("LocalAcceptBlock")
	bl := newbl.BchBlock
	if common.FLAG.TrustAll || newbl.BchBlockTreeNode.Trusted {
		bl.Trusted = true
	}

	common.BchBlockChain.Unspent.AbortWriting() // abort saving of UTXO.db
	common.BchBlockChain.BchBlocks.BchBlockAdd(newbl.BchBlockTreeNode.Height, bl)
	newbl.TmQueue = time.Now()

	if newbl.DoInvs {
		common.Busy()
		network.NetRouteInv(network.MSG_BLOCK, bl.Hash, newbl.Conn)
	}

	network.MutexRcv.Lock()
	bl.LastKnownHeight = network.LastCommitedHeader.Height
	network.MutexRcv.Unlock()
	e = common.BchBlockChain.CommitBlock(bl, newbl.BchBlockTreeNode)

	if e == nil {
		// new block accepted
		newbl.TmAccepted = time.Now()

		newbl.NonWitnessSize = bl.NoWitnessSize

		common.RecalcAverageBlockSize()

		common.Last.Mutex.Lock()
		common.Last.Time = time.Now()
		common.Last.BchBlock = common.BchBlockChain.LastBlock()
		common.Last.Mutex.Unlock()

		reset_save_timer()
	} else {

		// Debugging Output (Optional)
		if common.CFG.TextUI_DevDebug {
			fmt.Println("Warning: AcceptBlock failed. If the block was valid, you may need to rebuild the unspent DB (-r)")
		}

		new_end := common.BchBlockChain.LastBlock()
		common.Last.Mutex.Lock()
		common.Last.BchBlock = new_end
		common.Last.Mutex.Unlock()
		// update network.LastCommitedHeader
		network.MutexRcv.Lock()
		if network.LastCommitedHeader != new_end {
			network.LastCommitedHeader = new_end

			// Debugging Output (Optional)
			if common.CFG.TextUI_DevDebug {
				println("LastCommitedHeader moved to", network.LastCommitedHeader.Height)
			}

		}
		network.DiscardedBlocks[newbl.Hash.BIdx()] = true
		network.MutexRcv.Unlock()
	}
	return
}

func retry_cached_blocks() bool {
	var idx int
	common.CountSafe("RedoCachedBlks")
	for idx < len(network.CachedBlocks) {
		newbl := network.CachedBlocks[idx]
		if CheckParentDiscarded(newbl.BchBlockTreeNode) {
			common.CountSafe("DiscardCachedBlock")
			if newbl.BchBlock == nil {
				os.Remove(common.TempBlocksDir() + newbl.BchBlockTreeNode.BchBlockHash.String())
			}
			network.CachedBlocks = append(network.CachedBlocks[:idx], network.CachedBlocks[idx+1:]...)
			network.CachedBlocksLen.Store(len(network.CachedBlocks))
			return len(network.CachedBlocks) > 0
		}
		if common.BchBlockChain.HasAllParents(newbl.BchBlockTreeNode) {
			common.Busy()

			if newbl.BchBlock == nil {
				tmpfn := common.TempBlocksDir() + newbl.BchBlockTreeNode.BchBlockHash.String()
				dat, e := ioutil.ReadFile(tmpfn)
				os.Remove(tmpfn)
				if e != nil {
					panic(e.Error())
				}
				if newbl.BchBlock, e = bch.NewBchBlock(dat); e != nil {
					panic(e.Error())
				}
				if e = newbl.BchBlock.BuildTxList(); e != nil {
					panic(e.Error())
				}
				newbl.BchBlock.BchBlockExtraInfo = *newbl.BchBlockExtraInfo
			}

			e := LocalAcceptBlock(newbl)
			if e != nil {
				fmt.Println("AcceptBlock2", newbl.BchBlockTreeNode.BchBlockHash.String(), "-", e.Error())
				newbl.Conn.Misbehave("LocalAcceptBl2", 250)
			}
			if usif.Exit_now.Get() {
				return false
			}
			// remove it from cache
			network.CachedBlocks = append(network.CachedBlocks[:idx], network.CachedBlocks[idx+1:]...)
			network.CachedBlocksLen.Store(len(network.CachedBlocks))
			return len(network.CachedBlocks) > 0
		} else {
			idx++
		}
	}
	return false
}

// Return true iof the block's parent is on the DiscardedBlocks list
// Add it to DiscardedBlocks, if returning true
func CheckParentDiscarded(n *bch_chain.BchBlockTreeNode) bool {
	network.MutexRcv.Lock()
	defer network.MutexRcv.Unlock()
	if network.DiscardedBlocks[n.Parent.BchBlockHash.BIdx()] {
		network.DiscardedBlocks[n.BchBlockHash.BIdx()] = true
		return true
	}
	return false
}

// Called from the blockchain thread
func HandleNetBlock(newbl *network.BchBlockRcvd) {
	defer func() {
		common.CountSafe("MainNetBlock")
		if common.GetUint32(&common.WalletOnIn) > 0 {
			common.SetUint32(&common.WalletOnIn, 5) // snooze the timer to 5 seconds from now
		}
	}()

	if CheckParentDiscarded(newbl.BchBlockTreeNode) {
		common.CountSafe("DiscardFreshBlockA")
		if newbl.BchBlock == nil {
			os.Remove(common.TempBlocksDir() + newbl.BchBlockTreeNode.BchBlockHash.String())
		}
		retryCachedBlocks = len(network.CachedBlocks) > 0
		return
	}

	if !common.BchBlockChain.HasAllParents(newbl.BchBlockTreeNode) {
		// it's not linking - keep it for later
		network.CachedBlocks = append(network.CachedBlocks, newbl)
		network.CachedBlocksLen.Store(len(network.CachedBlocks))
		common.CountSafe("BlockPostone")
		return
	}

	if newbl.BchBlock == nil {
		tmpfn := common.TempBlocksDir() + newbl.BchBlockTreeNode.BchBlockHash.String()
		dat, e := ioutil.ReadFile(tmpfn)
		os.Remove(tmpfn)
		if e != nil {
			panic(e.Error())
		}
		if newbl.BchBlock, e = bch.NewBchBlock(dat); e != nil {
			panic(e.Error())
		}
		if e = newbl.BchBlock.BuildTxList(); e != nil {
			panic(e.Error())
		}
		newbl.BchBlock.BchBlockExtraInfo = *newbl.BchBlockExtraInfo
	}

	common.Busy()
	if e := LocalAcceptBlock(newbl); e != nil {
		common.CountSafe("DiscardFreshBlockB")
		fmt.Println("AcceptBlock1", newbl.BchBlock.Hash.String(), "-", e.Error())
		newbl.Conn.Misbehave("LocalAcceptBl1", 250)
	} else {

		// Debugging Output (Optional)
		if common.CFG.TextUI_DevDebug {
			println("block", newbl.BchBlock.Height, "accepted")
		}

		retryCachedBlocks = retry_cached_blocks()
	}
}

func HandleRpcBlock(msg *rpcapi.BchBlockSubmited) {
	common.CountSafe("RPCNewBlock")

	network.MutexRcv.Lock()
	rb := network.ReceivedBlocks[msg.BchBlock.Hash.BIdx()]
	network.MutexRcv.Unlock()
	if rb == nil {
		panic("Block " + msg.BchBlock.Hash.String() + " not in ReceivedBlocks map")
	}

	common.BchBlockChain.Unspent.AbortWriting()
	rb.TmQueue = time.Now()

	e, _, _ := common.BchBlockChain.CheckBlock(msg.BchBlock)
	if e == nil {
		e = common.BchBlockChain.AcceptBlock(msg.BchBlock)
		rb.TmAccepted = time.Now()
	}
	if e != nil {
		common.CountSafe("RPCBlockError")
		msg.Error = e.Error()
		msg.Done.Done()
		return
	}

	network.NetRouteInv(network.MSG_BLOCK, msg.BchBlock.Hash, nil)
	common.RecalcAverageBlockSize()

	common.CountSafe("RPCBlockOK")
	println("New mined block", msg.BchBlock.Height, "accepted OK in", rb.TmAccepted.Sub(rb.TmQueue).String())

	common.Last.Mutex.Lock()
	common.Last.Time = time.Now()
	common.Last.BchBlock = common.BchBlockChain.LastBlock()
	common.Last.Mutex.Unlock()

	msg.Done.Done()
}

func main() {
	var ptr *byte
	if unsafe.Sizeof(ptr) < 8 {
		fmt.Println("WARNING: Gocoin-cash client shall be build for 64-bit arch. It will likely crash now.")
	}

	fmt.Println("")
	fmt.Println("// ======================================================================")
	fmt.Println("// == Welcome to Gocoin-cash client version ==", gocoincash.Version)
	fmt.Println("")
	fmt.Println("// Credits:")
	fmt.Println("")
	fmt.Println("// Piotr Narewski, Gocoin Founder")
	fmt.Println("")
	fmt.Println("// Julian Smith, Direction + Development")
	fmt.Println("// Arsen Yeremin, Development")
	fmt.Println("// Sumanth Kumar, Development")
	fmt.Println("// Clayton Wong, Development")
	fmt.Println("// Liming Jiang, Development")
	fmt.Println("")
	fmt.Println("// Includes reference work of btsuite:")
	fmt.Println("")
	fmt.Println("// Copyright (c) 2013-2017 The btcsuite developers")
	fmt.Println("// Copyright (c) 2018 The bcext developers")
	fmt.Println("// Use of this source code is governed by an ISC")
	fmt.Println("// license that can be found in the LICENSE file.")
	fmt.Println("")
	fmt.Println("// Includes reference work of Bitcoin Core (https://github.com/bitcoin/bitcoin)")
	fmt.Println("// Includes reference work of Bitcoin-ABC (https://github.com/Bitcoin-ABC/bitcoin-abc)")
	fmt.Println("// Includes reference work of Bitcoin Unlimited (https://github.com/BitcoinUnlimited/BitcoinUnlimited/tree/BitcoinCash)")
	fmt.Println("// Includes reference work of gcash by Shuai Qi \"qshuai\" (https://github.com/bcext/gcash)")
	fmt.Println("// Includes reference work of gcash (https://github.com/gcash/bchd)")
	fmt.Println("")
	fmt.Println("// +Other contributors")
	fmt.Println("")
	fmt.Println("// Some rights of 3rd party, derivative and included works remain the")
	fmt.Println("// property of thier respective owners. All marks, brands and logos of")
	fmt.Println("// member groups remain the exclusive property of their owners and no")
	fmt.Println("// right or endorsement is conferred by reference to thier organization")
	fmt.Println("// or brand(s) by CCA.")
	fmt.Println("")
	fmt.Println("// Copyright © 2018. Counterparty Cash Association (CCA) Zug, CH.")
	fmt.Println("// All Rights Reserved. All work owned by CCA is herby released")
	fmt.Println("// under Creative Commons Zero (0) License.")
	fmt.Println("")
	fmt.Println("// ======================================================================")
	fmt.Println("")
	fmt.Println("        cccccccccc          pppppppppp")
	fmt.Println("      cccccccccccccc      pppppppppppppp")
	fmt.Println("    ccccccccccccccc    ppppppppppppppppppp")
	fmt.Println("   cccccc       cc    ppppppp        pppppp")
	fmt.Println("   cccccc          pppppppp          pppppp")
	fmt.Println("   cccccc        ccccpppp            pppppp")
	fmt.Println("   cccccccc    cccccccc    pppp    ppppppp")
	fmt.Println("    ccccccccccccccccc     ppppppppppppppp")
	fmt.Println("       cccccccccccc      pppppppppppppp")
	fmt.Println("         cccccccc        pppppppppppp")
	fmt.Println("                         pppppp")
	fmt.Println("                         pppppp")
	fmt.Println("")
	fmt.Println("// ======================================================================")

	runtime.GOMAXPROCS(runtime.NumCPU()) // It seems that Go does not do it by default

	// Disable Ctrl+C
	signal.Notify(common.KillChan, os.Interrupt, os.Kill)
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			fmt.Println("main panic recovered:", err.Error())
			fmt.Println(string(debug.Stack()))
			network.NetCloseAll()
			common.CloseBlockChain()
			peersdb.ClosePeerDB()
			sys.UnlockDatabaseDir()
			os.Exit(1)
		}
	}()

	common.InitConfig()

	if common.FLAG.SaveConfig {
		common.SaveConfig()
		fmt.Println("Configuration file saved")
		os.Exit(0)
	}

	if common.FLAG.VolatileUTXO {
		fmt.Println("WARNING! Using UTXO database in a volatile mode. Make sure to close the client properly (do not kill it!)")
	}

	if common.FLAG.TrustAll {
		fmt.Println("WARNING! Assuming all scripts inside new blocks to PASS. Verify the last block's hash when finished.")
	}

	fmt.Println("Starting Gocoin-cash client version ==", gocoincash.Version)
	fmt.Println("")
	hostInit() // This will create the DB lock file and keep it open

	os.RemoveAll(common.TempBlocksDir())
	common.MkTempBlocksDir()

	if common.FLAG.UndoBlocks > 0 {
		usif.Exit_now.Set()
	}

	if common.FLAG.Rescan && common.FLAG.VolatileUTXO {

		fmt.Println("UTXO database rebuilt complete in the volatile mode, so flush DB to disk and exit...")

	} else if !usif.Exit_now.Get() {

		common.RecalcAverageBlockSize()

		peersTick := time.Tick(5 * time.Minute)
		netTick := time.Tick(time.Second)

		reset_save_timer() // we wil do one save try after loading, in case if ther was a rescan

		peersdb.Testnet = common.Testnet
		peersdb.ConnectOnly = common.CFG.ConnectOnly
		peersdb.Services = common.Services
		peersdb.InitPeers(common.GocoinCashHomeDir)
		if common.FLAG.UnbanAllPeers {
			var keys []qdb.KeyType
			var vals [][]byte
			peersdb.PeerDB.Browse(func(k qdb.KeyType, v []byte) uint32 {
				peer := peersdb.NewPeer(v)
				if peer.Banned != 0 {
					fmt.Println("Unban", peer.NetAddr.String())
					peer.Banned = 0
					keys = append(keys, k)
					vals = append(vals, peer.Bytes())
				}
				return 0
			})
			for i := range keys {
				peersdb.PeerDB.Put(keys[i], vals[i])
			}

			fmt.Println(len(keys), "peers un-baned")
		}

		for k, v := range common.BchBlockChain.BchBlockIndex {
			network.ReceivedBlocks[k] = &network.OneReceivedBlock{TmStart: time.Unix(int64(v.Timestamp()), 0)}
		}
		network.LastCommitedHeader = common.Last.BchBlock

		if common.CFG.TXPool.SaveOnDisk {
			network.MempoolLoad2()
		}

		if common.CFG.TextUI_Enabled {
			go textui.MainThread()
		}

		if common.CFG.WebUI.Interface != "" {
			fmt.Println("Starting WebUI at", common.CFG.WebUI.Interface)
			go webui.ServerThread(common.CFG.WebUI.Interface)
		}

		if common.CFG.RPC.Enabled {
			go rpcapi.StartServer(common.RPCPort())
		}

		usif.LoadBlockFees()

		wallet.FetchingBalanceTick = func() bool {
			select {
			case rec := <-usif.LocksChan:
				common.CountSafe("DoMainLocks")
				rec.In.Done()
				rec.Out.Wait()

			case newtx := <-network.NetTxs:
				common.CountSafe("DoMainNetTx")
				network.HandleNetTx(newtx, false)

			case <-netTick:
				common.CountSafe("DoMainNetTick")
				network.NetworkTick()

			case on := <-wallet.OnOff:
				if !on {
					return true
				}

			default:
			}
			return usif.Exit_now.Get()
		}

		startup_ticks := 5 // give 5 seconds for finding out missing blocks
		if !common.FLAG.NoWallet {
			// snooze the timer to 10 seconds after startup_ticks goes down
			common.SetUint32(&common.WalletOnIn, 10)
		}

		for !usif.Exit_now.Get() {
			common.Busy()

			common.CountSafe("MainThreadLoops")
			for retryCachedBlocks {
				retryCachedBlocks = retry_cached_blocks()
				// We have done one per loop - now do something else if pending...
				if len(network.NetBlocks) > 0 || len(usif.UiChannel) > 0 {
					break
				}
			}

			// first check for priority messages; kill signal or a new block
			select {
			case <-common.KillChan:
				common.Busy()
				usif.Exit_now.Set()
				continue

			case newbl := <-network.NetBlocks:
				common.Busy()
				HandleNetBlock(newbl)

			case rpcbl := <-rpcapi.RpcBlocks:
				common.Busy()
				HandleRpcBlock(rpcbl)

			default: // timeout immediatelly if no priority message
			}

			common.Busy()

			select {
			case <-common.KillChan:
				common.Busy()
				usif.Exit_now.Set()
				continue

			case newbl := <-network.NetBlocks:
				common.Busy()
				HandleNetBlock(newbl)

			case rpcbl := <-rpcapi.RpcBlocks:
				common.Busy()
				HandleRpcBlock(rpcbl)

			case rec := <-usif.LocksChan:
				common.Busy()
				common.CountSafe("MainLocks")
				rec.In.Done()
				rec.Out.Wait()

			case <-SaveBlockChain.C:
				common.Busy()
				common.CountSafe("SaveBlockChain")
				if common.BchBlockChain.Idle() {
					common.CountSafe("ChainIdleUsed")
				}

			case newtx := <-network.NetTxs:
				common.Busy()
				common.CountSafe("MainNetTx")
				network.HandleNetTx(newtx, false)

			case <-netTick:
				common.Busy()
				common.CountSafe("MainNetTick")
				network.NetworkTick()

				if common.BchBlockChainSynchronized {
					if common.WalletPendingTick() {
						wallet.OnOff <- true
					}
					break // BlockChainSynchronized so never mind checking it
				}

				if network.HeadersReceived.Get() >= 15 && network.BchBlocksToGetCnt() == 0 &&
					len(network.NetBlocks) == 0 && network.CachedBlocksLen.Get() == 0 {
					// only when we have no pending blocks and rteceived header messages, startup_ticks can go down..
					if startup_ticks > 0 {
						startup_ticks--
						break
					}
					common.SetBool(&common.BchBlockChainSynchronized, true)
					reset_save_timer()
				} else {
					startup_ticks = 5 // snooze by 5 seconds each time we're in here
				}

			case cmd := <-usif.UiChannel:
				common.Busy()
				common.CountSafe("MainUICmd")
				cmd.Handler(cmd.Param)
				cmd.Done.Done()

			case <-peersTick:
				common.Busy()
				peersdb.ExpirePeers()
				usif.ExpireBlockFees()

			case on := <-wallet.OnOff:
				common.Busy()
				if on {
					wallet.LoadBalance()
				} else {
					wallet.Disable()
					common.SetUint32(&common.WalletOnIn, 0)
				}
			}
		}

		common.BchBlockChain.Unspent.HurryUp()
		wallet.UpdateMapSizes()
		network.NetCloseAll()
	}

	sta := time.Now()
	common.CloseBlockChain()
	if common.FLAG.UndoBlocks == 0 {
		network.MempoolSave(false)
	}
	fmt.Println("Blockchain closed in", time.Now().Sub(sta).String())
	peersdb.ClosePeerDB()
	usif.SaveBlockFees()
	sys.UnlockDatabaseDir()
	os.RemoveAll(common.TempBlocksDir())
}
