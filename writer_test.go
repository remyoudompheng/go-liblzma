// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

var digits []byte

func init() {
	buf := new(bytes.Buffer)
	for i := 0; i < 1e6; i++ {
		fmt.Fprintf(buf, "%d\n", i)
	}
	digits = buf.Bytes()
}

func TestCompress(T *testing.T) {
	if testing.Short() {
		digits = digits[:int(1e5)]
	}
	inbuf := bytes.NewBuffer(digits)
	outbuf := new(bytes.Buffer)

	enc, _ := NewWriter(outbuf, LevelDefault)
	io.Copy(enc, inbuf)
	enc.Close()

	T.Logf("%d bytes written (compressed size: %d bytes)", len(digits), len(outbuf.Bytes()))
}

func TestIdentity(T *testing.T) {
	if testing.Short() {
		digits = digits[:int(1e5)]
	}
	inbuf := bytes.NewBuffer(digits)
	tempbuf := new(bytes.Buffer)
	outbuf := new(bytes.Buffer)

	enc, _ := NewWriter(tempbuf, LevelDefault)
	io.Copy(enc, inbuf)
	enc.Close()

	dec, _ := NewReader(tempbuf)
	io.Copy(outbuf, dec)
	dec.Close()

	if !bytes.Equal(digits, outbuf.Bytes()) {
		T.Fatalf("decompressed data not equal to input")
	}
}

// Benchmark compression at a given level.
func benchmarkCompress(B *testing.B, preset Preset) {
	B.SetBytes(int64(len(digits)))

	for i := 0; i < B.N; i++ {
		inbuf := bytes.NewBuffer(digits)
		outbuf := new(bytes.Buffer)
		enc, _ := NewWriter(outbuf, preset)
		io.Copy(enc, inbuf)
		enc.Close()
	}
}

func BenchmarkCompressLvl1(B *testing.B) {
	benchmarkCompress(B, Level1)
}
func BenchmarkCompressLvl3(B *testing.B) {
	benchmarkCompress(B, Level3)
}
func BenchmarkCompressLvl6(B *testing.B) {
	benchmarkCompress(B, Level6)
}
func BenchmarkCompressExtremeLvl3(B *testing.B) {
	benchmarkCompress(B, Level3|LevelExtreme)
}

func BenchmarkCompressSmallBufferLvl3(B *testing.B) {
	B.SetBytes(int64(len(digits)))

	for i := 0; i < B.N; i++ {
		inbuf := bytes.NewBuffer(digits)
		outbuf := new(bytes.Buffer)
		enc, _ := NewWriterCustom(outbuf, Level3, CheckCRC64, 4096)
		io.Copy(enc, inbuf)
		enc.Close()
	}
}
