# Benchmark Tool

A comprehensive tool for measuring the performance of the MTranCore translation server across all protocols.

## Features

- ✅ **Multi-Protocol Support**: Test HTTP, gRPC, and WebSocket protocols
- ✅ **Real-World Test Cases**: 10 diverse test scenarios covering various content types
- ✅ **Detailed Statistics**: P50, P90, P95, P99 latency percentiles with consistency analysis
- ✅ **Protocol Comparison**: Comprehensive comparison with performance analysis
- ✅ **Sample Translation Output**: View actual translation results during testing
- ✅ **Real-Time Progress**: Live progress updates with current statistics
- ✅ **Warmup Support**: Pre-test warmup to stabilize performance
- ✅ **Concurrent Testing**: Test with multiple parallel workers
- ✅ **Performance Indicators**: Automatic categorization of throughput and consistency
- ✅ **Error Tracking**: Detailed error reporting with percentages and categorization

## Build

```bash
cd /home/xxnuo/projects/MTranCore
make benchmark
# or
go build -o build/benchmark ./cmd/benchmark
```

## Usage

### Quick Start

```bash
# Test all three protocols with all test cases (recommended)
./build/benchmark -protocol all -n 100

# Test specific protocol
./build/benchmark -protocol http -n 100
./build/benchmark -protocol grpc -url localhost:9090 -n 100
./build/benchmark -protocol ws -n 100

# Test with concurrency
./build/benchmark -protocol all -n 1000 -c 10

# Test specific scenario
./build/benchmark -protocol all -test "Medium Paragraph" -n 100
```

### Parameter Description

- `-url string`: Server URL (default: `http://localhost:8988`)
  - HTTP protocol: `http://localhost:8988`
  - gRPC protocol: `localhost:9090` (host:port format)
  - WebSocket protocol: `http://localhost:8988` (automatically converted to ws://)
  
- `-protocol string`: Communication protocol (default: `all`)
  - `all`: Test all three protocols and compare results
  - `http`: HTTP REST API
  - `grpc`: gRPC protocol
  - `ws`: WebSocket protocol
  
- `-model string`: Model directory path (default: `./models/enzh`)

- `-n int`: Number of iterations per test (default: `100`)

- `-c int`: Number of concurrent workers (default: `1`)

- `-warmup int`: Number of warmup requests before testing (default: `10`)

- `-test string`: Test type (default: `all`)
  - `all`: Run all test cases
  - `Short Greeting`: Simple conversational text
  - `News Headline`: Breaking news content
  - `Product Description`: E-commerce product copy
  - `Email Message`: Professional email communication
  - `Technical Article`: Technical documentation
  - `Legal Notice`: Legal and formal text
  - `HTML Article`: HTML formatted web content
  - `Medical Information`: Healthcare and medical terminology
  - `Customer Support`: Support communication
  - `Long Document`: Extended multi-paragraph content
  - `parallel`: Concurrent load test (only when `-c > 1`)

## Output Example

```
╔═══════════════════════════════════════════════════════════╗
║       🚀 Translation Service Benchmark Tool 🚀           ║
╚═══════════════════════════════════════════════════════════╝

═══════════════════ Configuration ═════════════════════════
Server URL:    http://localhost:8988
Protocol(s):   all
Model Path:    /path/to/models/enzh
Iterations:    100 per test
Concurrency:   1 workers
Warmup:        10 requests
Test Type:     all
═══════════════════════════════════════════════════════════


╔═══════════════════════════════════════════════════════════╗
║  Testing Protocol: HTTP                                   ║
╚═══════════════════════════════════════════════════════════╝

🔌 Establishing HTTP connection...
✅ Connection established successfully!

📦 Loading translation engine...
   Model path: /path/to/models/enzh
   Protocol: HTTP
✅ Engine loaded successfully in 1.23s!

🔥 Warming up with 10 requests...
   Progress: 10/10 - ✅ Completed in 523ms (Success: 10/10)

🚀 Starting benchmark tests...
   Test type: all
   Iterations per test: 100
   Concurrency: 1

📋 Running all 10 test cases:

═══════════════════ Test 1/10 ═══════════════════
📊 Running test: Short Greeting
   Text length: 26 chars, HTML: false
   Preview: Hello, how are you today?
   Progress: 100/100 | Success: 100.0% | Avg Latency: 52.34ms   
   Sample translation: 你好，你今天好吗？

┌───────────────────────────────────────────────────────────┐
│ Test: Short Greeting                                      │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                                 │
│   ✅ Successful:     100                                 │
│   ❌ Failed:         0                                   │
│   Success Rate:      100.00%                             │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    5.23s                               │
│   Throughput:        19.12 req/s                         │
│   Avg Time/Request:  52.30ms                             │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               45.12ms                             │
│   Average:           52.34ms                             │
│   Median (P50):      51.23ms                             │
│   P90:               58.45ms                             │
│   P95:               62.11ms                             │
│   P99:               69.87ms                             │
│   Max:               89.56ms                             │
│   Spread (Max-Min):  44.44ms                             │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Excellent (1.21x)                   │
│   Throughput Class:  Moderate                            │
└───────────────────────────────────────────────────────────┘

✅ All tests completed for HTTP protocol!


╔═══════════════════════════════════════════════════════════╗
║              📊 Protocol Comparison Summary               ║
╚═══════════════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 1: Short Greeting
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
HTTP         │  100.0% │    19.12/s │   52.34ms │   51.23ms │   62.11ms │   69.87ms 🏆
GRPC         │  100.0% │    20.74/s │   48.23ms │   47.12ms │   56.78ms │   64.32ms ⚡
WS           │  100.0% │    18.29/s │   54.67ms │   53.45ms │   65.43ms │   73.21ms

Performance Analysis:
  • GRPC is 13.4% faster than WS in throughput
  • GRPC has the best P95 latency: 56.78ms

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Overall Protocol Performance Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │ Avg Throughput │  Avg Latency │ Wins
─────────────┼──────────────┼──────────────┼──────────
HTTP         │    18.45/s │      54.23ms │ 3/10
GRPC         │    20.12/s │      49.87ms │ 5/10
WS           │    17.89/s │      56.45ms │ 2/10

✅ Benchmark completed successfully!
```

## Performance Metrics Description

### Request Statistics
- **Total Requests**: Total number of requests sent
- **Successful**: Number of requests completed successfully (with ✅)
- **Failed**: Number of failed requests (with ❌)
- **Success Rate**: Percentage of successful requests
- **Failure Rate**: Percentage of failed requests (shown if failures occurred)

### Timing Statistics
- **Total Duration**: Total time taken for all requests
- **Throughput**: Requests per second (req/s)
- **Avg Time/Request**: Average wall-clock time per request

### Latency Distribution
- **Min**: Minimum latency observed
- **Average**: Mean latency across all requests
- **Median (P50)**: 50th percentile - half of requests were faster
- **P90**: 90th percentile - 90% of requests were faster
- **P95**: 95th percentile - 95% of requests were faster
- **P99**: 99th percentile - 99% of requests were faster
- **Max**: Maximum latency observed
- **Spread (Max-Min)**: Range between fastest and slowest requests

### Performance Indicators
- **Consistency**: Based on P95/P50 ratio
  - Excellent: < 1.2x (very stable performance)
  - Good: 1.2-1.5x (stable performance)
  - Fair: 1.5-2.0x (moderate variance)
  - Poor: > 2.0x (high variance)
- **Throughput Class**: Categorization of request throughput
  - Excellent: > 100 req/s
  - High: 50-100 req/s
  - Good: 20-50 req/s
  - Moderate: 10-20 req/s
  - Low: < 10 req/s

### Protocol Comparison
- **🏆 Symbol**: Best overall performer (highest throughput + lowest latency)
- **⚡ Symbol**: Highest throughput for that test
- **🎯 Symbol**: Lowest average latency for that test
- **Performance Analysis**: Automatic calculation of performance differences
- **Overall Summary**: Cross-test averages and win counts per protocol

## Test Cases

The benchmark includes 10 realistic test scenarios covering diverse use cases:

1. **Short Greeting** (~26 chars)
   - `"Hello, how are you today?"`
   - Tests baseline performance with simple conversation

2. **News Headline** (~131 chars)
   - Breaking news about renewable energy
   - Tests short-form journalism content

3. **Product Description** (~224 chars)
   - Wireless headphone features and benefits
   - Tests e-commerce and marketing copy

4. **Email Message** (~338 chars)
   - Professional business email
   - Tests formal communication style

5. **Technical Article** (~470 chars)
   - Machine learning frameworks and challenges
   - Tests technical documentation with specialized terminology

6. **Legal Notice** (~324 chars)
   - Terms and conditions text
   - Tests legal language and formal writing

7. **HTML Article** (~321 chars)
   - Web development guide with HTML structure
   - Tests HTML parsing and content preservation

8. **Medical Information** (~298 chars)
   - Patient care and treatment planning
   - Tests healthcare and medical terminology

9. **Customer Support** (~336 chars)
   - Support team response message
   - Tests customer service communication

10. **Long Document** (~918 chars)
    - Technology impact on business operations
    - Tests performance with lengthy, complex content

11. **Parallel Concurrent Load** (only when `-c > 1`)
    - Tests system performance under concurrent load
    - Simulates multiple simultaneous users

## Protocol Description

### HTTP Protocol

- Standard REST API using JSON
- Best for: Simple integration, debugging, compatibility
- Pros: Easy to use, widely supported, human-readable
- Cons: Higher overhead compared to binary protocols

### gRPC Protocol

- High-performance RPC framework based on HTTP/2 and Protocol Buffers
- Best for: Low-latency requirements, high-throughput scenarios
- Pros: Binary protocol, streaming support, efficient serialization
- Cons: More complex setup, requires `.proto` definitions

### WebSocket Protocol

- Full-duplex communication over a single TCP connection
- Best for: Persistent connections, real-time applications
- Pros: Low overhead after connection, bidirectional communication
- Cons: Connection management complexity, not suitable for one-off requests
- Note: Creates connection pool (`-c` value) for concurrent testing

## Best Practices

### Choosing Test Parameters

- **Quick Test**: `-n 10 -warmup 5` for rapid iteration during development
- **Standard Test**: `-n 100 -warmup 10` for regular performance checks
- **Production Benchmark**: `-n 1000 -warmup 50` for accurate measurements
- **Load Test**: `-n 1000 -c 10` for stress testing

### Interpreting Results

1. **Look at P95/P99**: These are more important than average for user experience
2. **Compare Protocols**: Use `-protocol all` to find the best fit for your use case
3. **Watch Success Rate**: Should be 100% under normal conditions
4. **Check Max Latency**: Indicates worst-case performance

### Performance Optimization Tips

1. **Warmup is Important**: First requests are always slower due to cold start
2. **Network Matters**: Test from the same network as production for realistic results
3. **Concurrency Level**: Start with `-c 1`, then increase to find optimal throughput
4. **Test Realistic Workloads**: Use the test case that matches your actual use case

## Example Workflows

### Development Testing
```bash
# Quick sanity check during development
./build/benchmark -protocol http -n 10 -test "Short Greeting"

# Test specific content type
./build/benchmark -protocol grpc -n 50 -test "Product Description"
```

### Pre-Deployment Validation
```bash
# Comprehensive test before deploying
./build/benchmark -protocol all -n 500 -c 5 -warmup 50

# Test all protocols with detailed output
./build/benchmark -protocol all -n 200 -test all
```

### Protocol Selection
```bash
# Compare all protocols to choose the best one
./build/benchmark -protocol all -n 200 -test all

# Focus on specific workload
./build/benchmark -protocol all -n 300 -test "Email Message"
```

### Stress Testing
```bash
# Test system under heavy load
./build/benchmark -protocol http -n 10000 -c 50 -test parallel

# Test with realistic concurrent users
./build/benchmark -protocol all -n 5000 -c 20
```

### Content-Specific Testing
```bash
# Test HTML content handling
./build/benchmark -protocol all -n 100 -test "HTML Article"

# Test long-form content performance
./build/benchmark -protocol all -n 50 -test "Long Document"

# Test technical terminology
./build/benchmark -protocol grpc -n 200 -test "Technical Article"
```

## Troubleshooting

### "Failed to create client" Error
- Check that the server is running and accessible
- Verify the URL and port are correct
- For gRPC, ensure you're using `host:port` format (no http://)

### "Failed to load engine" Error
- Verify the model path exists and is accessible
- Check server logs for detailed error messages
- Ensure the server has sufficient memory

### Low Success Rate
- Check server capacity (CPU, memory)
- Reduce concurrency level (`-c`)
- Check network connectivity and stability

### Inconsistent Results
- Increase warmup requests (`-warmup`)
- Run more iterations (`-n`)
- Ensure no other processes are consuming resources

## Notes

1. **Server Required**: Ensure the translation server is running before benchmarking
2. **Model Files**: Verify model files exist at the specified path
3. **First Run**: Engine loading happens once per protocol and takes time
4. **System Resources**: Concurrent tests (`-c > 1`) may require significant resources
5. **URL Format**: 
   - HTTP: `http://localhost:8988`
   - gRPC: `localhost:9090` (host:port only)
   - WebSocket: `http://localhost:8988` (auto-converted to ws://)
6. **Connection Pool**: WebSocket creates a connection pool sized to `-c` value
7. **Fair Comparison**: When comparing protocols, use same `-n` and `-c` values
