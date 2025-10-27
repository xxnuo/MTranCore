.PHONY: build-benchmark build-mt build-stresstest build-worker gen help proto test update-models update-wasm

BUILD_DIR := build

update-wasm:
	mkdir -p data
	curl -L -o data/translations-wasm.json https://github.com/mozilla-firefox/firefox/raw/refs/heads/main/services/settings/dumps/main/translations-wasm.json
	filepath=$$(cat data/translations-wasm.json | jq -r '.data[0].attachment.location'); \
	echo curl -L -o internal/wasm/bergamot-translator-worker.wasm https://firefox-settings-attachments.cdn.mozilla.net/$$filepath
	filehash=$$(cat data/translations-wasm.json | jq -r '.data[0].attachment.hash'); \
	sha256sum internal/wasm/bergamot-translator-worker.wasm | grep $$filehash || echo "File hash mismatch"

gen:
	@go generate ./...

proto:
	@echo "Generating proto code from proto files..."
	@mkdir -p proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/worker.proto
	@echo "Proto code generation completed successfully"

prepare: gen proto

build-worker:
	@echo "Building worker server..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/worker ./cmd/worker
	@echo "worker server built successfully"

build-benchmark:
	@echo "Building benchmark tool..."
	@go build -o $(BUILD_DIR)/benchmark ./cmd/benchmark
	@strip $(BUILD_DIR)/benchmark
	@echo "benchmark tool built successfully"

build-stresstest:
	@echo "Building stress test tool..."
	@go build -o $(BUILD_DIR)/stresstest ./cmd/stresstest
	@strip $(BUILD_DIR)/stresstest
	@echo "stress test tool built successfully"

build-mt:
	@echo "Building mt (command line translation tool)..."
	@go build -o $(BUILD_DIR)/mt ./cmd/mt
	@strip $(BUILD_DIR)/mt
	@echo "mt (command line translation tool) built successfully"

build: prepare build-worker build-benchmark build-stresstest build-mt

actions: build-worker

test:
	@echo "Running all tests..."
	@go test -v ./...
