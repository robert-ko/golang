// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// OFB AES test vectors.

// See U.S. National Institute of Standards and Technology (NIST)
// Special Publication 800-38A, ``Recommendation for Block Cipher
// Modes of Operation,'' 2001 Edition, pp. 52-55.

package block

import (
	"bytes";
	"crypto/aes";
	"io";
	"testing";
)

type ofbTest struct {
	name	string;
	key	[]byte;
	iv	[]byte;
	in	[]byte;
	out	[]byte;
}

var ofbAESTests = []ofbTest{
	// NIST SP 800-38A pp 52-55
	ofbTest{
		"OFB-AES128",
		commonKey128,
		commonIV,
		commonInput,
		[]byte{
			0x3b, 0x3f, 0xd9, 0x2e, 0xb7, 0x2d, 0xad, 0x20, 0x33, 0x34, 0x49, 0xf8, 0xe8, 0x3c, 0xfb, 0x4a,
			0x77, 0x89, 0x50, 0x8d, 0x16, 0x91, 0x8f, 0x03, 0xf5, 0x3c, 0x52, 0xda, 0xc5, 0x4e, 0xd8, 0x25,
			0x97, 0x40, 0x05, 0x1e, 0x9c, 0x5f, 0xec, 0xf6, 0x43, 0x44, 0xf7, 0xa8, 0x22, 0x60, 0xed, 0xcc,
			0x30, 0x4c, 0x65, 0x28, 0xf6, 0x59, 0xc7, 0x78, 0x66, 0xa5, 0x10, 0xd9, 0xc1, 0xd6, 0xae, 0x5e,
		},
	},
	ofbTest{
		"OFB-AES192",
		commonKey192,
		commonIV,
		commonInput,
		[]byte{
			0xcd, 0xc8, 0x0d, 0x6f, 0xdd, 0xf1, 0x8c, 0xab, 0x34, 0xc2, 0x59, 0x09, 0xc9, 0x9a, 0x41, 0x74,
			0xfc, 0xc2, 0x8b, 0x8d, 0x4c, 0x63, 0x83, 0x7c, 0x09, 0xe8, 0x17, 0x00, 0xc1, 0x10, 0x04, 0x01,
			0x8d, 0x9a, 0x9a, 0xea, 0xc0, 0xf6, 0x59, 0x6f, 0x55, 0x9c, 0x6d, 0x4d, 0xaf, 0x59, 0xa5, 0xf2,
			0x6d, 0x9f, 0x20, 0x08, 0x57, 0xca, 0x6c, 0x3e, 0x9c, 0xac, 0x52, 0x4b, 0xd9, 0xac, 0xc9, 0x2a,
		},
	},
	ofbTest{
		"OFB-AES256",
		commonKey256,
		commonIV,
		commonInput,
		[]byte{
			0xdc, 0x7e, 0x84, 0xbf, 0xda, 0x79, 0x16, 0x4b, 0x7e, 0xcd, 0x84, 0x86, 0x98, 0x5d, 0x38, 0x60,
			0x4f, 0xeb, 0xdc, 0x67, 0x40, 0xd2, 0x0b, 0x3a, 0xc8, 0x8f, 0x6a, 0xd8, 0x2a, 0x4f, 0xb0, 0x8d,
			0x71, 0xab, 0x47, 0xa0, 0x86, 0xe8, 0x6e, 0xed, 0xf3, 0x9d, 0x1c, 0x5b, 0xba, 0x97, 0xc4, 0x08,
			0x01, 0x26, 0x14, 0x1d, 0x67, 0xf3, 0x7b, 0xe8, 0x53, 0x8f, 0x5a, 0x8b, 0xe7, 0x40, 0xe4, 0x84,
		},
	},
}

func TestOFB_AES(t *testing.T) {
	for _, tt := range ofbAESTests {
		test := tt.name;

		c, err := aes.NewCipher(tt.key);
		if err != nil {
			t.Errorf("%s: NewCipher(%d bytes) = %s", test, len(tt.key), err);
			continue;
		}

		for j := 0; j <= 5; j += 5 {
			var crypt bytes.Buffer;
			in := tt.in[0 : len(tt.in)-j];
			w := NewOFBWriter(c, tt.iv, &crypt);
			var r io.Reader = bytes.NewBuffer(in);
			n, err := io.Copy(w, r);
			if n != int64(len(in)) || err != nil {
				t.Errorf("%s/%d: OFBWriter io.Copy = %d, %v want %d, nil", test, len(in), n, err, len(in))
			} else if d, out := crypt.Bytes(), tt.out[0:len(in)]; !same(out, d) {
				t.Errorf("%s/%d: OFBWriter\ninpt %x\nhave %x\nwant %x", test, len(in), in, d, out)
			}
		}

		for j := 0; j <= 7; j += 7 {
			var plain bytes.Buffer;
			out := tt.out[0 : len(tt.out)-j];
			r := NewOFBReader(c, tt.iv, bytes.NewBuffer(out));
			w := &plain;
			n, err := io.Copy(w, r);
			if n != int64(len(out)) || err != nil {
				t.Errorf("%s/%d: OFBReader io.Copy = %d, %v want %d, nil", test, len(out), n, err, len(out))
			} else if d, in := plain.Bytes(), tt.in[0:len(out)]; !same(in, d) {
				t.Errorf("%s/%d: OFBReader\nhave %x\nwant %x", test, len(out), d, in)
			}
		}

		if t.Failed() {
			break
		}
	}
}
