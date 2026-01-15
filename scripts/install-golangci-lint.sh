#!/bin/bash

# Bash 'Strict Mode'
# http://redsymbol.net/articles/unofficial-bash-strict-mode
set -euo pipefail

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

# Find the project root by looking for Makefile
ROOT_DIR="$(git rev-parse --show-toplevel 2>/dev/null || (cd "$DIR/.." && pwd))"

# Change into the root directory
cd "$ROOT_DIR"

# install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v2.3.0

./bin/golangci-lint --version

echo "Golangci-lint installed."
