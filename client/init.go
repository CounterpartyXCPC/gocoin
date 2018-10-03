package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_chain"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

func host_init() {

	fmt.Println("Init function called")

	common.GocoinCashHomeDir = common.CFG.Datadir + string(os.PathSeparator)

	common.Testnet = common.CFG.Testnet // So chaging this value would will only affect the behaviour after restart
	if common.CFG.Testnet {             // testnet3
		common.GenesisBlock = bch.NewUint256FromString("000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943")
		common.Magic = [4]byte{0x0B, 0x11, 0x09, 0x07}
		common.GocoinCashHomeDir += common.DataSubdir() + string(os.PathSeparator)
		common.MaxPeersNeeded = 2000
	} else {
		common.GenesisBlock = bch.NewUint256FromString("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
		common.Magic = [4]byte{0xF9, 0xBE, 0xB4, 0xD9}
		common.GocoinCashHomeDir += common.DataSubdir() + string(os.PathSeparator)
		common.MaxPeersNeeded = 5000
	}

	// Lock the folder
	os.MkdirAll(common.GocoinCashHomeDir, 0770)
	sys.LockDatabaseDir(common.GocoinCashHomeDir)

	common.SecretKey, _ = ioutil.ReadFile(common.GocoinCashHomeDir + "authkey")
	if len(common.SecretKey) != 32 {
		common.SecretKey = make([]byte, 32)
		rand.Read(common.SecretKey)
		ioutil.WriteFile(common.GocoinCashHomeDir+"authkey", common.SecretKey, 0600)
	}
	common.PublicKey = bch.Encodeb58(bch.PublicFromPrivate(common.SecretKey, true))
	fmt.Println("Public auth key:", common.PublicKey)

	__exit := make(chan bool)
	__done := make(chan bool)
	go func() {
		for {
			select {
			case s := <-common.KillChan:
				fmt.Println(s)
				bch_chain.AbortNow = true
			case <-__exit:
				__done <- true
				return
			}
		}
	}()

	if bch_chain.AbortNow {
		sys.UnlockDatabaseDir()
		os.Exit(1)
	}

	if common.CFG.Memory.UseGoHeap {
		fmt.Println("Using native Go heap with the garbage collector for UTXO records")
	} else {
		utxo.MembindInit()
	}

	fmt.Print(string(common.LogBuffer.Bytes()))
	common.LogBuffer = nil

	if bch.EC_Verify == nil {
		fmt.Println("Using native secp256k1 lib for EC_Verify (consider installing a speedup)")
	}

	ext := &bch_chain.NewChanOpts{
		UTXOVolatileMode: common.FLAG.VolatileUTXO,
		UndoBlocks:       common.FLAG.UndoBlocks,
		BchBlockMinedCB:  blockMined}

	sta := time.Now()
	common.BchBlockChain = bch_chain.NewChainExt(common.GocoinCashHomeDir, common.GenesisBlock, common.FLAG.Rescan, ext,
		&bch_chain.BchBlockDBOpts{
			MaxCachedBlocks: int(common.CFG.Memory.MaxCachedBlks),
			MaxDataFileSize: uint64(common.CFG.Memory.MaxDataFileMB) << 20,
			DataFilesKeep:   common.CFG.Memory.DataFilesKeep})
	if bch_chain.AbortNow {
		fmt.Printf("Blockchain opening aborted after %s seconds\n", time.Now().Sub(sta).String())
		common.BchBlockChain.Close()
		sys.UnlockDatabaseDir()
		os.Exit(1)
	}

	common.Last.BchBlock = common.BchBlockChain.LastBlock()
	common.Last.Time = time.Unix(int64(common.Last.BchBlock.Timestamp()), 0)
	if common.Last.Time.After(time.Now()) {
		common.Last.Time = time.Now()
	}

	common.LockCfg()
	common.ApplyLastTrustedBlock()
	common.UnlockCfg()

	if common.CFG.Memory.FreeAtStart {
		fmt.Print("Freeing memory... ")
		sys.FreeMem()
		fmt.Print("\r                  \r")
	}
	sto := time.Now()

	al, sy := sys.MemUsed()
	fmt.Printf("Blockchain open in %s.  %d + %d MB of RAM used (%d)\n",
		sto.Sub(sta).String(), al>>20, utxo.ExtraMemoryConsumed()>>20, sy>>20)

	common.StartTime = time.Now()
	__exit <- true
	_ = <-__done

}
