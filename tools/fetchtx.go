package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/counterpartyxcpc/gocoin-cash"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/utils"
)

func main() {
	fmt.Println("Gocoin FetchTx version", gocoincash.Version)

	if len(os.Args) < 2 {
		fmt.Println("Specify transaction id on the command line (MSB).")
		return
	}

	txid := bch.NewUint256FromString(os.Args[1])
	rawtx := utils.GetTxFromWeb(txid)
	if rawtx == nil {
		fmt.Println("Error fetching the transaction")
	} else {
		ioutil.WriteFile(txid.String()+".tx", rawtx, 0666)
	}
}
