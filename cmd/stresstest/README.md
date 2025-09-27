# Stress Test Tool - 压力测试工具

用于对 MTranCore 翻译服务器进行压力测试和负载测试。

## 编译

```bash
cd /home/xxnuo/projects/MTranCore
go build -o stresstest ./cmd/stresstest
```

## 使用方法

### 基本用法

```bash
# 运行所有压力测试（默认，使用HTTP协议）
./stresstest -url http://localhost:8080 -model ./models/enzh

# 使用不同的协议
./stresstest -protocol http -url http://localhost:8080    # HTTP协议
./stresstest -protocol grpc -url localhost:9090           # gRPC协议
./stresstest -protocol ws -url http://localhost:8080      # WebSocket协议

# 运行特定测试
./stresstest -test concurrency -c 50     # 高并发测试
./stresstest -test sustained -d 30s      # 持续负载测试
./stresstest -test memory -n 1000        # 内存稳定性测试
./stresstest -test reload -r 5           # 快速重载测试
./stresstest -test mixed                 # 混合工作负载测试
```

### 参数说明

- `-url string`: 服务器URL（默认：`http://localhost:8080`）
  - HTTP协议：`http://localhost:8080`
  - gRPC协议：`localhost:9090`（主机:端口格式）
  - WebSocket协议：`http://localhost:8080`（自动转换为ws://）
- `-protocol string`: 通信协议（默认：`http`）
  - `http`: HTTP REST API
  - `grpc`: gRPC 协议
  - `ws`: WebSocket 协议
- `-model string`: 模型目录路径（默认：`./models/enzh`）
- `-test string`: 测试类型（默认：`all`）
  - `all`: 运行所有测试
  - `concurrency`: 高并发测试
  - `sustained`: 持续负载测试
  - `memory`: 内存稳定性测试
  - `reload`: 快速引擎重载测试
  - `mixed`: 混合工作负载测试
- `-c int`: 高并发测试的并发工作线程数（默认：`50`）
- `-d duration`: 持续负载测试的持续时间（默认：`30s`）
- `-n int`: 内存稳定性测试的迭代次数（默认：`1000`）
- `-r int`: 快速重载测试的重载次数（默认：`5`）

## 测试类型说明

### 1. High Concurrency Test（高并发测试）
测试服务器在高并发负载下的表现，同时发起大量并发请求。

### 2. Sustained Load Test（持续负载测试）
在指定时间内保持恒定负载，测试服务器的稳定性和持续性能。

### 3. Memory Stability Test（内存稳定性测试）
通过大量迭代检测潜在的内存泄漏问题。

### 4. Rapid Reload Test（快速重载测试）
测试引擎快速加载和卸载的稳定性。

### 5. Mixed Workload Test（混合工作负载测试）
模拟真实场景，混合翻译请求、健康检查和就绪检查。

## 输出示例

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

## 性能指标说明

- **Total Requests**: 总请求数
- **Successful**: 成功的请求数
- **Failed**: 失败的请求数
- **Duration**: 测试总耗时
- **Throughput**: 吞吐量（请求/秒）
- **Failure Rate**: 失败率
- **Concurrent Load**: 并发负载（工作线程数）

## 评估标准

- ✅ **Excellent**: 失败率 = 0%
- ✓ **Acceptable**: 失败率 < 10%
- ⚠️ **Warning**: 失败率 ≥ 10%

## 协议说明

### HTTP 协议
标准的 REST API，适合大多数场景。

### gRPC 协议
基于 HTTP/2 的高性能 RPC 框架，适合低延迟场景和高并发压力测试。

### WebSocket 协议
全双工通信，适合需要持久连接的场景。注意：WebSocket 连接在测试过程中保持打开状态。

## 注意事项

1. 确保服务器已启动并可访问
2. 确保模型文件存在于指定路径
3. 压力测试可能对系统资源要求很高，建议在测试环境运行
4. 持续负载测试和混合工作负载测试会实时显示进度
5. 建议根据服务器配置调整并发数和测试时长
6. 运行所有测试（`-test all`）可能需要较长时间
7. 使用 gRPC 时，URL 参数应为 `host:port` 格式
8. 使用 WebSocket 时，工具会自动将 HTTP URL 转换为 WS URL
9. WebSocket 协议下的并发测试会共享一个连接，可能影响测试结果

## 示例场景

### 快速检查服务器性能
```bash
./stresstest -test concurrency -c 20
```

### 长时间稳定性测试
```bash
./stresstest -test sustained -d 5m
```

### 内存泄漏检测
```bash
./stresstest -test memory -n 5000
```

### 全面压力测试
```bash
./stresstest -test all
```
