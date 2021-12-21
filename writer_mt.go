// Copyright 2011-2019 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

/*
#cgo LDFLAGS: -llzma
#include <lzma.h>
#include <stdlib.h>
int go_lzma_code(
    lzma_stream* handle,
    void* next_in,
    void* next_out,
    lzma_action action
);
*/
import "C"

import (
	"bytes"
	"io"
	"math"
	"runtime"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sync/errgroup"
)

// CompressorMt is for multi-threaded implementation of lzma wrapper
type CompressorMt struct {
	handle []*C.lzma_stream
	writer io.Writer
	cores  int
	preset Preset
}

var _ io.WriteCloser = &CompressorMt{}

// NewWriterMt is for creation and initialization of supplementary stuff
func NewWriterMt(w io.Writer, preset Preset) (*CompressorMt, error) {

	enc := new(CompressorMt)
	enc.cores = runtime.NumCPU()
	runtime.GOMAXPROCS(enc.cores)
	enc.writer = w
	enc.preset = preset

	enc.handle = make([]*C.lzma_stream, enc.cores)

	for i := 0; i < enc.cores; i++ {
		enc.handle[i] = allocLzmaStream(enc.handle[i])
	}

	return enc, nil
}

// CompressThread uses lzma instance to compress and finish for each part
func CompressThread(
	stream *C.lzma_stream,
	inBuf *byte,
	inLen int,
	preset Preset,
	outBuf *bytes.Buffer,
	localBuffer []byte,
	outNum *int64) error {

	ret := C.lzma_easy_encoder(stream, C.uint32_t(preset), C.lzma_check(CheckCRC64))
	if Errno(ret) != Ok {
		return Errno(ret)
	}

	totalCount := 0

	for totalCount < inLen {

		stream.avail_in = C.size_t(inLen - totalCount)
		stream.avail_out = C.size_t(len(localBuffer))

		ret := C.go_lzma_code(
			stream,
			// unsafe.Add(unsafe.Pointer(inBuf), totalCount),
			unsafe.Pointer(uintptr(unsafe.Pointer(inBuf))+uintptr(totalCount)),
			unsafe.Pointer(&localBuffer[0]),
			C.lzma_action(Run),
		)
		switch Errno(ret) {
		case Ok:
			break
		default:
			return Errno(ret)
		}

		totalCount = inLen - int(stream.avail_in)
		produced := len(localBuffer) - int(stream.avail_out)
		if _, err := outBuf.Write(localBuffer[:produced]); err != nil {
			return err
		}
	}

	for {
		stream.avail_out = C.size_t(len(localBuffer))
		stream.avail_in = 0
		ret := C.go_lzma_code(
			stream,
			nil,
			unsafe.Pointer(&localBuffer[0]),
			C.lzma_action(Finish),
		)
		if Errno(ret) != Ok && Errno(ret) != StreamEnd {
			return Errno(ret)
		}

		// Write back result.
		produced := len(localBuffer) - int(stream.avail_out)
		if _, err := outBuf.Write(localBuffer[:produced]); err != nil {
			return err
		}

		if Errno(ret) == StreamEnd {
			C.lzma_end(stream)
			atomic.AddInt64(outNum, int64(totalCount))
			return nil
		}
	}
}

// Write splits input to peaces and compress it
// This can be improved by parallelizing file io
func (enc *CompressorMt) Write(in []byte) (n int, er error) {

	threads := new(errgroup.Group)

	var outNum int64

	// reducing allocations
	tmpBuf := make([][]byte, enc.cores)
	for idx := range tmpBuf {
		tmpBuf[idx] = make([]byte, DefaultBufsize)
	}

	for cycle := 0; cycle < len(in)/DefaultPartSize+1; cycle++ {

		offsetCycle := DefaultPartSize * cycle

		var lenCycle int
		if cycle < len(in)/DefaultPartSize {
			lenCycle = DefaultPartSize
		} else {
			lenCycle = len(in) - DefaultPartSize*cycle
		}

		buffers := make([]bytes.Buffer, enc.cores)

		byteStep := lenCycle / enc.cores

		for i := 0; i < enc.cores; i++ {

			// reducing allocations, set max possible size
			buffers[i].Grow(DefaultPartSize)

			var inLen int
			inOffset := byteStep*i + offsetCycle

			if i < enc.cores-1 {
				inLen = int(math.Floor(float64(lenCycle) / float64(enc.cores)))
			} else {
				inLen = lenCycle - byteStep*i
			}

			id := i

			threads.Go(func() error {
				return CompressThread(
					enc.handle[id],
					&in[inOffset],
					inLen,
					enc.preset,
					&buffers[id],
					tmpBuf[id],
					&outNum)
			})
		}

		if err := threads.Wait(); err != nil {
			return 0, err
		}

		for _, buffer := range buffers {
			if _, err := buffer.WriteTo(enc.writer); err != nil {
				return 0, err
			}
		}
	}

	n = int(atomic.LoadInt64(&outNum))

	return n, nil
}

// Close frees any resources allocated by liblzma. It does not close the
// underlying reader.
func (enc *CompressorMt) Close() error {
	if enc != nil {
		for idx, stream := range enc.handle {
			C.free(unsafe.Pointer(stream))
			enc.handle[idx] = nil
		}
	}
	return nil
}
