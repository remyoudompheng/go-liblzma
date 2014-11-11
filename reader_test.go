// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

import (
	"bytes"
	"io"
	"os"
	"testing"
)

const testFile = "testdata/go_spec.html.xz"

func TestDecompress(T *testing.T) {
	f, er := os.Open(testFile)
	if er != nil {
		T.Fatalf("could not open test file: %s", er)
	}
	defer f.Close()
	dec, _ := NewReader(f)
	total := 0
	for {
		var buf [2048]byte
		n, er := dec.Read(buf[:])
		total += n
		if n == 0 || er != nil {
			T.Log(er)
			break
		}
	}
	T.Logf("Total %d bytes written", total)
}

func TestDecompressSmall(t *testing.T) {
	f, _ := os.Open(testFile)
	dec, _ := NewReader(f)
	buf := new(bytes.Buffer)
	io.Copy(buf, dec)
	contents := buf.Bytes()
	f.Close()

	f, _ = os.Open(testFile)
	dec, _ = NewReader(f)
	var contents2 []byte
	for {
		var buf [14]byte
		n, er := dec.Read(buf[:])
		contents2 = append(contents2, buf[:n]...)
		if n == 0 || er != nil {
			t.Log(er)
			break
		}
	}

	if !bytes.Equal(contents, contents2) {
		t.Fatalf("contents (%d bytes) and contents2 (%d bytes) differ!", len(contents), len(contents2))
	}
}

func TestDecompressConcatenated(t *testing.T) {
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test data file: %s", err)
	}
	defer f.Close()
	dec, err := NewReader(f)
	if err != nil {
		t.Fatalf("Failed to open decompressor: %s", err)
	}
	defer dec.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, dec)
	contents := buf.Bytes()

	f2a, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test data file: %s", err)
	}
	defer f2a.Close()
	f2b, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test data file: %s", err)
	}
	defer f2b.Close()
	f2 := io.MultiReader(f2a, f2b)
	dec2, err := NewReader(f2)
	if err != nil {
		t.Fatalf("Failed to open decompressor: %s", err)
	}
	defer dec2.Close()
	buf2 := new(bytes.Buffer)
	io.Copy(buf2, dec2)
	contents2 := buf2.Bytes()

	if 2*len(contents) != len(contents2) {
		t.Fatalf("contents2 has wrong length (expected %d bytes, got %d bytes)", 2*len(contents), len(contents2))
	}
	if !bytes.Equal(contents, contents2[:len(contents)]) {
		t.Errorf("contents does not match first half (%d bytes) of contents2", len(contents))
	}
	if !bytes.Equal(contents, contents2[len(contents):]) {
		t.Fatalf("contents does not match second half (%d bytes) of contents2", len(contents))
	}
}
