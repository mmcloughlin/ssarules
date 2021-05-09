#!/usr/bin/env bash

set -exuo pipefail

version="1.16.4"

archive=$(mktemp)
wget -O "${archive}" "https://golang.org/dl/go${version}.src.tar.gz"
tar --strip-components=7 -xvzf ${archive} 'go/src/cmd/compile/internal/ssa/gen/*.rules'
rm ${archive}

