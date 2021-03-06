// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packet

import (
	"crypto/cipher"
	"crypto/openpgp/error"
	"crypto/sha1"
	"crypto/subtle"
	"hash"
	"io"
	"os"
	"strconv"
)

// SymmetricallyEncrypted represents a symmetrically encrypted byte string. The
// encrypted contents will consist of more OpenPGP packets. See RFC 4880,
// sections 5.7 and 5.13.
type SymmetricallyEncrypted struct {
	MDC      bool // true iff this is a type 18 packet and thus has an embedded MAC.
	contents io.Reader
	prefix   []byte
}

func (se *SymmetricallyEncrypted) parse(r io.Reader) os.Error {
	if se.MDC {
		// See RFC 4880, section 5.13.
		var buf [1]byte
		_, err := readFull(r, buf[:])
		if err != nil {
			return err
		}
		if buf[0] != 1 {
			return error.UnsupportedError("unknown SymmetricallyEncrypted version")
		}
	}
	se.contents = r
	return nil
}

// Decrypt returns a ReadCloser, from which the decrypted contents of the
// packet can be read. An incorrect key can, with high probability, be detected
// immediately and this will result in a KeyIncorrect error being returned.
func (se *SymmetricallyEncrypted) Decrypt(c CipherFunction, key []byte) (io.ReadCloser, os.Error) {
	keySize := c.keySize()
	if keySize == 0 {
		return nil, error.UnsupportedError("unknown cipher: " + strconv.Itoa(int(c)))
	}
	if len(key) != keySize {
		return nil, error.InvalidArgumentError("SymmetricallyEncrypted: incorrect key length")
	}

	if se.prefix == nil {
		se.prefix = make([]byte, c.blockSize()+2)
		_, err := readFull(se.contents, se.prefix)
		if err != nil {
			return nil, err
		}
	} else if len(se.prefix) != c.blockSize()+2 {
		return nil, error.InvalidArgumentError("can't try ciphers with different block lengths")
	}

	ocfbResync := cipher.OCFBResync
	if se.MDC {
		// MDC packets use a different form of OCFB mode.
		ocfbResync = cipher.OCFBNoResync
	}

	s := cipher.NewOCFBDecrypter(c.new(key), se.prefix, ocfbResync)
	if s == nil {
		return nil, error.KeyIncorrectError
	}

	plaintext := cipher.StreamReader{S: s, R: se.contents}

	if se.MDC {
		// MDC packets have an embedded hash that we need to check.
		h := sha1.New()
		h.Write(se.prefix)
		return &seMDCReader{in: plaintext, h: h}, nil
	}

	// Otherwise, we just need to wrap plaintext so that it's a valid ReadCloser.
	return seReader{plaintext}, nil
}

// seReader wraps an io.Reader with a no-op Close method.
type seReader struct {
	in io.Reader
}

func (ser seReader) Read(buf []byte) (int, os.Error) {
	return ser.in.Read(buf)
}

func (ser seReader) Close() os.Error {
	return nil
}

const mdcTrailerSize = 1 /* tag byte */ + 1 /* length byte */ + sha1.Size

// An seMDCReader wraps an io.Reader, maintains a running hash and keeps hold
// of the most recent 22 bytes (mdcTrailerSize). Upon EOF, those bytes form an
// MDC packet containing a hash of the previous contents which is checked
// against the running hash. See RFC 4880, section 5.13.
type seMDCReader struct {
	in          io.Reader
	h           hash.Hash
	trailer     [mdcTrailerSize]byte
	scratch     [mdcTrailerSize]byte
	trailerUsed int
	error       bool
	eof         bool
}

func (ser *seMDCReader) Read(buf []byte) (n int, err os.Error) {
	if ser.error {
		err = io.ErrUnexpectedEOF
		return
	}
	if ser.eof {
		err = os.EOF
		return
	}

	// If we haven't yet filled the trailer buffer then we must do that
	// first.
	for ser.trailerUsed < mdcTrailerSize {
		n, err = ser.in.Read(ser.trailer[ser.trailerUsed:])
		ser.trailerUsed += n
		if err == os.EOF {
			if ser.trailerUsed != mdcTrailerSize {
				n = 0
				err = io.ErrUnexpectedEOF
				ser.error = true
				return
			}
			ser.eof = true
			n = 0
			return
		}

		if err != nil {
			n = 0
			return
		}
	}

	// If it's a short read then we read into a temporary buffer and shift
	// the data into the caller's buffer.
	if len(buf) <= mdcTrailerSize {
		n, err = readFull(ser.in, ser.scratch[:len(buf)])
		copy(buf, ser.trailer[:n])
		ser.h.Write(buf[:n])
		copy(ser.trailer[:], ser.trailer[n:])
		copy(ser.trailer[mdcTrailerSize-n:], ser.scratch[:])
		if n < len(buf) {
			ser.eof = true
			err = os.EOF
		}
		return
	}

	n, err = ser.in.Read(buf[mdcTrailerSize:])
	copy(buf, ser.trailer[:])
	ser.h.Write(buf[:n])
	copy(ser.trailer[:], buf[n:])

	if err == os.EOF {
		ser.eof = true
	}
	return
}

func (ser *seMDCReader) Close() os.Error {
	if ser.error {
		return error.SignatureError("error during reading")
	}

	for !ser.eof {
		// We haven't seen EOF so we need to read to the end
		var buf [1024]byte
		_, err := ser.Read(buf[:])
		if err == os.EOF {
			break
		}
		if err != nil {
			return error.SignatureError("error during reading")
		}
	}

	// This is a new-format packet tag byte for a type 19 (MDC) packet.
	const mdcPacketTagByte = byte(0x80) | 0x40 | 19
	if ser.trailer[0] != mdcPacketTagByte || ser.trailer[1] != sha1.Size {
		return error.SignatureError("MDC packet not found")
	}
	ser.h.Write(ser.trailer[:2])

	final := ser.h.Sum()
	if subtle.ConstantTimeCompare(final, ser.trailer[2:]) == 1 {
		return error.SignatureError("hash mismatch")
	}
	return nil
}
