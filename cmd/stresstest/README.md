# Stress Test Tool

A tool for stress testing and load testing the MTranCore translation server.

## Build

```bash
cd /home/xxnuo/projects/MTranCore
go build -o stresstest ./cmd/stresstest
```

## Usage

### Basic Usage

```bash
# Run all stress tests (default, using HTTP protocol)
./stresstest -url http://localhost:8080 -model ./models/enzh

# Use different protocols
./stresstest -protocol http -url http://localhost:8080    # HTTP protocol
./stresstest -protocol grpc -url localhost:9090           # gRPC protocol
./stresstest -protocol ws -url http://localhost:8080      # WebSocket protocol

# Run specific tests
./stresstest -test concurrency -c 50     # High concurrency test
./stresstest -test sustained -d 30s      # Sustained load test
./stresstest -test memory -n 1000        # Memory stability test
./stresstest -test reload -r 5           # Rapid reload test
./stresstest -test mixed                 # Mixed workload test
```

### Parameter Description

- `-url string`: Server URL (default: `http://localhost:8080`)
  - HTTP protocol: `http://localhost:8080`
  - gRPC protocol: `localhost:9090` (host:port format)
  - WebSocket protocol: `http://localhost:8080` (automatically converted to ws://)
- `-protocol string`: Communication protocol (default: `http`)
  - `http`: HTTP REST API
  - `grpc`: gRPC protocol
  - `ws`: WebSocket protocol
- `-model string`: Model directory path (default: `./models/enzh`)
- `-test string`: Test type (default: `all`)
  - `all`: Run all tests
  - `concurrency`: High concurrency test
  - `sustained`: Sustained load test
  - `memory`: Memory stability test
  - `reload`: Rapid engine reload test
  - `mixed`: Mixed workload test
- `-c int`: Number of concurrent workers for high concurrency test (default: `50`)
- `-d duration`: Duration for sustained load test (default: `30s`)
- `-n int`: Number of iterations for memory stability test (default: `1000`)
- `-r int`: Number of reloads for rapid reload test (default: `5`)

## Test Type Descriptions

### 1. High Concurrency Test
Tests server performance under high concurrent load by sending many concurrent requests simultaneously.

### 2. Sustained Load Test
Maintains constant load for a specified time to test server stability and sustained performance.

### 3. Memory Stability Test
Detects potential memory leaks through numerous iterations.

### 4. Rapid Reload Test
Tests stability of rapid engine loading and unloading.

### 5. Mixed Workload Test
Simulates real scenarios with mixed translation requests, health checks, and ready checks.

## Output Example

```
=== Stress Test Configuration ===
Server URL: http://localhost:8080
Protocol: http
Model Path: /path/to/models/enzh
Test Type: concurrency
=================================

=== High Concurrency Test ===
Loading engine...
Engine loaded successfully!

Starting high concurrency test: 50 workers, 10 requests each

--- High Concurrency Result ---
Total Requests:  500
Successful:      495
Failed:          5
Duration:        8.234s
Throughput:      60.72 req/s
Failure Rate:    1.00%
Concurrent Load: 50 workers

✓ Acceptable failure rate (<10%)
```

## Performance Metrics Description

- **Total Requests**: Total number of requests
- **Successful**: Number of successful requests
- **Failed**: Number of failed requests
- **Duration**: Total test duration
- **Throughput**: Throughput (requests/second)
- **Failure Rate**: Failure rate
- **Concurrent Load**: Concurrent load (number of workers)

## Evaluation Criteria

- ✅ **Excellent**: Failure rate = 0%
- ✓ **Acceptable**: Failure rate < 10%
- ⚠️ **Warning**: Failure rate ≥ 10%

## Protocol Description

### HTTP Protocol
Standard REST API, suitable for most scenarios.

### gRPC Protocol
High-performance RPC framework based on HTTP/2, suitable for low-latency scenarios and high-concurrency stress testing.

### WebSocket Protocol
Full-duplex communication, suitable for scenarios requiring persistent connections. Note: WebSocket connections remain open during testing.

## Notes

1. Ensure the server is started and accessible
2. Ensure model files exist at the specified path
3. Stress testing may require significant system resources; recommended to run in a test environment
4. Sustained load and mixed workload tests display real-time progress
5. Adjust concurrency and test duration based on server configuration
6. Running all tests (`-test all`) may take a considerable amount of time
7. When using gRPC, the URL parameter should be in `host:port` format
8. When using WebSocket, the tool automatically converts HTTP URL to WS URL
9. Under WebSocket protocol, concurrent tests share a single connection, which may affect test results

## Example Scenarios

### Quick Server Performance Check
```bash
./stresstest -test concurrency -c 20
```

### Long-term Stability Test
```bash
./stresstest -test sustained -d 5m
```

### Memory Leak Detection
```bash
./stresstest -test memory -n 5000
```

### Comprehensive Stress Test
```bash
./stresstest -test all
```
