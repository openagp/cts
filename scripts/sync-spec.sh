#!/usr/bin/env bash
# Sync canonical schemas + test vectors from openagp/spec into the CTS's
# embedded copies. Run from the cts repo root with the spec repo checked out
# as a sibling:
#
#   /workspace/openagp/spec/  <- canonical source
#   /workspace/openagp/cts/   <- this repo
#
# Usage:
#   scripts/sync-spec.sh [path-to-spec-repo]
#
# Mirrors the equivalent sync-schemas.sh in sdk-python and sdk-typescript.
# CI fails if the embedded copies drift from the canonical spec.

set -euo pipefail

CTS_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC_DIR="${1:-${CTS_ROOT}/../spec}"

if [[ ! -d "${SPEC_DIR}/schemas" ]]; then
  echo "error: ${SPEC_DIR}/schemas not found" >&2
  echo "       pass the spec repo path as the first arg, or check it out at ../spec" >&2
  exit 2
fi

CHANGED=0

# Schemas → internal/schema/embed/
DEST="${CTS_ROOT}/internal/schema/embed"
mkdir -p "${DEST}"
for f in "${SPEC_DIR}/schemas"/*.json; do
  name="$(basename "$f")"
  if ! cmp -s "$f" "${DEST}/${name}"; then
    cp "$f" "${DEST}/${name}"
    echo "updated schema: ${name}"
    CHANGED=1
  fi
done

# Test vectors → internal/vectors/embed/
DEST="${CTS_ROOT}/internal/vectors/embed"
mkdir -p "${DEST}"
for f in "${SPEC_DIR}/test-vectors"/*.json; do
  name="$(basename "$f")"
  if ! cmp -s "$f" "${DEST}/${name}"; then
    cp "$f" "${DEST}/${name}"
    echo "updated vector: ${name}"
    CHANGED=1
  fi
done

if [[ "${CHANGED}" -eq 0 ]]; then
  echo "all embedded files in sync"
fi

if [[ "${1:-}" == "--check" ]] || [[ "${CI:-}" == "true" ]]; then
  exit "${CHANGED}"
fi
