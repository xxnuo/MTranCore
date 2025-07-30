# MTranCore

## 高级环境变量

- `MTRAN_OFFLINE` 是否使用离线模式。在离线模式下，不会发送网络请求。默认值为 false。
- `MTRAN_WORKERS` 每个语言模型的工作线程数。默认值 1 对大多数场景已足够。仅在用作高并发服务器时需要调整。
- `MTRAN_LOG_LEVEL` 日志级别，可选项：Error、Warn、Info、Debug
- `MTRAN_DATA_DIR` 数据存储目录，默认为 ~/.cache/mtran。可以清空，文件会在下次翻译时重新下载。
- `MTRAN_AUTO_RELEASE` 是否启用模型自动释放功能。默认值为 true。
- `MTRAN_RELEASE_INTERVAL` 模型自动释放时间间隔，单位为分钟。默认值为 30 分钟。
- `MTRAN_CLEANUP_INTERVAL` 内存清理间隔，单位为次。默认值为 5000 次。
- `MTRAN_CLEANUP_TIME_THRESHOLD` 内存清理时间阈值，单位为分钟。默认值为 30 分钟。
- `MTRAN_MEMORY_CHECK_INTERVAL` 内存检查间隔，单位为毫秒。默认值为 60000 毫秒（1分钟）。
- `MTRAN_TIMEOUT_RESET_THRESHOLD` 超时重置阈值，单位为毫秒。默认值为 300000 毫秒（5分钟）。
- `MTRAN_WORKER_INIT_TIMEOUT` Worker 初始化超时时间，单位为毫秒。默认值为 600000 毫秒（600秒）。
- `MTRAN_MAX_DETECTION_LENGTH` 最大检测长度，单位为字符。默认值为 64。

## 默认模型存储路径

- 模型文件存储在 `~/.cache/mtran/models` 目录中
