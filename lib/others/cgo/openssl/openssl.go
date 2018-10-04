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

// File:		openssl.go
// Description:	Bictoin Cash Cash openssl Package

// Credits:

// Julian Smith, Direction, Development
// Arsen Yeremin, Development
// Sumanth Kumar, Development
// Clayton Wong, Development
// Liming Jiang, Development
// Piotr Narewski, Gocoin Founder

// Includes reference work of Shuai Qi "qshuai" (https://github.com/qshuai)

// Includes reference work of btsuite:

// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2018 The bcext developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// + Other contributors

// =====================================================================

package openssl

/*
#cgo !windows LDFLAGS: -lcrypto
#cgo windows LDFLAGS: /DEV/openssl-1.0.2a/libcrypto.a -lgdi32
#cgo windows CFLAGS: -I /DEV/openssl-1.0.2a/include

#include <stdio.h>
#include <openssl/ec.h>
#include <openssl/ecdsa.h>
#include <openssl/obj_mac.h>

int verify(void *pkey, unsigned int pkl, void *sign, unsigned int sil, void *hasz) {
	EC_KEY* ecpkey;
	ecpkey = EC_KEY_new_by_curve_name(NID_secp256k1);
	if (!ecpkey) {
		printf("EC_KEY_new_by_curve_name error!\n");
		return -1;
	}
	if (!o2i_ECPublicKey(&ecpkey, (const unsigned char **)&pkey, pkl)) {
		//printf("o2i_ECPublicKey fail!\n");
		return -2;
	}
	int res = ECDSA_verify(0, hasz, 32, sign, sil, ecpkey);
	EC_KEY_free(ecpkey);
	return res;
}
*/
import "C"
import "unsafe"

// Verify ECDSA signature
func EC_Verify(pkey, sign, hash []byte) int {
	return int(C.verify(unsafe.Pointer(&pkey[0]), C.uint(len(pkey)),
		unsafe.Pointer(&sign[0]), C.uint(len(sign)), unsafe.Pointer(&hash[0])))
}
