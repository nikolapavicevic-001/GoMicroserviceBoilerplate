#!/bin/bash

# Proto generation script for Go Microservices Boilerplate

set -e

echo "Generating protobuf code..."

# Colors for output
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Directory containing proto files
PROTO_DIR="shared/proto"
OUTPUT_DIR="shared/proto/gen"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Find all .proto files and generate Go code
find "$PROTO_DIR" -name "*.proto" | while read -r proto_file; do
    echo "Generating code for $proto_file..."

    protoc \
        --go_out="$OUTPUT_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$OUTPUT_DIR" \
        --go-grpc_opt=paths=source_relative \
        --proto_path="$PROTO_DIR" \
        "$proto_file"
done

echo -e "${GREEN}✓ Protobuf code generation completed${NC}"

# Format generated code
echo "Formatting generated code..."
gofmt -s -w "$OUTPUT_DIR"

echo -e "${GREEN}✓ Done!${NC}"
