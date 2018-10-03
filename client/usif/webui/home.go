package webui

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	"github.com/counterpartyxcpc/gocoin-cash/client/network"
	"github.com/counterpartyxcpc/gocoin-cash/client/usif"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/peersdb"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

var (
	mutexHrate sync.Mutex
	lastHrate  float64
	nextHrate  time.Time
)

func p_home(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	// The handler also gets called for /favicon.ico
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	s := load_template("home.html")

	if !common.CFG.WebUI.ServerMode {
		common.LockCfg()
		dat, _ := json.MarshalIndent(&common.CFG, "", "    ")
		common.UnlockCfg()
		s = strings.Replace(s, "{CONFIG_FILE}", strings.Replace(string(dat), ",\"", ", \"", -1), 1)
	}

	s = strings.Replace(s, "<!--PUB_AUTH_KEY-->", common.PublicKey, 1)

	write_html_head(w, r)
	w.Write([]byte(s))
	write_html_tail(w)
}

func json_status(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	var out struct {
		Height                    uint32
		Hash                      string
		Timestamp                 uint32
		Received                  int64
		Time_now                  int64
		Diff                      float64
		Median                    uint32
		Version                   uint32
		MinValue                  uint64
		WalletON                  bool
		LastTrustedBlockHeight    uint32
		LastHeaderHeight          uint32
		BchBlockChainSynchronized bool
	}
	common.Last.Mutex.Lock()
	out.Height = common.Last.BchBlock.Height
	out.Hash = common.Last.BchBlock.BchBlockHash.String()
	out.Timestamp = common.Last.BchBlock.Timestamp()
	out.Received = common.Last.Time.Unix()
	out.Time_now = time.Now().Unix()
	out.Diff = bch.GetDifficulty(common.Last.BchBlock.Bits())
	out.Median = common.Last.BchBlock.GetMedianTimePast()
	out.Version = common.Last.BchBlock.BchBlockVersion()
	common.Last.Mutex.Unlock()
	out.MinValue = common.AllBalMinVal()
	out.WalletON = common.GetBool(&common.WalletON)
	out.LastTrustedBlockHeight = common.GetUint32(&common.LastTrustedBlockHeight)
	network.MutexRcv.Lock()
	out.LastHeaderHeight = network.LastCommitedHeader.Height
	network.MutexRcv.Unlock()
	out.BchBlockChainSynchronized = common.GetBool(&common.BchBlockChainSynchronized)

	bx, er := json.Marshal(out)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}
}

func json_system(w http.ResponseWriter, r *http.Request) {
	if !ipchecker(r) {
		return
	}

	var out struct {
		BchBlocks_cached   int
		BchBlocksToGet     int
		Known_peers        int
		Node_uptime        uint64
		Net_block_qsize    int
		Net_tx_qsize       int
		Heap_size          uint64
		Heap_sysmem        uint64
		Qdb_extramem       int64
		Ecdsa_verify_cnt   uint64
		Average_block_size int
		Average_fee        float64
		LastHeaderHeight   uint32
		NetworkHashRate    float64
		SavingUTXO         bool
	}

	out.BchBlocks_cached = network.CachedBlocksLen.Get()
	network.MutexRcv.Lock()
	out.BchBlocksToGet = len(network.BchBlocksToGet)
	network.MutexRcv.Unlock()
	out.Known_peers = peersdb.PeerDB.Count()
	out.Node_uptime = uint64(time.Now().Sub(common.StartTime).Seconds())
	out.Net_block_qsize = len(network.NetBlocks)
	out.Net_tx_qsize = len(network.NetTxs)
	out.Heap_size, out.Heap_sysmem = sys.MemUsed()
	out.Qdb_extramem = utxo.ExtraMemoryConsumed()
	out.Ecdsa_verify_cnt = bch.EcdsaVerifyCnt()
	out.Average_block_size = common.AverageBlockSize.Get()
	out.Average_fee = common.GetAverageFee()
	network.MutexRcv.Lock()
	out.LastHeaderHeight = network.LastCommitedHeader.Height
	network.MutexRcv.Unlock()

	mutexHrate.Lock()
	if nextHrate.IsZero() || time.Now().After(nextHrate) {
		lastHrate = usif.GetNetworkHashRateNum()
		nextHrate = time.Now().Add(time.Minute)
	}
	out.NetworkHashRate = lastHrate
	mutexHrate.Unlock()

	out.SavingUTXO = common.BchBlockChain.Unspent.WritingInProgress.Get()

	bx, er := json.Marshal(out)
	if er == nil {
		w.Header()["Content-Type"] = []string{"application/json"}
		w.Write(bx)
	} else {
		println(er.Error())
	}
}
