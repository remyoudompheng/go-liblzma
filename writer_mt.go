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
	"runtime"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sync/errgroup"
)

// CompressorMt is for multi-threaded implementation of lzma wrapper
type CompressorMt struct {
	handle   []*C.lzma_stream
	writer   io.Writer
	cores    int
	preset   Preset
	checksum Checksum
	tmpBuf   [][]byte
	arrays   [][]byte
	buffers  []*bytes.Buffer
	partSize int
}

var _ io.WriteCloser = &CompressorMt{}

// NewWriterMt is for creation and initialization of supplementary stuff
func NewWriterMt(w io.Writer, preset Preset) (*CompressorMt, error) {

	enc := new(CompressorMt)
	enc.cores = runtime.GOMAXPROCS(0)
	enc.writer = w
	enc.preset = preset
	enc.checksum = CheckCRC64
	enc.partSize = DefaultPartSize

	enc.handle = make([]*C.lzma_stream, enc.cores)
	enc.tmpBuf = make([][]byte, enc.cores)

	enc.buffers = make([]*bytes.Buffer, enc.cores)
	enc.arrays = make([][]byte, enc.cores)

	for i := 0; i < enc.cores; i++ {
		enc.handle[i] = allocLzmaStream(enc.handle[i])
		enc.tmpBuf[i] = make([]byte, DefaultBufsize)
		enc.arrays[i] = make([]byte, 0, enc.partSize)
		enc.buffers[i] = bytes.NewBuffer(enc.arrays[i])
	}

	return enc, nil
}

// NewWriterCustomMt Initializes a XZ encoder with additional settings.
func NewWriterCustomMt(
	w io.Writer,
	preset Preset,
	check Checksum,
	threadsNum int,
	bufSize int,
	partSize int) (*CompressorMt, error) {

	enc := new(CompressorMt)
	enc.cores = runtime.GOMAXPROCS(0)

	if enc.cores > threadsNum && threadsNum > 0 {
		enc.cores = threadsNum
	}

	enc.writer = w
	enc.preset = preset
	enc.partSize = partSize

	enc.handle = make([]*C.lzma_stream, enc.cores)
	enc.tmpBuf = make([][]byte, enc.cores)

	enc.buffers = make([]*bytes.Buffer, enc.cores)
	enc.arrays = make([][]byte, enc.cores)

	for i := 0; i < enc.cores; i++ {
		enc.handle[i] = allocLzmaStream(enc.handle[i])
		enc.tmpBuf[i] = make([]byte, bufSize)
		enc.arrays[i] = make([]byte, 0, enc.partSize)
		enc.buffers[i] = bytes.NewBuffer(enc.arrays[i])
	}

	return enc, nil
}

// CompressThread uses lzma instance to compress and finish for each part
func CompressThread(
	stream *C.lzma_stream,
	inBuf []byte,
	inOffset int,
	inLen int,
	preset Preset,
	checksum Checksum,
	outBuf *bytes.Buffer,
	localBuffer []byte,
	outNum *int64) error {

	ret := C.lzma_easy_encoder(stream, C.uint32_t(preset), C.lzma_check(checksum))
	if Errno(ret) != Ok {
		return Errno(ret)
	}

	totalCount := 0

	for totalCount < inLen {

		stream.avail_in = C.size_t(inLen - totalCount)
		stream.avail_out = C.size_t(len(localBuffer))

		ret := C.go_lzma_code(
			stream,
			unsafe.Pointer(&inBuf[inOffset+totalCount]),
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
			return Errno(ret)
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
			atomic.AddInt64(outNum, int64(totalCount))
			return nil
		}
	}
}

// Write splits input to peaces and compress it
// This can be improved by parallelizing file io
func (enc *CompressorMt) Write(in []byte) (n int, er error) {

	if len(in) <= 0 {
		return 0, nil
	}

	threads := new(errgroup.Group)

	var outNum int64

	partSize := enc.partSize

	if len(in) < partSize {
		partSize = len(in)
	}

	for partNum := 0; partNum < len(in)/partSize; partNum++ {

		offsetPart := partSize * partNum

		// last part with uncommon size
		if partNum == len(in)/partSize-1 {
			partSize = len(in) - partSize*partNum
		}

		for cycle := 0; cycle < enc.cores; cycle++ {

			offsetCycle := (partSize / enc.cores) * cycle

			lenCycle := partSize / enc.cores
			if cycle == enc.cores-1 {
				lenCycle = partSize - lenCycle*cycle
			}

			id := cycle

			threads.Go(func() error {
				return CompressThread(
					enc.handle[id],
					in,
					offsetPart+offsetCycle,
					lenCycle,
					enc.preset,
					enc.checksum,
					enc.buffers[id],
					enc.tmpBuf[id],
					&outNum)
			})
		}

		if err := threads.Wait(); err != nil {
			return 0, err
		}

		for idx, buffer := range enc.buffers {
			if _, err := buffer.WriteTo(enc.writer); err != nil {
				return 0, err
			}
			enc.arrays[idx] = enc.arrays[idx][:0]
		}
	}

	n = int(atomic.LoadInt64(&outNum))

	return n, nil
}

// Close frees any resources allocated by liblzma. It does not close the
// underlying reader.
func (enc *CompressorMt) Close() error {
	if enc != nil {
		for idx := range enc.handle {
			C.lzma_end(enc.handle[idx])
			C.free(unsafe.Pointer(enc.handle[idx]))
			enc.handle[idx] = nil
			enc.arrays[idx] = nil
			enc.buffers[idx] = nil
			enc.tmpBuf[idx] = nil
		}
	}
	return nil
}
