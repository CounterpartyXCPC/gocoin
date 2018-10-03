package gocoincash

// This file is only to make "go get" working

import (
	_ "github.com/dchest/siphash"
	_ "github.com/golang/snappy"
	_ "golang.org/x/crypto/ripemd160"
)

const Version = "195.V1(BCH)"
