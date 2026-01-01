# API Examples

This document provides quick usage examples for the three protocols of MTranCore.

## Server Configuration

### Starting the Server

The worker service can be configured using command line arguments, environment variables, or default values. The priority order is: **Command Line Arguments > Environment Variables > Default Values**.

#### Using Command Line Arguments

```bash
# Basic usage with custom port
./worker -port 9000

# Custom host and port
./worker -host 127.0.0.1 -port 9000

# Set log level
./worker -log-level debug

# Auto-load model on startup (Option 1: using directory)
./worker -model-dir /path/to/models/enzh

# Auto-load model on startup (Option 2: using individual files)
./worker -model-path models/enzh/model.enzh.intgemm.alphas.bin \
  -lexical-shortlist-path models/enzh/lex.50.50.enzh.s2t.bin \
  -vocabulary-path models/enzh/vocab.enzh.spm

# Enable gRPC Unix socket
./worker -grpc-unix-socket /tmp/mtrancore.sock

# Disable specific protocols
./worker -enable-http false -enable-websocket false

# Enable only gRPC
./worker -enable-http false -enable-websocket false -enable-grpc true
```

#### Using Environment Variables

```bash
# Set environment variables
export LOG_LEVEL=debug
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8988
export MODEL_DIR=./models/enzh
export GRPC_UNIX_SOCKET=/tmp/mtrancore.sock
export ENABLE_HTTP=true
export ENABLE_WEBSOCKET=true
export ENABLE_GRPC=true

# Start the server
./worker
```

#### Command Line Arguments Override Environment Variables

```bash
# Even if SERVER_PORT=8988 is set in environment, this will use port 9000
export SERVER_PORT=8988
./worker -port 9000
```

### Available Configuration Options

| CLI Flag | Environment Variable | Default | Description |
|----------|---------------------|---------|-------------|
| `-log-level` | `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `-model-dir` | `MODEL_DIR` | `./` | Model directory for auto-loading |
| `-model-path` | `MODEL_PATH` | (empty) | Model file path |
| `-lexical-shortlist-path` | `LEXICAL_SHORTLIST_PATH` | (empty) | Lexical shortlist file path |
| `-vocabulary-path` | (array flag) | (empty) | Vocabulary file path(s) |
| `-host` | `SERVER_HOST` | `0.0.0.0` | Server host address |
| `-port` | `SERVER_PORT` | `8988` | Server port |
| `-grpc-unix-socket` | `GRPC_UNIX_SOCKET` | (empty) | Path to Unix socket file for gRPC |
| `-enable-http` | `ENABLE_HTTP` | `true` | Enable HTTP server |
| `-enable-websocket` | `ENABLE_WEBSOCKET` | `true` | Enable WebSocket server |
| `-enable-grpc` | `ENABLE_GRPC` | `true` | Enable gRPC server |
| `-max-length-break` | `MAX_LENGTH_BREAK` | `128` | Maximum sentence length before breaking |

### Help Information

To see all available command line options:

```bash
./worker -h
```

## HTTP/REST API Examples

**Note**: Engine is auto-loaded on startup via `-model-dir` or individual file path parameters. No need to call poweron endpoint.

### 1. Health Check (includes ready status)

```bash
curl http://localhost:8988/health
```

Response:
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "ready": true
  }
}
```

### 2. Translate Text

```bash
curl -X POST http://localhost:8988/trans \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello, world!", "html": false}'
```

Success response (text/plain):
```
你好,世界!
```

Error response (application/json):
```json
{
  "code": 503,
  "message": "Translation engine not ready"
}
```

### 3. Shutdown Server

```bash
# Graceful Shutdown (wait for requests to complete)
curl -X POST http://localhost:8988/exit \
  -H "Content-Type: application/json" \
  -d '{"time": 0, "force": false}'

# Force Shutdown
curl -X POST http://localhost:8988/exit \
  -H "Content-Type: application/json" \
  -d '{"time": 0, "force": true}'
```

## gRPC API Examples

Use `grpcurl` tool to test gRPC API.

**Note**: Engine is auto-loaded on startup. No need to call Poweron.

### 1. Health Check (includes ready status)

```bash
grpcurl -plaintext localhost:8988 translator.TranslatorService/Health
```

Response:
```json
{
  "code": 200,
  "message": "OK",
  "ready": true
}
```

### 2. Translate Text

```bash
grpcurl -plaintext -d '{"text": "Hello, world!", "html": false}' \
  localhost:8988 translator.TranslatorService/Trans
```

### 3. Streaming Translation

```bash
grpcurl -plaintext -d @ localhost:8988 translator.TranslatorService/TransStream <<EOF
{"text": "Hello"}
{"text": "World"}
{"text": "Goodbye"}
EOF
```

### 4. Shutdown Server

```bash
grpcurl -plaintext -d '{"time": 0, "force": false}' \
  localhost:8988 translator.TranslatorService/Exit
```

### 5. Using Unix Domain Socket (for better local performance)

Enable gRPC Unix socket (server side):

```bash
# Set environment variable to enable gRPC Unix socket
export GRPC_UNIX_SOCKET=/tmp/mtrancore.sock
./worker
```

Enable gRPC Unix socket (client side):

```bash
# Use grpcurl to connect to Unix socket
grpcurl -plaintext -unix /tmp/mtrancore.sock translator.TranslatorService/Health

# Translate text
grpcurl -plaintext -unix /tmp/mtrancore.sock \
  -d '{"text": "Hello, world!", "html": false}' \
  translator.TranslatorService/Trans
```

Use benchmark tool to test Unix socket performance:

```bash
# Test Unix socket performance
./benchmark -protocol grpc-unix -grpc-unix /tmp/mtrancore.sock -n 1000 -c 10

# Compare all protocols (including Unix socket)
./benchmark -protocol all -grpc-unix /tmp/mtrancore.sock -n 1000 -c 10
```

**Performance tip**: For local communication, Unix domain socket is usually 20-40% faster than TCP, because it avoids the overhead of the TCP/IP protocol stack.

## WebSocket API Examples

**Note**: Engine is auto-loaded on startup. No need to send poweron message.

### Use JavaScript (Browser/Node.js)

```javascript
const ws = new WebSocket('ws://localhost:8988/ws');

ws.onopen = () => {
  console.log('Connected');

  // 1. Check Health
  ws.send(JSON.stringify({
    type: 'health',
    data: {}
  }));
};

ws.onmessage = (event) => {
  const response = JSON.parse(event.data);
  console.log('Response:', response);

  if (response.type === 'health' && response.data.ready) {
    // 2. Execute Translation
    ws.send(JSON.stringify({
      type: 'trans',
      data: {
        text: 'Hello, world!',
        html: false
      }
    }));
  }

  if (response.type === 'trans' && response.code === 200) {
    console.log('Translation:', response.data.translated_text);

    // 3. Close Connection
    ws.send(JSON.stringify({
      type: 'exit',
      data: { time: 0, force: true }
    }));
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected');
};
```

### Use Python (websockets library)

```python
import asyncio
import json
import websockets

async def translate():
    uri = "ws://localhost:8988/ws"
    async with websockets.connect(uri) as websocket:
        # 1. Check Health
        await websocket.send(json.dumps({
            "type": "health",
            "data": {}
        }))
        response = json.loads(await websocket.recv())
        print("Health response:", response)

        # 2. Execute Translation
        if response.get("data", {}).get("ready"):
            await websocket.send(json.dumps({
                "type": "trans",
                "data": {
                    "text": "Hello, world!",
                    "html": False
                }
            }))
            response = json.loads(await websocket.recv())
            print("Translation:", response.get("data", {}).get("translated_text"))

        # 3. Close Connection
        await websocket.send(json.dumps({
            "type": "exit",
            "data": {"time": 0, "force": True}
        }))

asyncio.run(translate())
```

### Use wscat (command line tool)

```bash
# Install wscat: npm install -g wscat
wscat -c ws://localhost:8988/ws

# Send messages after connection:

# 1. Check Health
> {"type":"health","data":{}}

# 2. Translate Text
> {"type":"trans","data":{"text":"Hello, world!","html":false}}

# 3. Shutdown Server
> {"type":"exit","data":{"time":0,"force":true}}
```

## Error Handling

All protocols use simplified HTTP-standard error codes:

| Code | Description |
|------|-------------|
| 200  | Success |
| 400  | Invalid parameters |
| 500  | Translation failure |
| 502  | Internal error |
| 503  | Engine not ready |

### Error Response Examples

#### HTTP
```json
{
  "code": 400,
  "message": "text is required"
}
```

#### gRPC
```json
{
  "code": 503,
  "message": "Translation engine not ready"
}
```

#### WebSocket
```json
{
  "type": "trans",
  "code": 500,
  "msg": "Translation failed: WASM module error"
}
```

## Complete Workflow Examples

### HTTP API Complete Workflow

```bash
#!/bin/bash

# Engine is auto-loaded on startup, no need to call poweron

# 1. Check Server Health and Ready Status
health=$(curl -s http://localhost:8988/health)
echo "Health: $health"

ready=$(echo $health | jq -r '.data.ready')
if [ "$ready" = "true" ]; then
  # 2. Execute Translation (returns plain text directly)
  result=$(curl -s -X POST http://localhost:8988/trans \
    -H "Content-Type: application/json" \
    -d '{"text": "Hello, world!", "html": false}')

  echo "Translation result: $result"

  # 3. Graceful Shutdown Server
  curl -X POST http://localhost:8988/exit \
    -H "Content-Type: application/json" \
    -d '{"time": 5, "force": false}'
else
  echo "Engine not ready"
fi
```

## More Information

- Complete API Specification: [openapi.json](openapi.json)
- Protocol Detailed Documentation: [REFERENCE.md](REFERENCE.md)
- Project Homepage: [README.md](README.md)
