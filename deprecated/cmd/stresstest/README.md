# Stress Test Tool

A tool for stress testing and load testing the translation server, supporting multi-protocol testing with real test data and detailed performance analysis.

## Features

- **Multi-Protocol Support**: Full coverage of HTTP, WebSocket, and gRPC protocols
- **Real Test Data**: Contains texts of different lengths and types (short sentences, long sentences, paragraphs, HTML content, technical text, etc.)
- **Detailed Performance Metrics**: Latency distribution (P50, P95, P99), throughput, failure rate, etc.
- **Beautiful Output Format**: Tabular display, clear and intuitive
- **Multiple Test Scenarios**: High concurrency, sustained load, memory stability, rapid reload, mixed workload

## Build

```bash
cd /Volumes/MacData/Users/xxnuo/projects/gobergamot
go build -o build/stresstest ./cmd/stresstest
```

## Usage

### Basic Usage

```bash
# Run all tests, test all protocols (default)
./build/stresstest -url http://localhost:8988 -model ./models/enzh

# Test all protocols
./build/stresstest -protocol all -url http://localhost:8988

# Test specific protocol
./build/stresstest -protocol http -url http://localhost:8988    # HTTP protocol
./build/stresstest -protocol grpc -url localhost:9090           # gRPC protocol
./build/stresstest -protocol ws -url http://localhost:8988      # WebSocket protocol

# Run specific test
./build/stresstest -test concurrency -c 50     # High concurrency test
./build/stresstest -test sustained -d 30s      # Sustained load test
./build/stresstest -test memory -n 1000        # Memory stability test
./build/stresstest -test reload -r 5           # Rapid reload test
./build/stresstest -test mixed                 # Mixed workload test
```

### Parameter Description

- `-url string`: Server URL (default: `http://localhost:8988`)
  - HTTP protocol: `http://localhost:8988`
  - gRPC protocol: `localhost:9090` (host:port format)
  - WebSocket protocol: `http://localhost:8988` (automatically converted to ws://)
- `-protocol string`: Communication protocol (default: `all`)
  - `all`: Test all protocols
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

## Test Type Description

### 1. High Concurrency Test

Tests server performance under high concurrent load by sending a large number of concurrent requests simultaneously.

**Test Content**:

- Uses multiple worker threads to send requests simultaneously
- Uses texts from real test dataset
- Collects latency data for all requests

### 2. Sustained Load Test

Maintains constant load for a specified time to test server stability and sustained performance.

**Test Content**:

- Runs continuously for specified duration
- Displays real-time progress and statistics
- Tests server stability during long-term operation

### 3. Memory Stability Test

Detects potential memory leak issues through numerous iterations.

**Test Content**:

- Executes a large number of translation requests sequentially
- Monitors performance change trends
- Detects signs of memory leaks

### 4. Rapid Reload Test

Tests the stability of rapid engine loading and unloading.

**Test Content**:

- Repeatedly loads the engine multiple times
- Performs translation verification after each load
- Tests the reliability of engine initialization

### 5. Mixed Workload Test

Simulates real-world scenarios by mixing translation requests, health checks, and readiness checks.

**Test Content**:

- Runs multiple types of requests simultaneously
- Simulates real application scenarios
- Tests performance when multiple operations occur concurrently

## Output Example

```text
╔════════════════════════════════════════════════════╗
║          Stress Test Configuration                 ║
╠════════════════════════════════════════════════════╣
║ Server URL: http://localhost:8988                  ║
║ Protocol: all                                      ║
║ Model Path: ./models/enzh                          ║
║ Test Type: concurrency                             ║
║ Test Dataset Size: 18                              ║
╚════════════════════════════════════════════════════╝

╔════════════════════════════════════════════════════╗
║  Starting Protocol Test: http                      ║
╚════════════════════════════════════════════════════╝

=== High Concurrency Test ===
Loading engine...
Engine loaded successfully!

Starting high concurrency test: 50 workers, 10 requests each

╔════════════════════════════════════════════════════╗
║  Test Results: High Concurrency Test               ║
╠════════════════════════════════════════════════════╣
║ Protocol: http                                     ║
╠════════════════════════════════════════════════════╣
║ Basic Statistics                                   ║
╟────────────────────────────────────────────────────╢
║   Total Requests:  500                             ║
║   Success:         500                             ║
║   Failed:          0                               ║
║   Duration:        8.234s                          ║
║   Throughput:      60.72 req/s                     ║
║   Failure Rate:    0.00%                           ║
║   Concurrent Load: 50 workers                      ║
╠════════════════════════════════════════════════════╣
║ Latency Statistics                                 ║
╟────────────────────────────────────────────────────╢
║   Min:             45ms                            ║
║   Max:             250ms                           ║
║   Average:         120ms                           ║
║   P50 (Median):    115ms                           ║
║   P95:             180ms                           ║
║   P99:             230ms                           ║
║   Sample Count:    500                             ║
╠════════════════════════════════════════════════════╣
║ Evaluation                                         ║
╟────────────────────────────────────────────────────╢
║   Status: Excellent - All requests succeeded!     ║
║   Performance: Good - P95 latency < 500ms          ║
╚════════════════════════════════════════════════════╝
```

## Performance Metrics Description

### Basic Statistics

- **Total Requests**: Total number of requests sent
- **Success**: Number of successful requests
- **Failed**: Number of failed requests
- **Duration**: Total test time
- **Throughput**: Requests per second (req/s)
- **Failure Rate**: Percentage of failed requests to total requests
- **Concurrent Load**: Number of concurrent worker threads

### Latency Statistics

- **Min**: Minimum latency among all requests
- **Max**: Maximum latency among all requests
- **Average**: Average latency of all requests
- **P50 (Median)**: 50% of requests have latency lower than this value
- **P95**: 95% of requests have latency lower than this value
- **P99**: 99% of requests have latency lower than this value
- **Sample Count**: Number of latency samples collected

## Evaluation Criteria

### Status Evaluation

- **Excellent**: Failure rate = 0%
- **Good**: Failure rate < 10%
- **Warning**: Failure rate ≥ 10%

### Performance Evaluation

- **Excellent**: P95 latency < 100ms
- **Good**: P95 latency < 500ms
- **Fair**: P95 latency < 1s
- **Poor**: P95 latency ≥ 1s

## Test Dataset

The tool uses 18 real test texts, including:

- **Short sentences**: "Hello", "Good morning", etc.
- **Medium-length sentences**: Regular conversation and descriptive text
- **Long sentences**: Professional text with complex structures
- **Paragraphs**: Complete paragraph text
- **HTML content**: Text containing HTML tags
- **Technical text**: Technical documentation and API descriptions

Each request cycles through these texts to ensure test authenticity and diversity.

## Protocol Description

### HTTP Protocol

Standard REST API, suitable for most scenarios. Provides the best compatibility.

### gRPC Protocol

High-performance RPC framework based on HTTP/2, suitable for low-latency scenarios and high-concurrency stress testing.

### WebSocket Protocol

Full-duplex communication, suitable for scenarios requiring persistent connections. Note: WebSocket connections remain open during testing.

## Notes

1. Ensure the server is started and accessible
2. Ensure model files exist at the specified path
3. Stress testing may require significant system resources; recommend running in test environment
4. Sustained load and mixed load tests will display real-time progress
5. Adjust concurrency count and test duration based on server configuration
6. Running all tests (`-test all`) may take a considerable amount of time
7. When using gRPC, the URL parameter should be in `host:port` format
8. When using WebSocket, the tool will automatically convert HTTP URL to WS URL
9. Using `-protocol all` will automatically test all three protocols
10. There will be a 2-second interval between protocols to ensure server state cleanup

## Usage Scenario Examples

### Quick Server Performance Check

```bash
./build/stresstest -test concurrency -c 20 -protocol http
```

### Long-term Stability Test

```bash
./build/stresstest -test sustained -d 5m -protocol all
```

### Memory Leak Detection

```bash
./build/stresstest -test memory -n 5000 -protocol http
```

### Comprehensive Stress Test (All Protocols, All Tests)

```bash
./build/stresstest -test all -protocol all
```

### Compare Different Protocol Performance

```bash
# Test high concurrency performance of three protocols separately
./build/stresstest -test concurrency -protocol http -c 100
./build/stresstest -test concurrency -protocol grpc -c 100
./build/stresstest -test concurrency -protocol ws -c 100
```

## Performance Optimization Recommendations

Based on test results, you can make the following optimizations:

1. **If throughput is low**: Consider increasing worker thread count or optimizing server resources
2. **If P95/P99 latency is high**: Check server load and network latency
3. **If failure rate is high**: Check server logs and resource usage
4. **If memory test shows performance degradation**: Memory leak may exist, need to check code
