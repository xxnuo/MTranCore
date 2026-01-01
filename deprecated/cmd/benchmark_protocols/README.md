# 协议性能测试

## 简介

[benchmark_protocols](main.go) 测试 WebSocket、gRPC (TCP)、gRPC (Unix Socket) 和 HTTP 协议在高并发环境下的性能表现。

## 编译

```bash
go build -o benchmark_protocols
```

## 使用方法

### 基本用法

```bash
./benchmark_protocols -server localhost:8080
```

### 参数说明

- `-server`: 服务器地址 (默认: localhost:8080)
- `-unix-socket`: gRPC Unix Socket 路径 (可选，仅用于 gRPC Unix Socket 测试)
- `-requests`: 总请求数 (默认: 10000)
- `-concurrent`: 并发数 (默认: 100)
- `-text`: 测试文本 (默认: "Hello, world! This is a test message.")
- `-html`: 是否启用 HTML 模式 (默认: false)
- `-warmup`: 预热请求数 (默认: 100)
- `-verbose`: 显示详细错误信息 (默认: false)

### 示例

**基本测试（HTTP + WebSocket + gRPC TCP）:**
```bash
./benchmark_protocols \
  -server localhost:8080 \
  -requests 50000 \
  -concurrent 200 \
  -text "测试翻译性能" \
  -warmup 500 \
  -verbose
```

**包含 gRPC Unix Socket 测试:**
```bash
./benchmark_protocols \
  -server localhost:8080 \
  -unix-socket /tmp/mtrancore.sock \
  -requests 10000 \
  -concurrent 100
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

**基本启动（TCP 端口）:**
```bash
go run ./cmd/worker -model-dir ./models/enzh -port 8080
```

**启用 Unix Socket（用于 gRPC Unix Socket 测试）:**
```bash
go run ./cmd/worker \
  -model-dir ./models/enzh \
  -port 8080 \
  -grpc-unix-socket /tmp/mtrancore.sock
```

2. **等待翻译引擎初始化完成**

   服务器启动后，翻译引擎需要一些时间来加载模型。你可以通过以下方式检查：

   ```bash
   curl http://localhost:8080/health
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
- 测试协议会依次执行，每个协议测试间隔 1 秒
- 确保翻译引擎已完全初始化再进行测试
- **每个请求使用不同的文本**（添加序号后缀），避免模型缓存影响测试结果
- 如果指定了 `-unix-socket`，会额外测试 gRPC Unix Socket 的性能

## 性能对比

### 有缓存 vs 无缓存

- **有缓存**（相同文本）：平均延迟 ~12ms，QPS ~15000
- **无缓存**（不同文本）：平均延迟 ~3.2s，QPS ~60
- **性能差异**：~267倍

### 协议性能（无缓存场景）

在翻译这种计算密集型任务中，协议开销几乎可以忽略不计：

| 协议 | 平均延迟 | 相对差异 |
|------|---------| ---------|
| WebSocket | 3.21s | 基准 |
| HTTP | 3.25s | +1.3% |
| gRPC (TCP) | 3.32s | +3.4% |

**结论**：协议选择对性能影响很小（<5%），主要瓶颈在翻译计算本身。可以根据其他因素（易用性、双向通信、类型安全等）选择协议。

### Unix Socket vs TCP

理论上 Unix Socket 在本地进程间通信时比 TCP 更快，因为：
- 无需经过网络协议栈
- 减少了数据拷贝
- 降低了系统调用开销

实测中对于翻译这种计算密集型任务，协议开销占比极小，Unix Socket 的优势不明显。
