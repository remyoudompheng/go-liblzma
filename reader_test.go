package xz

import (
	"os"
	"testing"
)

func TestDecompress(T *testing.T) {
	f, er := os.Open("testdata/go_spec.html.xz")
  if er != nil { T.Fatalf("could not open test file: %s", er) }
	defer f.Close()
	dec, _ := NewReader(f, 4096)
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
