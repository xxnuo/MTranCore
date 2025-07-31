# MTranCore

## 高级环境变量

### 基础配置
- `MTRAN_OFFLINE` 是否使用离线模式。在离线模式下，不会发送网络请求。默认值为 false。
- `MTRAN_WORKERS` 每个语言模型的工作线程数。默认值 1 对大多数场景已足够。仅在用作高并发服务器时需要调整。
- `MTRAN_LOG_LEVEL` 日志级别，可选项：Error、Warn、Info、Debug
- `MTRAN_DATA_DIR` 数据存储目录，默认为 ~/.cache/mtran。可以清空，文件会在下次翻译时重新下载。

### 简化的内存管理配置
- `MTRAN_MODEL_IDLE_TIMEOUT` 模型空闲超时时间（分钟）。模型在此时间内未使用将被自动释放。设置为 0 禁用自动释放。默认值为 30 分钟。
- `MTRAN_WORKER_RESTART_TASK_COUNT` Worker重启任务数量阈值。Worker处理此数量的任务后会自动重启以释放内存。默认值为 1000。
- `MTRAN_MEMORY_CHECK_INTERVAL` 内存检查间隔（毫秒）。系统检查和释放未使用模型的频率。默认值为 60000 毫秒（1分钟）。

### 其他配置
- `MTRAN_WORKER_INIT_TIMEOUT` Worker 初始化超时时间，单位为毫秒。默认值为 600000 毫秒（10分钟）。
- `MTRAN_MAX_DETECTION_LENGTH` 最大检测长度，单位为字符。默认值为 64。

## 内存管理策略说明

新的简化内存管理策略：

| 特性 | 说明 | 默认值 |
|------|------|--------|
| 模型释放 | 按未使用时间自动释放模型 | 30分钟 |
| Worker重启 | 按任务量重启Worker，保证队列不中断 | 1000次任务 |
| 内存检查 | 定期检查和释放未使用的模型 | 每分钟 |

## 使用示例

```bash
# 设置模型10分钟后自动释放
export MTRAN_MODEL_IDLE_TIMEOUT=10

# 设置Worker每500次任务后重启
export MTRAN_WORKER_RESTART_TASK_COUNT=500

# 设置每30秒检查一次内存
export MTRAN_MEMORY_CHECK_INTERVAL=30000

# 禁用自动释放，适合服务器应用
export MTRAN_MODEL_IDLE_TIMEOUT=0
```

## 默认模型存储路径

- 模型文件存储在 `~/.cache/mtran/models` 目录中
