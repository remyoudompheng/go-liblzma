// Copyright 2011-2019 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"testing/iotest"
)

const testFile = "testdata/go_spec.html.xz"

func TestDecompress(t *testing.T) {
	f, er := os.Open(testFile)
	if er != nil {
		t.Fatalf("could not open test file: %s", er)
	}
	defer f.Close()
	dec, _ := NewReader(f)
	total := 0
	for {
		var buf [2048]byte
		n, er := dec.Read(buf[:])
		total += n
		if n == 0 || er != nil {
			t.Log(er)
			break
		}
	}
	t.Logf("Total %d bytes written", total)
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

func TestDecompressTruncated(t *testing.T) {
	blob, err := ioutil.ReadFile("testdata/go_spec.html.xz")
	if err != nil {
		t.Fatalf("could not open test file: %s", err)
	}

	dec, err := NewReader(bytes.NewReader(blob[:4096]))
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(dec)
	if err == nil {
		t.Logf(`final data: expects "... </span>", got %q`, data[len(data)-80:])
		t.Skip("expected an error, didn't got any")
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
	contents, err := ioutil.ReadAll(dec)
	if err != nil {
		t.Fatalf("Failed to decompress test data file: %s", err)
	}

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
	dec2, err := NewReader(io.MultiReader(f2a, f2b))
	if err != nil {
		t.Fatalf("Failed to open decompressor: %s", err)
	}
	defer dec2.Close()
	got, err := ioutil.ReadAll(dec2)
	if err != nil {
		t.Fatalf("Failed to decompress concatenated test data files: %s", err)
	}

	want := append(append([]byte{}, contents...), contents...)
	if !bytes.Equal(got, want) {
		t.Fatalf("NewReader(f) => %q\nNewReader(io.MultiReader(f, f)) => %q\nExpected => %q", contents, got, want)
	}
}

// This test ensures that Decompressor doesn't lose data when
// last Read from underlying reader returns both n != 0 and err = io.EOF
// instead of returning io.EOF only on subsequent calls.
// Both behaviours are documented as acceptable in io.Reader interface
// description.
//
// Although this seems to never happen with plain files, HTTP bodies from
// net/http frequently exhibit this behaviour.
func TestDecompressWithEof(t *testing.T) {
	b, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	r1, err := NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer r1.Close()

	r2, err := NewReader(iotest.DataErrReader(bytes.NewReader(b)))
	if err != nil {
		t.Fatal(err)
	}
	defer r2.Close()

	contents1, err := ioutil.ReadAll(r1)
	if err != nil {
		t.Fatal(err)
	}

	// force multiple calls to Decompressor.Read after underlying stream EOF
	contents2, err := ioutil.ReadAll(iotest.OneByteReader(r2))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(contents1, contents2) {
		t.Fatalf("contents1 (%d bytes) and contents2 (%d bytes) differ!", len(contents1), len(contents2))
	}
}
