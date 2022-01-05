// Copyright 2011-2019 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

var digits []byte

const shortSize = int(10e6)

func init() {
	buf := new(bytes.Buffer)
	for i := 0; i < 5e6; i++ {
		fmt.Fprintf(buf, "%d\n", i*1234567891)
	}
	digits = buf.Bytes()
}

func TestCompress(t *testing.T) {
	d := digits
	if testing.Short() {
		d = d[:shortSize]
	}
	outbuf := new(bytes.Buffer)

	enc, err := NewWriter(outbuf, LevelDefault)
	_, err = enc.Write(d)
	if err != nil {
		t.Fatal(err)
	}
	enc.Close()

	t.Logf("%d bytes written (compressed size: %d bytes)", len(d), outbuf.Len())
}

func TestIdentity(t *testing.T) {
	d := digits
	if testing.Short() {
		d = d[:shortSize]
	}
	tempbuf := new(bytes.Buffer)

	enc, err := NewWriter(tempbuf, LevelDefault)
	_, err = enc.Write(d)
	if err != nil {
		t.Fatal(err)
	}
	enc.Close()

	t.Logf("testing %d bytes (compressed size: %d bytes)",
		len(d), tempbuf.Len())

	dec, _ := NewReader(tempbuf)
	out, err := ioutil.ReadAll(dec)
	dec.Close()
	if err != nil {
		t.Fatalf("read error: %s", err)
	}
	if !bytes.Equal(d, out) {
		t.Fatalf("decompressed data not equal to input")
	}
}

func TestIdentitySmallBuffer(t *testing.T) {
	d := []byte{0x55}
	tempbuf := new(bytes.Buffer)

	enc, err := NewWriter(tempbuf, LevelDefault)
	_, err = enc.Write(d)
	if err != nil {
		t.Fatal(err)
	}
	enc.Close()

	t.Logf("testing %d bytes (compressed size: %d bytes)",
		len(d), tempbuf.Len())

	dec, _ := NewReader(tempbuf)
	out, err := ioutil.ReadAll(dec)
	dec.Close()
	if err != nil {
		t.Fatalf("read error: %s", err)
	}

	if !bytes.Equal(d, out) {
		t.Fatalf("decompressed data not equal to input")
	}
}

// Benchmark compression at a given level.
func benchmarkCompress(B *testing.B, preset Preset) {
	B.SetBytes(int64(len(digits)))

	for i := 0; i < B.N; i++ {
		outbuf := new(bytes.Buffer)
		enc, _ := NewWriter(outbuf, preset)
		_, err := enc.Write(digits)
		if err != nil {
			B.Fatal(err)
		}
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
func BenchmarkCompressLvl9(B *testing.B) {
	benchmarkCompress(B, Level9)
}
func BenchmarkCompressExtremeLvl3(B *testing.B) {
	benchmarkCompress(B, Level3|LevelExtreme)
}

func BenchmarkCompressSmallBufferLvl3(B *testing.B) {
	B.SetBytes(int64(len(digits)))

	for i := 0; i < B.N; i++ {
		outbuf := new(bytes.Buffer)
		enc, _ := NewWriterCustom(outbuf, Level3, CheckCRC64, 4096, 1)
		_, err := enc.Write(digits)
		if err != nil {
			B.Fatal(err)
		}
		enc.Close()
	}
}
