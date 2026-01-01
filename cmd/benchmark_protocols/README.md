# 协议性能测试

## 简介

[benchmark_protocols](cmd/benchmark_protocols/main.go) 测试 WebSocket、gRPC 和 HTTP 三种协议在高并发环境下的性能表现。

## 编译

```bash
go build -o benchmark_protocols ./cmd/benchmark_protocols
```

## 使用方法

### 基本用法

```bash
./benchmark_protocols -server localhost:8080
```

### 参数说明

- `-server`: 服务器地址 (默认: localhost:8080)
- `-requests`: 总请求数 (默认: 10000)
- `-concurrent`: 并发数 (默认: 100)
- `-text`: 测试文本 (默认: "Hello, world! This is a test message.")
- `-html`: 是否启用 HTML 模式 (默认: false)
- `-warmup`: 预热请求数 (默认: 100)
- `-verbose`: 显示详细错误信息 (默认: false)

### 示例

```bash
./benchmark_protocols \
  -server localhost:8080 \
  -requests 50000 \
  -concurrent 200 \
  -text "测试翻译性能" \
  -warmup 500 \
  -verbose
```

## 输出指标

- **Total Requests**: 总请求数
- **Success**: 成功请求数
- **Failure**: 失败请求数
- **Total Duration**: 总耗时
- **Requests/sec**: 每秒请求数 (QPS)
- **Avg Latency**: 平均延迟
- **Min/Max Latency**: 最小/最大延迟
- **P50/P95/P99 Latency**: 50%/95%/99% 分位延迟

## 测试前准备

1. 启动 MTranCore worker 服务器:

```bash
go run ./cmd/worker -enable-http -enable-ws -enable-grpc
```

2. **等待翻译引擎初始化完成**

   服务器启动后，翻译引擎需要一些时间来加载模型。你可以通过以下方式检查：

   ```bash
   # 检查健康状态
   curl http://localhost:8080/health
   # 应该返回 {"code":0,"message":"OK","data":{"ready":true}}
   ```

   如果 `ready` 为 `false`，说明引擎还在初始化中，请等待几秒钟。

3. 确认引擎就绪后再运行性能测试

## 调试

如果测试失败，使用 `-verbose` 参数查看详细错误：

```bash
./benchmark_protocols -server localhost:8080 -requests 10 -concurrent 2 -verbose
```

常见错误：
- `connection refused`: 服务器未启动
- `Translation engine not ready`: 引擎还在初始化，请等待
- `HTTP 404`: 端点路径错误或服务未启用

## 注意事项

- 测试会自动进行预热，避免冷启动影响结果
- 建议在稳定网络环境下测试
- 高并发测试时需确保系统资源充足
- 三个协议会依次测试，每个协议测试间隔 1 秒
- 确保翻译引擎已完全初始化再进行测试
