#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
BFF_DIR=$(dirname "$SCRIPT_DIR")
ROOT_DIR=$(dirname "$BFF_DIR")
DELIVERY_DIR="$ROOT_DIR/delivery-graphql"
OUTPUT_DIR="$BFF_DIR/subgraphs/delivery"
TMP_DIR=$(mktemp -d)

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT INT TERM

if ! command -v buf >/dev/null 2>&1; then
  echo "buf CLI is required to sync delivery subgraph artifacts" >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

cp "$DELIVERY_DIR/pkg/graph/schema.graphql" "$OUTPUT_DIR/schema.graphql"
cp "$DELIVERY_DIR/pkg/proto/service/v1/mapping.json" "$OUTPUT_DIR/mapping.json"

buf export "$DELIVERY_DIR" -o "$TMP_DIR"

if [ -f "$TMP_DIR/service/v1/service.proto" ]; then
  cp "$TMP_DIR/service/v1/service.proto" "$OUTPUT_DIR/service.proto"
elif [ -f "$TMP_DIR/pkg/proto/service/v1/service.proto" ]; then
  cp "$TMP_DIR/pkg/proto/service/v1/service.proto" "$OUTPUT_DIR/service.proto"
else
  echo "failed to locate exported delivery service.proto" >&2
  exit 1
fi

rm -rf "$OUTPUT_DIR/google"
if [ -d "$TMP_DIR/google" ]; then
  cp -R "$TMP_DIR/google" "$OUTPUT_DIR/google"
elif [ -d "$TMP_DIR/pkg/proto/google" ]; then
  cp -R "$TMP_DIR/pkg/proto/google" "$OUTPUT_DIR/google"
fi
