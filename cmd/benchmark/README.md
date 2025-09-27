# Benchmark Tool - 基准测试工具

用于测量 MTranCore 翻译服务器的性能。

## 编译

```bash
cd /home/xxnuo/projects/MTranCore
go build -o benchmark ./cmd/benchmark
```

## 使用方法

### 基本用法

```bash
# 运行所有基准测试（默认，使用HTTP协议）
./benchmark -url http://localhost:8080 -model ./models/enzh -n 100

# 使用不同的协议
./benchmark -protocol http -url http://localhost:8080 -n 100    # HTTP协议
./benchmark -protocol grpc -url localhost:9090 -n 100            # gRPC协议
./benchmark -protocol ws -url http://localhost:8080 -n 100       # WebSocket协议

# 运行特定测试
./benchmark -test compute -n 100      # 简单文本翻译
./benchmark -test html -n 100         # HTML翻译
./benchmark -test long -n 100         # 长文本翻译
./benchmark -test parallel -c 10 -n 100  # 并发翻译测试
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
- `-n int`: 迭代次数（默认：`100`）
- `-c int`: 并发工作线程数（默认：`1`）
- `-test string`: 测试类型（默认：`all`）
  - `all`: 运行所有测试
  - `compute`: 简单文本翻译测试
  - `html`: HTML翻译测试
  - `long`: 长文本翻译测试
  - `parallel`: 并发翻译测试

## 输出示例

```
=== Benchmark Configuration ===
Server URL: http://localhost:8080
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

## 性能指标说明

- **Total Requests**: 总请求数
- **Successful**: 成功的请求数
- **Failed**: 失败的请求数
- **Duration**: 测试总耗时
- **Avg Latency**: 平均延迟
- **Min Latency**: 最小延迟
- **Max Latency**: 最大延迟
- **Throughput**: 吞吐量（请求/秒）

## 协议说明

### HTTP 协议
标准的 REST API，适合大多数场景。

### gRPC 协议
基于 HTTP/2 的高性能 RPC 框架，适合低延迟场景。

### WebSocket 协议
全双工通信，适合需要持久连接的场景。

## 注意事项

1. 确保服务器已启动并可访问
2. 确保模型文件存在于指定路径
3. 首次运行会自动加载引擎，需要一些时间
4. 并发测试（`-c > 1`）可能对系统资源要求较高
5. 使用 gRPC 时，URL 参数应为 `host:port` 格式
6. 使用 WebSocket 时，工具会自动将 HTTP URL 转换为 WS URL
