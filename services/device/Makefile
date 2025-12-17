.PHONY: sqlc-generate proto-generate build run test clean

# Generate code from SQL using sqlc
sqlc-generate:
	@echo "Generating repository code from SQL..."
	sqlc generate

# Generate gRPC code from proto files
proto-generate:
	@echo "Generating gRPC code from proto files..."
	@mkdir -p proto/device
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--proto_path=./proto \
		proto/device/device.proto
	@echo "Moving generated files to correct location if needed..."
	@if [ -d proto/device/device ]; then \
		mv proto/device/device/*.pb.go proto/device/ 2>/dev/null || true; \
		rmdir proto/device/device 2>/dev/null || true; \
	fi
	@if [ ! -f proto/device/device.pb.go ] && [ -f device/device.pb.go ]; then \
		mv device/*.pb.go proto/device/ 2>/dev/null || true; \
		rmdir device 2>/dev/null || true; \
	fi

# Build the service
build: proto-generate
	@echo "Building device service..."
	go build -o bin/device-service ./cmd/server

# Run the service locally
run: build
	@echo "Running device service..."
	./bin/device-service

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Generate all code
generate: sqlc-generate proto-generate

