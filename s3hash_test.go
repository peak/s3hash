package s3hash

import (
	"bytes"
	"testing"
)

const bytesInMb = 1024 * 1024

type hashTest struct {
	out        string
	genesis    string
	numRepeats int
	chunkSize  int64
}

var golden = []hashTest{
	// Single-part run
	{"bf8043c1e6890929374ea8f19828acbb", "Time flies like an arrow; fruit flies like a banana", 1, bytesInMb},

	// Multipart run
	{"38a7e5991be21b577978abb001323b0a-20", "0123456789", 1e7, 5 * bytesInMb},
}

func TestGolden(t *testing.T) {
	for i, g := range golden {
		data := bytes.Repeat([]byte(g.genesis), g.numRepeats)
		rdr := bytes.NewReader(data)
		result, err := Calculate(rdr, g.chunkSize)
		if err != nil {
			t.Fatalf("Error calculating golden #%v: %v", i, err)
		}
		if result != g.out {
			t.Fatalf("hash[%d](%s)(%d) = %s want %s", i, g.genesis, g.numRepeats, result, g.out)
		}
	}
}
