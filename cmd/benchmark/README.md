# Benchmark Tool

A tool for measuring the performance of the MTranCore translation server.

## Build

```bash
cd /home/xxnuo/projects/MTranCore
go build -o benchmark ./cmd/benchmark
```

## Usage

### Basic Usage

```bash
# Run all benchmarks (default, using HTTP protocol)
./benchmark -url http://localhost:8988 -model ./models/enzh -n 100

# Use different protocols
./benchmark -protocol http -url http://localhost:8988 -n 100    # HTTP protocol
./benchmark -protocol grpc -url localhost:9090 -n 100            # gRPC protocol
./benchmark -protocol ws -url http://localhost:8988 -n 100       # WebSocket protocol

# Run specific tests
./benchmark -test compute -n 100      # Simple text translation
./benchmark -test html -n 100         # HTML translation
./benchmark -test long -n 100         # Long text translation
./benchmark -test parallel -c 10 -n 100  # Concurrent translation test
```

### Parameter Description

- `-url string`: Server URL (default: `http://localhost:8988`)
  - HTTP protocol: `http://localhost:8988`
  - gRPC protocol: `localhost:9090` (host:port format)
  - WebSocket protocol: `http://localhost:8988` (automatically converted to ws://)
- `-protocol string`: Communication protocol (default: `http`)
  - `http`: HTTP REST API
  - `grpc`: gRPC protocol
  - `ws`: WebSocket protocol
- `-model string`: Model directory path (default: `./models/enzh`)
- `-n int`: Number of iterations (default: `100`)
- `-c int`: Number of concurrent workers (default: `1`)
- `-test string`: Test type (default: `all`)
  - `all`: Run all tests
  - `compute`: Simple text translation test
  - `html`: HTML translation test
  - `long`: Long text translation test
  - `parallel`: Concurrent translation test

## Output Example

```
=== Benchmark Configuration ===
Server URL: http://localhost:8988
Protocol: http
Model Path: /path/to/models/enzh
Iterations: 100
Concurrency: 1
Test Type: all
==============================

Loading engine...
Engine loaded successfully!

Running benchmark: Simple Text Translation
Progress: 100/100

--- Simple Text Translation ---
Total Requests:  100
Successful:      100
Failed:          0
Duration:        5.234s
Avg Latency:     52.34ms
Min Latency:     45.12ms
Max Latency:     89.56ms
Throughput:      19.11 req/s
```

## Performance Metrics Description

- **Total Requests**: Total number of requests
- **Successful**: Number of successful requests
- **Failed**: Number of failed requests
- **Duration**: Total test duration
- **Avg Latency**: Average latency
- **Min Latency**: Minimum latency
- **Max Latency**: Maximum latency
- **Throughput**: Throughput (requests/second)

## Protocol Description

### HTTP Protocol

Standard REST API, suitable for most scenarios.

### gRPC Protocol

High-performance RPC framework based on HTTP/2, suitable for low-latency scenarios.

### WebSocket Protocol

Full-duplex communication, suitable for scenarios requiring persistent connections.

## Notes

1. Ensure the server is started and accessible
2. Ensure model files exist at the specified path
3. First run will automatically load the engine, which takes some time
4. Concurrent tests (`-c > 1`) may require higher system resources
5. When using gRPC, the URL parameter should be in `host:port` format
6. When using WebSocket, the tool automatically converts HTTP URL to WS URL
