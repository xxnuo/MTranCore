# MTranCore

## 高级环境变量

### 基础配置
- `MTRAN_OFFLINE` 是否使用离线模式。在离线模式下，不会发送网络请求。默认值为 false。
- `MTRAN_WORKERS` 每个语言模型的工作线程数。默认值 1 对大多数场景已足够。仅在用作高并发服务器时需要调整。
- `MTRAN_LOG_LEVEL` 日志级别，可选项：Error、Warn、Info、Debug
- `MTRAN_DATA_DIR` 数据存储目录，默认为 ~/.cache/mtran。可以清空，文件会在下次翻译时重新下载。

### 内存管理配置
- `MTRAN_MEMORY_STRATEGY` 内存管理策略，可选项：
  - `conservative`（保守）：更快的性能，但占用更多内存
  - `balanced`（平衡，默认）：在内存使用和性能之间取得平衡
  - `aggressive`（积极）：内存占用更少，但可能影响性能
- `MTRAN_MODEL_IDLE_TIMEOUT` 模型空闲超时时间（分钟）。模型在此时间内未使用将被自动释放。设置为 0 禁用自动释放。默认值为 30 分钟。

### 其他配置
- `MTRAN_WORKER_INIT_TIMEOUT` Worker 初始化超时时间，单位为毫秒。默认值为 600000 毫秒（10分钟）。
- `MTRAN_MAX_DETECTION_LENGTH` 最大检测长度，单位为字符。默认值为 64。

## 内存管理策略说明

| 策略 | 模型释放时间 | 内存检查频率 | WASM清理频率 | 适用场景 |
|------|-------------|-------------|-------------|----------|
| conservative | 60分钟 | 5分钟 | 10000次翻译 | 高频使用，内存充足 |
| balanced | 30分钟 | 1分钟 | 5000次翻译 | 一般使用场景（默认） |
| aggressive | 10分钟 | 30秒 | 1000次翻译 | 内存受限环境，偶尔使用 |

## 使用示例

```bash
# 使用保守策略，适合内存受限环境
export MTRAN_MEMORY_STRATEGY=conservative

# 使用积极策略，5分钟后释放模型
export MTRAN_MEMORY_STRATEGY=aggressive
export MTRAN_MODEL_IDLE_TIMEOUT=5

# 禁用自动释放，适合服务器应用
export MTRAN_MODEL_IDLE_TIMEOUT=0
```

## 默认模型存储路径

- 模型文件存储在 `~/.cache/mtran/models` 目录中
