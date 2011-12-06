package xz

/*
#cgo LDFLAGS: -llzma
#include <lzma.h>
*/
import "C"

import (
	"io"
	"math"
	"os"
	"unsafe"
)

type Decompressor struct {
	handle C.lzma_stream
	rd     io.Reader
	buffer []byte
	offset int
}

var _ io.ReadCloser = &Decompressor{}

func NewReader(r io.Reader, bufsize int) (*Decompressor, os.Error) {
	dec := new(Decompressor)
	// The zero lzma_stream is the same thing as LZMA_STREAM_INIT.
	dec.rd = r
	dec.buffer = make([]byte, bufsize)
	dec.offset = bufsize

	// Initialize decoder
	ret := C.lzma_auto_decoder(&dec.handle, math.MaxUint64, 0)
	if Errno(ret) != Ok {
		return nil, Errno(ret)
	}

	return dec, nil
}

func (r *Decompressor) Read(out []byte) (out_count int, er os.Error) {
	if r.offset == len(r.buffer) {
		var n int
		n, er = r.rd.Read(r.buffer)
		if n == 0 {
			return 0, er
		}
		r.offset = 0
		r.handle.next_in = (*C.uint8_t)(unsafe.Pointer(&r.buffer[0]))
		r.handle.avail_in = C.size_t(n)
	}

	r.handle.next_out = (*C.uint8_t)(unsafe.Pointer(&out[0]))
	r.handle.avail_out = C.size_t(len(out))

	ret := C.lzma_code(&r.handle, C.lzma_action(Run))
	switch Errno(ret) {
	case Ok:
		break
	case StreamEnd:
		er = os.EOF
	default:
		er = Errno(ret)
	}

	r.offset = len(r.buffer) - int(r.handle.avail_in)

	return len(out) - int(r.handle.avail_out), er
}

func (r *Decompressor) Close() os.Error {
	if r != nil {
		C.lzma_end(&r.handle)
	}
	return nil
}
