# Worker Protocol Reference

This document provides a comprehensive reference for all communication protocols supported by the MTranCore Worker service.

## Table of Contents

- [Overview](#overview)
- [Error Codes](#error-codes)
- [gRPC Protocol](#grpc-protocol)
- [HTTP Protocol](#http-protocol)
- [WebSocket Protocol](#websocket-protocol)
- [Configuration](#configuration)

## Overview

MTranCore Worker 是一个翻译服务,通过三种协议提供翻译能力:

- **gRPC**: 高性能 RPC 协议,支持流式传输
- **HTTP**: RESTful API,使用 JSON
- **WebSocket**: 实时双向通信

**重要**:
- Worker 在启动时会自动加载模型(需要通过环境变量或命令行参数指定模型路径)
- 所有协议共享相同的翻译引擎实例
- 翻译失败或引擎未就绪时,Worker 会自动退出
- 简化的 API 仅保留核心功能:翻译(Trans/TransStream)、状态检查(Health)、关闭(Exit)

## Error Codes

The service uses standardized error codes across all protocols:

| Code | Constant                     | Description                                |
| ---- | ---------------------------- | ------------------------------------------ |
| 200  | `CodeSuccess`                | Operation completed successfully           |
| 1000 | `CodeLoadInvalidParams`   | Invalid parameters for load request     |
| 1001 | `CodeLoadPathNotExists`   | Model path does not exist                  |
| 1002 | `CodeLoadIncompleteFiles` | Model files are incomplete or missing      |
| 1003 | `CodeLoadInternalError`   | Internal error during load              |
| 1009 | `CodeLoadUnknownError`    | Unknown load error                      |
| 1100 | `CodeExitInvalidParams`  | Invalid parameters for exit request    |
| 1101 | `CodeExitWaitingTask`    | Server is shutting down, waiting for tasks |
| 1109 | `CodeExitInternalError`  | Internal error during exit             |
| 1200 | `CodeTransInvalidParams`   | Invalid parameters for trans request     |
| 1201 | `CodeTransFailure`         | Translation computation failed             |
| 1209 | `CodeTransInternalError`   | Internal error during trans              |

## gRPC Protocol

### Connection

Connect to the gRPC server using the configured address (default: `0.0.0.0:8988`).

```
grpc://<host>:<port>
```

### Service: TranslatorService

The gRPC service is defined in `worker.proto`:

```proto
service TranslatorService {
  rpc Health(HealthRequest) returns (HealthResponse);
  rpc Trans(TransRequest) returns (TransResponse);
  rpc TransStream(stream TransRequest) returns (stream TransResponse);
  rpc Exit(ExitRequest) returns (ExitResponse);
}
```

### Methods

#### Health

检查翻译引擎是否就绪。

**Request**: `HealthRequest` (empty)

**Response**: `HealthResponse`

```proto
message HealthResponse {
  int32 code = 1;      // Always 200
  string message = 2;  // "OK"
  bool health = 3;      // true if engine is loaded
}
```

#### Trans

翻译单个文本。翻译失败时 Worker 会自动退出。

**Request**: `TransRequest`

```proto
message TransRequest {
  string text = 1;  // Text to translate (required)
  bool html = 2;    // Whether to treat input as HTML (default: false)
}
```

**Response**: `TransResponse`

```proto
message TransResponse {
  int32 code = 1;             // 200 on success
  string message = 2;         // "OK"
  string translated_text = 3; // Translated text
}
```

**Error Codes**:

- `1200`: Text is required
- Worker 会在翻译失败或引擎未就绪时自动退出

#### TransStream

使用双向流翻译多个文本。翻译失败时 Worker 会自动退出。

**Request Stream**: `TransRequest`

```proto
message TransRequest {
  string text = 1;  // Text to translate
  bool html = 2;    // Whether to treat input as HTML
}
```

**Response Stream**: `TransResponse`

```proto
message TransResponse {
  int32 code = 1;             // 200 on success
  string message = 2;         // "OK"
  string translated_text = 3; // Translated text
}
```

**Behavior**:

- Client sends multiple `TransRequest` messages
- Server responds with corresponding `TransResponse` for each request
- Client closes stream by sending EOF
- Empty text returns empty translation with success code
- Worker 会在翻译失败或引擎未就绪时自动退出

#### Exit

Shut down the server gracefully or forcefully.

**Request**: `ExitRequest`

```proto
message ExitRequest {
  int32 time = 1;   // Seconds to wait before shutdown (default: 0)
  bool force = 2;   // Force shutdown without waiting for active streams
}
```

**Response**: `ExitResponse`

```proto
message ExitResponse {
  int32 code = 1;      // 200 on success, error code otherwise
  string message = 2;  // Description of shutdown status
}
```

**Error Codes**:

- `1100`: Time must be non-negative
- `1101`: Waiting for active streams to complete (non-force shutdown)

**Behavior**:

- If `force=true`: Immediate shutdown
- If `force=false`: Waits up to 30 seconds for active streams, then forces shutdown

## HTTP Protocol

### Connection

Connect to the HTTP server using the configured address (default: `http://0.0.0.0:8988`).

All endpoints use JSON for request and response bodies.

### Standard Response Format

Success responses:

```json
{
  "code": 200,
  "data": {
    /* endpoint-specific data */
  }
}
```

Error responses:

```json
{
  "code": <error_code>,
  "message": "Error description"
}
```

### Endpoints

#### GET /health

检查翻译引擎是否就绪。

**Success Response**: HTTP 200

```json
{
  "code": 200,
  "data": {
    "health": true // or false
  }
}
```

#### POST /trans

翻译文本。翻译失败时 Worker 会自动退出。

**Request Body**:

```json
{
  "text": "Text to translate", // Required
  "html": false // Optional, default: false
}
```

**Success Response**: HTTP 200

Content-Type: `text/plain; charset=utf-8`

Returns the translated text directly as plain text (not JSON).

**Error Responses**:

- HTTP 400 (Invalid parameters):
  ```json
  {
    "code": 1200,
    "message": "text is required"
  }
  ```
- Worker 会在翻译失败或引擎未就绪时自动退出

#### POST /exit

Shut down the server gracefully or forcefully.

**Request Body** (optional):

```json
{
  "time": 0, // Seconds to wait (default: 0)
  "force": false // Force shutdown (default: false)
}
```

**Success Response**: HTTP 200

```json
{
  "code": 200,
  "data": {
    "message": "Server is shutting down"
  }
}
```

**Error Response**: HTTP 200

```json
{
  "code": 1101,
  "message": "Server is shutting down, waiting for requests to complete"
}
```

## WebSocket Protocol

### Connection

Connect to the WebSocket server at the configured address (default: `ws://0.0.0.0:8988/ws`).

### Message Format

All messages are JSON objects with the following structure:

**Client Message**:

```json
{
  "type": "message_type",
  "data": {
    /* type-specific data */
  }
}
```

**Server Response**:

```json
{
  "type": "message_type", // Same as request type
  "code": 200, // Error code
  "msg": "success", // Message or error description
  "data": {
    /* type-specific data */
  }
}
```

### Message Types

#### trans

翻译文本。翻译失败时 Worker 会自动退出。

**Client Message**:

```json
{
  "type": "trans",
  "data": {
    "text": "Text to translate",
    "html": false
  }
}
```

**Server Response**:

```json
{
  "type": "trans",
  "code": 200,
  "msg": "success",
  "data": {
    "translated_text": "Translated text"
  }
}
```

**Error Codes**:
- `1200`: text is required
- Worker 会在翻译失败或引擎未就绪时自动退出

#### exit

Shut down the server.

**Client Message**:

```json
{
  "type": "exit",
  "data": {
    "time": 0, // Optional, default: 0
    "force": false // Optional, default: false
  }
}
```

**Server Response**:

```json
{
  "type": "exit",
  "code": 200,
  "msg": "success",
  "data": {
    "message": "Server is shutting down"
  }
}
```

**Note**: The WebSocket connection is closed after exit response.

### Unknown Message Type

If an unknown message type is sent:

**Server Response**:

```json
{
  "type": "unknown_type",
  "code": 1009,
  "msg": "Unknown message type: unknown_type"
}
```

## Configuration

The worker service can be configured using environment variables or command-line flags:

### Required: Model Configuration

You **must** configure the model path for auto-loading on startup using one of these methods:

**Environment Variables**:
```bash
# Option 1: Model directory
export MODEL_PATH=/path/to/model

# Option 2: Individual files
export MODEL_FILE=/path/to/model.bin
export SHORTLIST_FILE=/path/to/lex.bin
export VOCAB_FILE=/path/to/vocab.spm  # Single vocabulary
# OR
export VOCAB_FILES=/path/to/src.spm,/path/to/trg.spm  # Dual vocabularies

# Note: Supports ~ expansion for home directory
# Both work: MODEL_PATH=~/.config/mtran/models/en_zh-Hans
#            MODEL_PATH=$HOME/.config/mtran/models/en_zh-Hans
```

**Command-line Flags**:
```bash
# Option 1: Model directory
./worker --model-path=/path/to/model

# Option 2: Individual files
./worker \
  --model-file=/path/to/model.bin \
  --shortlist-file=/path/to/lex.bin \
  --vocab-file=/path/to/vocab.spm

# OR with dual vocabularies
./worker \
  --model-file=/path/to/model.bin \
  --shortlist-file=/path/to/lex.bin \
  --vocab-files=/path/to/src.spm,/path/to/trg.spm
```

### General Settings

| Variable      | Default   | Description                                          |
| ------------- | --------- | ---------------------------------------------------- |
| `LOG_LEVEL`   | `info`    | Logging level (debug, info, warn, error)             |
| `WORK_DIR`    | `./`      | Working directory for relative model paths           |
| `SERVER_HOST` | `0.0.0.0` | Server host address (shared by all enabled services) |
| `SERVER_PORT` | `8988`    | Server port (shared by all enabled services)         |

### Service Control

All services (HTTP, WebSocket, gRPC) run on the same port using connection multiplexing.

| Variable           | Default | Description          |
| ------------------ | ------- | -------------------- |
| `ENABLE_HTTP`      | `true`  | Enable HTTP REST API |
| `ENABLE_WEBSOCKET` | `true`  | Enable WebSocket API |
| `ENABLE_GRPC`      | `true`  | Enable gRPC API      |

### gRPC Unix Domain Socket (Advanced)

For better local performance, gRPC can also listen on a Unix domain socket:

| Variable           | Default | Description                                         |
| ------------------ | ------- | --------------------------------------------------- |
| `GRPC_UNIX_SOCKET` | (empty) | Path to Unix socket file (e.g., `/tmp/mtrancore.sock`) |

When enabled, gRPC will listen on both TCP and Unix socket simultaneously. Unix domain sockets provide 20-40% better performance for local IPC by avoiding TCP/IP stack overhead.

### Example Configuration

Disable WebSocket and change server port:

```bash
export ENABLE_WEBSOCKET=false
export SERVER_PORT=9000
./worker
```

Run only gRPC server:

```bash
export ENABLE_HTTP=false
export ENABLE_WEBSOCKET=false
export ENABLE_GRPC=true
./worker
```

Enable gRPC with Unix socket for better local performance:

```bash
export GRPC_UNIX_SOCKET=/tmp/mtrancore.sock
./worker
```

This will start both TCP (on port 8988) and Unix socket listeners for gRPC.

## Common Workflows

### Basic Translation Flow

**Note**: Worker auto-loads the model on startup, so no manual poweron is needed.

1. **Check Health** (optional)
   - gRPC: Call `Health()`
   - HTTP: `GET /health`

2. **Translate**
   - gRPC: Call `Trans(text="Hello", html=false)`
   - HTTP: `POST /trans` with `{"text": "Hello", "html": false}`
   - WebSocket: Send `{"type": "trans", "data": {"text": "Hello", "html": false}}`

### Batch Translation (gRPC Stream)

1. Open `TransStream` connection
2. Send multiple `TransRequest` messages
3. Receive corresponding `TransResponse` for each request
4. Close stream when done

### Graceful Shutdown

1. **Wait for completion**:
   - Call `Exit(time=5, force=false)` to wait 5 seconds + up to 30 seconds for active tasks

2. **Force shutdown**:
   - Call `Exit(time=0, force=true)` for immediate shutdown

## Notes

- Worker auto-loads the model on startup (requires model path configuration)
- Translation failures cause Worker to exit automatically
- All three protocols share the same translation engine instance
- Paths can be absolute or relative to the configured `WORK_DIR`
- The service handles concurrent requests through an internal queue
- gRPC `TransStream` is optimized for batch processing
- Empty text in trans requests returns empty translation with success code
- All servers can be enabled/disabled independently via environment variables
