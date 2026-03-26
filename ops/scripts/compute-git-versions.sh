#!/usr/bin/env bash
set -euo pipefail

# Computes GIT_VERSION for all OP Stack images based on git tags pointing at a commit.
# All services share a single version tag (e.g. v1.5.1).
# Outputs JSON mapping image names to their GIT_VERSION values.
#
# Usage:
#   GIT_COMMIT=$(git rev-parse HEAD) ./ops/scripts/compute-git-versions.sh
#
# Output format:
#   {"op-node":"v1.5.1","op-batcher":"v1.5.1",...}

GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse HEAD)}"

IMAGES=(
  "op-node"
  "op-batcher"
  "op-proposer"
  "gas-oracle"
)

echo "Checking git tags pointing at $GIT_COMMIT:" >&2
tags_at_commit=$(git tag --points-at "$GIT_COMMIT" || true)
echo "Tags at commit: $tags_at_commit" >&2

# All services share the same simple version tag (e.g. v1.5.1)
filtered_tags=$(echo "$tags_at_commit" | grep "^v" || true)
echo "Filtered version tags: $filtered_tags" >&2

if [ -z "$filtered_tags" ]; then
  SHARED_VERSION="untagged"
else
  sorted_tags=$(echo "$filtered_tags" | sort -V)
  echo "Sorted tags: $sorted_tags" >&2

  # prefer full release tag over "-rc" release candidate tag if both exist
  full_release_tag=$(echo "$sorted_tags" | grep -v -- "-rc" || true)
  if [ -z "$full_release_tag" ]; then
    SHARED_VERSION=$(echo "$sorted_tags" | tail -n 1)
  else
    SHARED_VERSION=$(echo "$full_release_tag" | tail -n 1)
  fi
fi

echo "GIT_VERSION (shared): $SHARED_VERSION" >&2

# Output as JSON — all images share the same version
JSON="{"
FIRST=true
for IMAGE_NAME in "${IMAGES[@]}"; do
  if [ "$FIRST" = true ]; then
    FIRST=false
  else
    JSON="${JSON},"
  fi
  JSON="${JSON}\"${IMAGE_NAME}\":\"${SHARED_VERSION}\""
done
JSON="${JSON}}"

echo "$JSON"

