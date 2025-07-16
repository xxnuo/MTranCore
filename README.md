# MTranCore

## Advanced Environment Variables

- `MTRAN_OFFLINE` Whether to use offline mode. In offline mode, no network requests will be made. Default value is false.
- `MTRAN_WORKERS` Number of worker threads for each language model. The default value of 1 is sufficient for most scenarios. Only needs adjustment when used as a high-concurrency server.
- `MTRAN_LOG_LEVEL` Log level, available options: Error, Warn, Info, Debug
- `MTRAN_DATA_DIR` Data storage directory, default is ~/.cache/mtran. Can be cleared, files will be re-downloaded on next translation.

## Default Model Storage Path

- Model files are stored in the `~/.cache/mtran/models` directory
