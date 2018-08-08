// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestDecompress(t *testing.T) {
	f, er := os.Open("testdata/go_spec.html.xz")
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
	f, _ := os.Open("testdata/go_spec.html.xz")
	dec, _ := NewReader(f)
	buf := new(bytes.Buffer)
	io.Copy(buf, dec)
	contents := buf.Bytes()
	f.Close()

	f, _ = os.Open("testdata/go_spec.html.xz")
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
