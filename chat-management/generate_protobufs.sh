#!/bin/bash

PROTO_DIR="./src/proto"
OUT_DIR="./src/gen/go"

if [ ! -d "$OUT_DIR" ]; then
  mkdir -p "$OUT_DIR"
fi

for proto_file in $(find "$PROTO_DIR" -name '*.proto'); do
  protoc -I="$PROTO_DIR" --go_out="$OUT_DIR" --go_opt=paths=source_relative \
         --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative "$proto_file"
  echo "INFO: Generated go code for $proto_file"
done
