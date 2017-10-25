package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/peakgames/s3hash"
)

const bytesInMb = 1024 * 1024

var (
	SL = log.New(os.Stdout, "", 0)
	EL = log.New(os.Stderr, "", 0)
)

func printUsageLine() {
	SL.Printf("Usage: %s [OPTION]... <chunk size in MB> <file>\n\n", os.Args[0])
}

func main() {

	hashToVerify := flag.String("e", "", "Verify the S3 hash of file")

	flag.Usage = func() {
		printUsageLine()
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	mb, err := strconv.Atoi(flag.Arg(0))
	if err != nil || mb < 5 {
		EL.Fatal("Please specify a valid chunk size")
	}

	chunkSize := int64(mb * bytesInMb)

	result, err := s3hash.CalculateForFile(flag.Arg(1), chunkSize)
	if err != nil {
		EL.Fatal(err)
	}

	if *hashToVerify != "" {
		if result == *hashToVerify {
			SL.Println("OK")
			return
		}  else {
			EL.Fatalln("ERROR")
		}
	}

	SL.Println(result)
}
