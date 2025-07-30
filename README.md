# MTranCore

## Advanced Environment Variables

### Basic Configuration
- `MTRAN_OFFLINE` Whether to use offline mode. In offline mode, no network requests will be made. Default value is false.
- `MTRAN_WORKERS` Number of worker threads for each language model. The default value of 1 is sufficient for most scenarios. Only needs adjustment when used as a high-concurrency server.
- `MTRAN_LOG_LEVEL` Log level, available options: Error, Warn, Info, Debug
- `MTRAN_DATA_DIR` Data storage directory, default is ~/.cache/mtran. Can be cleared, files will be re-downloaded on next translation.

### Memory Management Configuration
- `MTRAN_MEMORY_STRATEGY` Memory management strategy, available options:
  - `conservative`: Faster performance, but uses more memory
  - `balanced` (default): Balance between memory usage and performance
  - `aggressive`: Lower memory usage, but may affect performance
- `MTRAN_MODEL_IDLE_TIMEOUT` Model idle timeout (minutes). Models will be automatically released if not used within this time. Set to 0 to disable auto-release. Default value is 30 minutes.

### Other Configuration
- `MTRAN_WORKER_INIT_TIMEOUT` Worker initialization timeout, in milliseconds. Default value is 600000 milliseconds (10 minutes).
- `MTRAN_MAX_DETECTION_LENGTH` Maximum detection length, in characters. Default value is 64.

## Memory Management Strategy Details

| Strategy | Model Release Time | Memory Check Frequency | WASM Cleanup Frequency | Use Case |
|----------|-------------------|----------------------|----------------------|----------|
| conservative | 60 minutes | 5 minutes | 10000 translations | High-frequency use, sufficient memory |
| balanced | 30 minutes | 1 minute | 5000 translations | General use cases (default) |
| aggressive | 10 minutes | 30 seconds | 1000 translations | Memory-constrained environments, occasional use |

## Usage Examples

```bash
# Use conservative strategy for memory-constrained environments
export MTRAN_MEMORY_STRATEGY=conservative

# Use aggressive strategy with 5-minute model release
export MTRAN_MEMORY_STRATEGY=aggressive
export MTRAN_MODEL_IDLE_TIMEOUT=5

# Disable auto-release for server applications
export MTRAN_MODEL_IDLE_TIMEOUT=0
```

## Default Model Storage Path

- Model files are stored in the `~/.cache/mtran/models` directory