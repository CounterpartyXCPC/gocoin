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

// File:		ver.go
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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/counterpartyxcpc/gocoin-cash/client/common"
	bch "github.com/counterpartyxcpc/gocoin-cash/lib/bch"
	"github.com/counterpartyxcpc/gocoin-cash/lib/others/sys"
)

var IgnoreExternalIpFrom = []string{}

func (c *OneConnection) SendVersion() {
	b := bytes.NewBuffer([]byte{})

	binary.Write(b, binary.LittleEndian, uint32(common.Version))
	binary.Write(b, binary.LittleEndian, uint64(common.Services))
	binary.Write(b, binary.LittleEndian, uint64(time.Now().Unix()))

	b.Write(c.PeerAddr.NetAddr.Bytes())
	if ExternalAddrLen() > 0 {
		b.Write(BestExternalAddr())
	} else {
		b.Write(bytes.Repeat([]byte{0}, 26))
	}

	b.Write(nonce[:])

	common.LockCfg()
	bch.WriteVlen(b, uint64(len(common.UserAgent)))
	b.Write([]byte(common.UserAgent))
	common.UnlockCfg()

	binary.Write(b, binary.LittleEndian, uint32(common.Last.BchBlockHeight()))
	if !common.GetBool(&common.CFG.TXPool.Enabled) {
		b.WriteByte(0) // don't notify me about txs
	}

	c.SendRawMsg("version", b.Bytes())
}

func (c *OneConnection) IsGocoin() bool {
	return strings.HasPrefix(c.Node.Agent, "/Gocoin:")
}

func (c *OneConnection) HandleVersion(pl []byte) error {
	if len(pl) >= 80 /*Up to, includiong, the nonce */ {
		if bytes.Equal(pl[72:80], nonce[:]) {
			common.CountSafe("VerNonceUs")
			return errors.New("Connecting to ourselves")
		}

		// check if we don't have this nonce yet
		Mutex_net.Lock()
		for _, v := range OpenCons {
			if v != c {
				v.Mutex.Lock()
				yes := v.X.VersionReceived && bytes.Equal(v.Node.Nonce[:], pl[72:80])
				v.Mutex.Unlock()
				if yes {
					Mutex_net.Unlock()
					v.Mutex.Lock()
					/*println("Peer with nonce", hex.EncodeToString(pl[72:80]), "from", c.PeerAddr.Ip(),
					"already connected as ", v.ConnID, "from ", v.PeerAddr.Ip(), v.Node.Agent)*/
					v.Mutex.Unlock()
					common.CountSafe("VerNonceSame")
					return errors.New("Peer already connected")
				}
			}
		}
		Mutex_net.Unlock()

		c.Mutex.Lock()
		c.Node.Version = binary.LittleEndian.Uint32(pl[0:4])
		if c.Node.Version < MIN_PROTO_VERSION {
			c.Mutex.Unlock()
			return errors.New("Client version too low")
		}

		copy(c.Node.Nonce[:], pl[72:80])
		c.Node.Services = binary.LittleEndian.Uint64(pl[4:12])
		c.Node.Timestamp = binary.LittleEndian.Uint64(pl[12:20])
		c.Node.ReportedIp4 = binary.BigEndian.Uint32(pl[40:44])

		use_this_ip := sys.ValidIp4(pl[40:44])

		if len(pl) >= 86 {
			le, of := bch.VLen(pl[80:])
			of += 80
			c.Node.Agent = string(pl[of : of+le])
			of += le
			if len(pl) >= of+4 {
				c.Node.Height = binary.LittleEndian.Uint32(pl[of : of+4])
				c.X.GetBlocksDataNow = true
				of += 4
				if len(pl) > of && pl[of] == 0 {
					c.Node.DoNotRelayTxs = true
				}
			}
			c.X.IsGocoin = strings.HasPrefix(c.Node.Agent, "/Gocoin:")
		}
		c.X.VersionReceived = true
		c.Mutex.Unlock()

		if use_this_ip {
			if bytes.Equal(pl[40:44], c.PeerAddr.Ip4[:]) {
				if common.FLAG.Log {
					ExternalIpMutex.Lock()
					f, _ := os.OpenFile("badip_log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
					if f != nil {
						fmt.Fprintf(f, "%s: OWN IP from %s @ %s - %d\n",
							time.Now().Format("2006-01-02 15:04:05"),
							c.Node.Agent, c.PeerAddr.Ip(), c.ConnID)
						f.Close()
					}
					ExternalIpMutex.Unlock()
				}
				common.CountSafe("IgnoreExtIP-O")
				use_this_ip = false
			} else if len(pl) >= 86 && binary.BigEndian.Uint32(pl[66:70]) != 0 &&
				!bytes.Equal(pl[66:70], c.PeerAddr.Ip4[:]) {
				if common.FLAG.Log {
					ExternalIpMutex.Lock()
					f, _ := os.OpenFile("badip_log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
					if f != nil {
						fmt.Fprintf(f, "%s: BAD IP=%d.%d.%d.%d from %s @ %s - %d\n",
							time.Now().Format("2006-01-02 15:04:05"),
							pl[66], pl[67], pl[68], pl[69], c.Node.Agent, c.PeerAddr.Ip(), c.ConnID)
						f.Close()
					}
					ExternalIpMutex.Unlock()
				}
				common.CountSafe("IgnoreExtIP-B")
				use_this_ip = false
			}
		}

		if use_this_ip {
			ExternalIpMutex.Lock()
			if _, known := ExternalIp4[c.Node.ReportedIp4]; !known { // New IP
				use_this_ip = true
				for x, v := range IgnoreExternalIpFrom {
					if c.Node.Agent == v {
						use_this_ip = false
						common.CountSafe(fmt.Sprint("IgnoreExtIP", x))
						break
					}
				}
				if use_this_ip && common.IsListenTCP() && common.GetExternalIp() == "" {
					fmt.Printf("New external IP %d.%d.%d.%d from ConnID=%d\n> ",
						pl[40], pl[41], pl[42], pl[43], c.ConnID)
				}
			}
			if use_this_ip {
				ExternalIp4[c.Node.ReportedIp4] = [2]uint{ExternalIp4[c.Node.ReportedIp4][0] + 1,
					uint(time.Now().Unix())}
			}
			ExternalIpMutex.Unlock()
		}

	} else {
		return errors.New("version message too short")
	}
	c.SendRawMsg("verack", []byte{})
	return nil
}
