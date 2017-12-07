![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)
![Tag](https://img.shields.io/github/tag/peakgames/s3hash.svg)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/peakgames/s3hash)
[![Go Report](https://goreportcard.com/badge/github.com/peakgames/s3hash)](https://goreportcard.com/report/github.com/peakgames/s3hash)

# s3hash #

Calculate/verify hash of an S3 object, given a file and chunk size.

# Purpose #

Files uploaded to Amazon S3 using the S3 multipart API will have unique `ETag` values depending on their contents **and** the chunk size used to upload the file. This package calculates what the `ETag` will be using local file contents, which is useful for:

- Comparing local and remote files without downloading them again
- Verifying an S3 upload by getting the `ETag` of the uploaded file and comparing it to the one generated locally.

This will work on all types of S3 objects, regardless of whether they're uploaded using the multipart API or not.

## Installation ##

    go get -u github.com/peakgames/s3hash/cmd/s3hash

This will install `s3hash` in your `$GOPATH/bin` directory.

### Build ###

To build, just run:

    go build ./cmd/s3hash


## Usage ##

    Usage: s3hash [OPTION]... <chunk size in MB> <file>

      -e string
            Verify the S3 hash of file
      -p int
            Use NUM workers to run in parallel (default: number of cores)

## Examples ##

Get hash of local file, to be uploaded to S3 using 15 MB chunks:

    $ s3hash 15 filename.gz
    adf101740e60ba411adb21d2c50feb64-3

Verify hash of local file

    $ s3hash -e adf101740e60ba411adb21d2c50feb64-3 15 filename.gz
    OK
    (exit code 0)

    $ s3hash -e wronghash 15 filename.gz
    ERROR
    (exit code 1)
