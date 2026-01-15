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

mkdir -p .git/hooks

cat << 'EOF' > .git/hooks/pre-commit
#!/bin/sh

echo "Running pre-commit hook..."
make lint

if [ $? -ne 0 ]; then
  echo "Pre-commit checks failed. Commit aborted."
  exit 1
fi

echo "Pre-commit checks passed."
EOF

chmod +x .git/hooks/pre-commit

echo "Pre-commit hook create."
