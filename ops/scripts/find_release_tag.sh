#!/bin/bash

set -euo pipefail

# Function to print usage
usage() {
  echo "Usage: $0 <commit-hash>"
  echo "  <commit-hash> : The commit hash to check."
}

# Check for at least one argument
if [ "$#" -lt 1 ]; then
  usage
  exit 1
fi

commit_hash=$1

# Get all tags containing the commit, sorted by creation date
tags=$(git tag --contains "$commit_hash" --sort=taggerdate)

# Find the first release tag matching v*
for tag in $tags; do
  if [[ $tag == v* ]]; then
    echo "First release tag containing commit $commit_hash: $tag"
    exit 0
  fi
done

echo "Commit $commit_hash is not in any v* release tag."
