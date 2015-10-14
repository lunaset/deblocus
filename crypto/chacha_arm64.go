// +build arm64
// +build cgo

package crypto

import (
	"crypto/cipher"
	"unsafe"
)

//#cgo CFLAGS: -O3 -Wall -I.
//#cgo linux LDFLAGS: -L${SRCDIR} -lcrypto_linux_arm64
//#include "openssl/chacha.h"
import "C"

type ChaCha struct {
	cState *C.chacha_state
	tState *chacha_state_t
}

type chacha_state_t struct {
	state  [16]uint32
	stream [_CHACHA_STREAM_SIZE / 4]uint32
	rounds uint
	offset uint
}

func NewChaCha(key, iv []byte, rounds uint) (cipher.Stream, error) {
	if ks := len(key); ks != CHACHA_KeySize {
		return nil, KeySizeError(ks)
	}
	ivLen := len(iv)
	switch {
	case ivLen < CHACHA_IVSize:
		return nil, KeySizeError(ivLen)
	case ivLen == CHACHA_IVSize:
	default:
		iv = iv[:CHACHA_IVSize]
	}

	var s chacha_state_t
	chacha_init(&s.state, key, iv)
	s.rounds = rounds

	cState := (*C.chacha_state)(unsafe.Pointer(&s))
	var chacha = &ChaCha{
		tState: &s,
		cState: cState,
	}
	return chacha, nil
}

func (c *ChaCha) XORKeyStream(dst, src []byte) {
	cIn := (*C.uint8_t)(unsafe.Pointer(&src[0]))
	cOut := (*C.uint8_t)(unsafe.Pointer(&dst[0]))
	C.CRYPTO_neon_chacha_xor(c.cState, cOut, cIn, C.size_t(len(dst)))
}

func (c *ChaCha) initStream(iv []byte) {
	stream := (*[_CHACHA_STREAM_SIZE]byte)(unsafe.Pointer(&c.tState.stream[0]))
	var x uint16
	iv = iv[:cap(iv)]
	for i := 0; i < 256; i++ {
		x = uint16(sbox0[i]) * uint16(iv[i%len(iv)])
		stream[2*i] = byte(x >> 8)
		stream[2*i+1] = byte(x)
	}
	buf := make([]byte, _CHACHA_STREAM_SIZE)
	c.XORKeyStream(buf, buf)
}
