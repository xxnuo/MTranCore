# Worker Service Definition

The server supports three protocols: gRPC, HTTP, and WebSocket, all running on the same port.

## Environment Variables

- `LOG_LEVEL`: Log level, default `info`, options: `debug`, `info`, `warn`, `error` (case-insensitive).
- `WORK_DIR`: Working directory, default `./`, mainly used to prepend to the model directory path passed in poweron.
- `SERVER_HOST`: Server host address, default `0.0.0.0`.
- `SERVER_PORT`: Unified server port, default `8988`, all enabled services (HTTP, WebSocket, gRPC) will run on this port.
- `ENABLE_HTTP`: Whether to enable HTTP service, default `true`, options: `true`, `false`, 0, 1, 'yes', 'no' (case-insensitive).
- `ENABLE_WEBSOCKET`: Whether to enable WebSocket service, default `true`, options: `true`, `false`, 0, 1, 'yes', 'no' (case-insensitive).
- `ENABLE_GRPC`: Whether to enable gRPC service, default `true`, options: `true`, `false`, 0, 1, 'yes', 'no' (case-insensitive).

## Main API Definitions

### health - Health Check

### poweron - Start

Parameters:

- path: Model directory path relative to the program's working directory, must contain lex*.bin, model*.bin, vocab*.spm (or srcvocab*.spm, trgvocab\*.spm) files.

Error code definitions for return values:

- 0: Success
- 1000: Invalid parameters
- 1001: Model directory does not exist
- 1002: Model files incomplete
- 1003: Internal error
- 1009: Unknown error

### exit - Shutdown

Parameters:

- time: How many seconds before shutting down the server
- force: Whether to force shutdown (default false, will wait for all requests to complete)

Error code definitions for return values:

- 0: Success
- 1100: Invalid parameters
- 1101: Waiting for requests to complete
- 1109: Internal error

### health - Get Status

Returns whether the service is health

### trans - Translate Text

Main parameters:

- text: Text to translate
- html: Whether the text is HTML (default false)

Error code definitions for return values:

- 0: Success
- 1200: Invalid parameters
- 1201: Translation failed
- 1209: Internal error
