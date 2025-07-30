# MTranCore

## Advanced Environment Variables

- `MTRAN_OFFLINE` Whether to use offline mode. In offline mode, no network requests will be made. Default value is false.
- `MTRAN_WORKERS` Number of worker threads for each language model. The default value of 1 is sufficient for most scenarios. Only needs adjustment when used as a high-concurrency server.
- `MTRAN_LOG_LEVEL` Log level, available options: Error, Warn, Info, Debug
- `MTRAN_DATA_DIR` Data storage directory, default is ~/.cache/mtran. Can be cleared, files will be re-downloaded on next translation.
- `MTRAN_AUTO_RELEASE` Whether to enable model auto-release feature. Default value is true.
- `MTRAN_RELEASE_INTERVAL` Model auto-release interval, in minutes. Default value is 30 minutes.
- `MTRAN_CLEANUP_INTERVAL` Memory cleanup interval, in times. Default value is 5000 times.
- `MTRAN_CLEANUP_TIME_THRESHOLD` Memory cleanup time threshold, in minutes. Default value is 30 minutes.
- `MTRAN_MEMORY_CHECK_INTERVAL` Memory check interval, in milliseconds. Default value is 60000 milliseconds (1 minute).
- `MTRAN_TIMEOUT_RESET_THRESHOLD` Timeout reset threshold, in milliseconds. Default value is 300000 milliseconds (5 minutes).
- `MTRAN_WORKER_INIT_TIMEOUT` Worker initialization timeout, in milliseconds. Default value is 600000 milliseconds (600 seconds).
- `MTRAN_MAX_DETECTION_LENGTH` Maximum detection length, in characters. Default value is 64.

## Default Model Storage Path

- Model files are stored in the `~/.cache/mtran/models` directory