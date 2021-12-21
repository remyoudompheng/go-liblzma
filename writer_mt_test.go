package xz

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var digitsMt []byte

const shortSizeMt = int(4e6)

func init() {
	buf := new(bytes.Buffer)
	for i := 0; i < 2e6; i++ {
		fmt.Fprintf(buf, "%d\n", i*1234567891)
	}
	digitsMt = buf.Bytes()
}

func TestCompressMt(t *testing.T) {
	d := digitsMt
	if testing.Short() {
		d = d[:shortSizeMt]
	}
	outbuf := new(bytes.Buffer)

	enc, err := NewWriterMt(outbuf, LevelDefault)
	_, err = enc.Write(d)
	if err != nil {
		t.Fatal(err)
	}
	enc.Close()

	t.Logf("%d bytes written (compressed size: %d bytes)", len(d), outbuf.Len())
}

func TestIdentityMt(t *testing.T) {
	d := digitsMt
	if testing.Short() {
		d = d[:shortSizeMt]
	}
	tempbuf := new(bytes.Buffer)

	enc, err := NewWriterMt(tempbuf, LevelDefault)
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

func TestFileCompressDecompressMt(t *testing.T) {

	infile := "testdata/go_spec.pdf"

	inData, err := ioutil.ReadFile(infile)
	if err != nil {
		t.Fatalf("Cannot read the file: %s", err)
	}

	outbuf := new(bytes.Buffer)

	enc, err := NewWriterMt(outbuf, LevelDefault)
	_, err = enc.Write(inData)
	if err != nil {
		t.Fatal(err)
	}
	enc.Close()

	err = os.WriteFile(infile+".xz", outbuf.Bytes(), 0666)
	if err != nil {
		t.Fatalf("Cannot write compressed file: %s", err)
	}
	defer func() {
		e := os.Remove(infile + ".xz")
		if e != nil {
			t.Fatal(e)
		}
	}()

	b, err := ioutil.ReadFile(infile + ".xz")
	if err != nil {
		t.Fatalf("Cannot read compressed file: %s", err)
	}

	r, err := NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Cannot decompress data: %s", err)
	}

	if !bytes.Equal(inData, contents) {
		t.Fatalf("contents1 (%d bytes) and contents2 (%d bytes) differ!", len(inData), len(contents))
	}
}

// Benchmark compression at a given level.
func benchmarkCompressMt(B *testing.B, preset Preset) {
	B.SetBytes(int64(len(digits)))

	for i := 0; i < B.N; i++ {
		outbuf := new(bytes.Buffer)
		enc, _ := NewWriterMt(outbuf, preset)
		_, err := enc.Write(digits)
		if err != nil {
			B.Fatal(err)
		}
		enc.Close()
	}
}

func BenchmarkCompressLvl1Mt(B *testing.B) {
	benchmarkCompressMt(B, Level1)
}
func BenchmarkCompressLvl3Mt(B *testing.B) {
	benchmarkCompressMt(B, Level3)
}
func BenchmarkCompressLvl6Mt(B *testing.B) {
	benchmarkCompressMt(B, Level6)
}
func BenchmarkCompressLvl9Mt(B *testing.B) {
	benchmarkCompressMt(B, Level9)
}
func BenchmarkCompressExtremeLvl3Mt(B *testing.B) {
	benchmarkCompressMt(B, Level3|LevelExtreme)
}
