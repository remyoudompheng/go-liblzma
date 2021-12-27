package xz

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

var digitsMt []byte

const shortSizeMt = int(10e6)

func init() {
	buf := new(bytes.Buffer)
	for i := 0; i < 5e6; i++ {
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

func TestIdentitySmallBufferMt(t *testing.T) {
	d := []byte{0x55}
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

	// log.Println(d, out)

	if !bytes.Equal(d, out) {
		t.Fatalf("decompressed data not equal to input")
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

func BenchmarkCompressSmallBufferLvl3Mt(B *testing.B) {
	B.SetBytes(int64(len(digits)))

	for i := 0; i < B.N; i++ {
		outbuf := new(bytes.Buffer)
		enc, _ := NewWriterCustomMt(outbuf, Level3, CheckCRC64, 1, 4096, 4096)
		_, err := enc.Write(digits)
		if err != nil {
			B.Fatal(err)
		}
		enc.Close()
	}
}

func BenchmarkCompressBigBufferLvl6Mt(B *testing.B) {
	B.SetBytes(int64(len(digits)))

	for i := 0; i < B.N; i++ {
		outbuf := new(bytes.Buffer)
		enc, _ := NewWriterCustomMt(outbuf, Level6, CheckCRC64, 0, 1024*1024*4, 1024*1024*8)
		_, err := enc.Write(digits)
		if err != nil {
			B.Fatal(err)
		}
		enc.Close()
	}
}
