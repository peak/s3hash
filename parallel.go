package s3hash

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

type work struct {
	io.Reader
	partNum int
}

type result struct {
	partNum int
	sum     []byte
	err     error
}

// ReaderAtSeeker is both io.ReaderAt and io.Seeker. os.File satisfies it. To satisfy this from a byte slice, io.NewSectionReader() can be used.
type ReaderAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// CalculateForFileInParallel calculates the S3 hash of a given file with the given chunk size and number of workers.
func CalculateForFileInParallel(ctx context.Context, filename string, chunkSize int64, numWorkers int) (sum string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return CalculateInParallel(ctx, f, chunkSize, numWorkers)
}

// CalculateInParallel calculates the S3 hash of a given readerSeekerAt with the given chunk size and number of workers.
// io.NewSectionReader() can be used to create input from a byte slice.
//
// Example:
//  data := []byte("test data")
//  rdr := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))
//  result, err := CalculateInParallel(context.Background(), rdr, g.chunkSize, runtime.NumCPU())
func CalculateInParallel(ctx context.Context, input ReaderAtSeeker, chunkSize int64, numWorkers int) (sum string, err error) {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	var dataSize int64
	dataSize, err = input.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	ch := make(chan work)
	results := make(chan result)

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker(ctx, &wg, ch, results)
	}

	resultMap := make(map[int][]byte)

	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for r := range results {
			if r.err != nil {
				if err == nil || err != context.Canceled {
					err = r.err
				}
				cancelFunc()
				return
			}

			resultMap[r.partNum] = r.sum
		}
	}()

	parts := 0
	for i := int64(0); i < dataSize; i += chunkSize {
		parts++

		length := chunkSize
		if i+chunkSize > dataSize {
			length = dataSize - i
		}

		select {
		case <-ctx.Done():
			if err != nil {
				return
			}
			err = ctx.Err()
			return
		case ch <- work{
			io.NewSectionReader(input, i, length),
			parts,
		}:
		}
	}
	close(ch)

	wg.Wait()
	close(results)
	resultWg.Wait()

	var sumOfSums []byte
	for i := 1; i <= parts; i++ {
		sum, ok := resultMap[i]
		if !ok || sum == nil {
			return "", fmt.Errorf("resultMap incomplete %d", i)
		}
		sumOfSums = append(sumOfSums, sum...)
	}

	var finalSum []byte

	if parts == 1 {
		finalSum = sumOfSums
	} else {
		h := md5.New()
		_, err := h.Write(sumOfSums)
		if err != nil {
			return "", err
		}
		finalSum = h.Sum(nil)
	}

	sumHex := hex.EncodeToString(finalSum)

	if parts > 1 {
		sumHex += "-" + strconv.Itoa(parts)
	}

	return sumHex, err

}

func worker(ctx context.Context, wg *sync.WaitGroup, ch chan work, results chan result) {
	defer wg.Done()

	for w := range ch {
		select {
		case <-ctx.Done():
			return
		case results <- singleWork(w):
		}
	}
}

func singleWork(w work) result {
	r := result{partNum: w.partNum}

	h := md5.New()
	if _, err := io.Copy(h, w); err != nil {
		r.err = err
		return r
	}

	r.sum = h.Sum(nil)
	return r
}
