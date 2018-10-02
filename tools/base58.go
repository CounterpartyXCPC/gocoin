package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
)

var (
	decode = flag.Bool("d", false, "run base58 decode (instead of encode)")
	binary = flag.Bool("b", false, "binary (insted of hex) for decode output")
	help   = flag.Bool("h", false, "print this help")
)

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	msg, _ := ioutil.ReadAll(os.Stdin)
	if len(msg) == 0 {
		return
	}

	if *decode {
		res := bch.Decodeb58(strings.Trim(string(msg), " \t\n\r"))
		if *binary {
			os.Stdout.Write(res)
		} else {
			fmt.Println(hex.EncodeToString(res))
		}
	} else {
		fmt.Println(bch.Encodeb58(msg))
	}
}
