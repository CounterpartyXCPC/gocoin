package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/counterpartyxcpc/gocoin-cash"
	"github.com/counterpartyxcpc/gocoin-cash/lib/btc"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/utils"
)

func main() {
	fmt.Println("Gocoin FetchTx version", gocoin.Version)

	if len(os.Args) < 2 {
		fmt.Println("Specify transaction id on the command line (MSB).")
		return
	}

	txid := btc.NewUint256FromString(os.Args[1])
	rawtx := utils.GetTxFromWeb(txid)
	if rawtx == nil {
		fmt.Println("Error fetching the transaction")
	} else {
		ioutil.WriteFile(txid.String()+".tx", rawtx, 0666)
	}
}
