package ltc

import (
	"bytes"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/bch_utxo"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/utils"
)

const LTC_ADDR_VERSION = 48
const LTC_ADDR_VERSION_SCRIPT = 50

// LTC signing uses different seed string
func HashFromMessage(msg []byte, out []byte) {
	const MessageMagic = "Litecoin Signed Message:\n"
	b := new(bytes.Buffer)
	bch.WriteVlen(b, uint64(len(MessageMagic)))
	b.Write([]byte(MessageMagic))
	bch.WriteVlen(b, uint64(len(msg)))
	b.Write(msg)
	bch.ShaHash(b.Bytes(), out)
}

func AddrVerPubkey(testnet bool) byte {
	if !testnet {
		return LTC_ADDR_VERSION
	}
	return bch.AddrVerPubkey(testnet)
}

// At some point Litecoin started using addresses with M in front (version 50) - see github issue #41
func AddrVerScript(testnet bool) byte {
	if !testnet {
		return LTC_ADDR_VERSION_SCRIPT
	}
	return bch.AddrVerScript(testnet)
}

func NewAddrFromPkScript(scr []byte, testnet bool) (ad *bch.BtcAddr) {
	ad = bch.NewAddrFromPkScript(scr, testnet)
	if ad != nil && ad.Version == bch.AddrVerPubkey(false) {
		ad.Version = LTC_ADDR_VERSION
	}
	return
}

func GetUnspent(addr *bch.BtcAddr) (res utxo.AllUnspentTx) {
	var er error

	res, er = utils.GetUnspentFromBlockcypher(addr, "ltc")
	if er == nil {
		return
	}
	println("GetUnspentFromBlockcypher:", er.Error())

	return
}

// Download testnet's raw transaction from a web server
func GetTxFromWeb(txid *bch.Uint256) (raw []byte) {
	raw = utils.GetTxFromBlockcypher(txid, "ltc")
	if raw != nil && txid.Equal(bch.NewSha2Hash(raw)) {
		//println("GetTxFromBlockcypher - OK")
		return
	}

	return
}
