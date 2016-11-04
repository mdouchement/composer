#!/bin/bash
#
# This script helps to build corss-platform binaries.
#
# Usage:
#
# docker run --rm -v $(pwd):/go/src/github.com/mdouchement/composer golang:1.7 /go/src/github.com/mdouchement/composer/build.sh

cd /go/src/github.com/mdouchement/composer

echo '##### Installing system dependencies'
apt-get update && apt-get upgrade -qy
apt-get install --no-install-recommends -qy build-essential git

echo '##### Installing dependencies'
go get github.com/Masterminds/glide
glide install

# Fix cross-compilation
export GOOS=$(uname -s)
export GOARCH=$(uname -m)

echo '##### Building application'
for GOOS in darwin linux windows; do
  for GOARCH in amd64; do
    echo "--> composer-${GOOS^}-$GOARCH"

    # `-ldflags="-s -w"` for stripped binary
    go build -ldflags="-s -w" -o composer-${GOOS^}-x86_64 *.go
  done
done
