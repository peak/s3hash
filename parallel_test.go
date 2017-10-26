package s3hash

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"testing"
)

func TestGoldenParallel(t *testing.T) {
	for i, g := range golden {
		data := bytes.Repeat([]byte(g.genesis), g.numRepeats)
		rdr := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))
		result, err := CalculateInParallel(context.Background(), rdr, g.chunkSize, runtime.NumCPU())
		if err != nil {
			t.Fatalf("Error calculating golden #%v: %v", i, err)
		}
		if result != g.out {
			t.Fatalf("hash[%d](%s)(%d) = %s want %s", i, g.genesis, g.numRepeats, result, g.out)
		}
	}
}
