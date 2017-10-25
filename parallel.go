package s3hash

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sync"
)

type chunk struct {
	part   int
	start  int64
	length int64
}
type result struct {
	part int
	sum  []byte
	err  error
}

func CalculateForFileInParallel(ctx context.Context, filename string, chunkSize int64, numWorkers int) (sum string, err error) {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	var st os.FileInfo
	st, err = os.Stat(filename)
	if err != nil {
		return
	}
	dataSize := st.Size()

	var wg sync.WaitGroup
	ch := make(chan chunk)
	results := make(chan result)

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker(ctx, &wg, filename, ch, results)
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

			resultMap[r.part] = r.sum
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
		case ch <- chunk{
			parts,
			i,
			length,
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

func worker(ctx context.Context, wg *sync.WaitGroup, filename string, ch chan chunk, results chan result) {
	defer wg.Done()

	for c := range ch {
		select {
		case <-ctx.Done():
			return
		case results <- singleWork(filename, c):
		}
	}
}

func singleWork(filename string, c chunk) result {
	r := result{part: c.part}

	f, err := os.Open(filename)
	if err != nil {
		r.err = err
		return r
	}
	defer f.Close()

	sum, err := md5sum(f, c.start, c.length)
	r.sum = sum
	r.err = err
	return r
}
