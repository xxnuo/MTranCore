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

# Set working directory
./worker -work-dir /path/to/models

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
export WORK_DIR=./models
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
| `-work-dir` | `WORK_DIR` | `./` | Working directory |
| `-host` | `SERVER_HOST` | `0.0.0.0` | Server host address |
| `-port` | `SERVER_PORT` | `8988` | Server port |
| `-grpc-unix-socket` | `GRPC_UNIX_SOCKET` | (empty) | Path to Unix socket file for gRPC |
| `-enable-http` | `ENABLE_HTTP` | `true` | Enable HTTP server |
| `-enable-websocket` | `ENABLE_WEBSOCKET` | `true` | Enable WebSocket server |
| `-enable-grpc` | `ENABLE_GRPC` | `true` | Enable gRPC server |

### Help Information

To see all available command line options:

```bash
./worker -h
```

## HTTP/REST API Examples

### 1. Health Check

```bash
curl http://localhost:8988/health
```

### 2. Load Translation Engine

#### Option 1: Using Model Directory Path

```bash
curl -X POST http://localhost:8988/poweron \
  -H "Content-Type: application/json" \
  -d '{"path": "models/enzh"}'
```

#### Option 2: Using Individual File Paths

```bash
# With single vocabulary file
curl -X POST http://localhost:8988/poweron \
  -H "Content-Type: application/json" \
  -d '{
    "model_path": "models/enzh/model.enzh.intgemm.alphas.bin",
    "lexical_shortlist_path": "models/enzh/lex.50.50.enzh.s2t.bin",
    "vocabulary_path": "models/enzh/single_vocab.enzh.spm"
  }'

# With dual vocabulary files
curl -X POST http://localhost:8988/poweron \
  -H "Content-Type: application/json" \
  -d '{
    "model_path": "models/enzh/model.enzh.intgemm.alphas.bin",
    "lexical_shortlist_path": "models/enzh/lex.50.50.enzh.s2t.bin",
    "vocabulary_paths": [
      "models/enzh/srcvocab.enzh.spm",
      "models/enzh/trgvocab.enzh.spm"
    ]
  }'
```

**Note**: Individual file paths take priority over the `path` parameter. You can also combine `vocabulary_path` and `vocabulary_paths` - they will be merged with `vocabulary_path` first.

### 3. Check Ready Status

```bash
curl http://localhost:8988/ready
```

### 4. Translate Text

```bash
curl -X POST http://localhost:8988/compute \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello, world!", "html": false}'
```

### 5. Shutdown Server

```bash
# Graceful Shutdown (wait for requests to complete)
curl -X POST http://localhost:8988/poweroff \
  -H "Content-Type: application/json" \
  -d '{"time": 0, "force": false}'

# Force Shutdown
curl -X POST http://localhost:8988/poweroff \
  -H "Content-Type: application/json" \
  -d '{"time": 0, "force": true}'
```

## gRPC API Examples

Use `grpcurl` tool to test gRPC API:

### 1. Health Check

```bash
grpcurl -plaintext localhost:8988 translator.TranslatorService/Health
```

### 2. Load Translation Engine

#### Option 1: Using Model Directory Path

```bash
grpcurl -plaintext -d '{"path": "models/enzh"}' \
  localhost:8988 translator.TranslatorService/Poweron
```

#### Option 2: Using Individual File Paths

```bash
# With single vocabulary file
grpcurl -plaintext -d '{
  "model_path": "models/enzh/model.enzh.intgemm.alphas.bin",
  "lexical_shortlist_path": "models/enzh/lex.50.50.enzh.s2t.bin",
  "vocabulary_path": "models/enzh/single_vocab.enzh.spm"
}' localhost:8988 translator.TranslatorService/Poweron

# With dual vocabulary files
grpcurl -plaintext -d '{
  "model_path": "models/enzh/model.enzh.intgemm.alphas.bin",
  "lexical_shortlist_path": "models/enzh/lex.50.50.enzh.s2t.bin",
  "vocabulary_paths": [
    "models/enzh/srcvocab.enzh.spm",
    "models/enzh/trgvocab.enzh.spm"
  ]
}' localhost:8988 translator.TranslatorService/Poweron
```

### 3. Check Ready Status

```bash
grpcurl -plaintext localhost:8988 translator.TranslatorService/Ready
```

### 4. Translate Text

```bash
grpcurl -plaintext -d '{"text": "Hello, world!", "html": false}' \
  localhost:8988 translator.TranslatorService/Compute
```

### 5. Streaming Translation

```bash
grpcurl -plaintext -d @ localhost:8988 translator.TranslatorService/ComputeStream <<EOF
{"text": "Hello"}
{"text": "World"}
{"text": "Goodbye"}
EOF
```

### 6. Shutdown Server

```bash
grpcurl -plaintext -d '{"time": 0, "force": false}' \
  localhost:8988 translator.TranslatorService/Poweroff
```

### 7. Using Unix Domain Socket (for better local performance)

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
  translator.TranslatorService/Compute
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

### Use JavaScript (Browser/Node.js)

```javascript
const ws = new WebSocket('ws://localhost:8988/ws');

ws.onopen = () => {
  console.log('Connected');
  
  // 1. Load Engine (Option 1: using directory path)
  ws.send(JSON.stringify({
    type: 'poweron',
    data: { path: 'models/enzh' }
  }));
  
  // Alternative: Load Engine (Option 2: using individual file paths)
  // ws.send(JSON.stringify({
  //   type: 'poweron',
  //   data: {
  //     model_path: 'models/enzh/model.enzh.intgemm.alphas.bin',
  //     lexical_shortlist_path: 'models/enzh/lex.50.50.enzh.s2t.bin',
  //     vocabulary_path: 'models/enzh/single_vocab.enzh.spm'
  //   }
  // }));
};

ws.onmessage = (event) => {
  const response = JSON.parse(event.data);
  console.log('Response:', response);
  
  if (response.type === 'poweron' && response.code === 0) {
    // 2. Check Ready Status
    ws.send(JSON.stringify({
      type: 'ready',
      data: {}
    }));
  }
  
  if (response.type === 'ready' && response.data.ready) {
    // 3. Execute Translation
    ws.send(JSON.stringify({
      type: 'compute',
      data: {
        text: 'Hello, world!',
        html: false
      }
    }));
  }
  
  if (response.type === 'compute' && response.code === 0) {
    console.log('Translation:', response.data.translated_text);
    
    // 4. Close Connection
    ws.send(JSON.stringify({
      type: 'poweroff',
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
        # 1. Load Engine
        await websocket.send(json.dumps({
            "type": "poweron",
            "data": {"path": "models/enzh"}
        }))
        response = json.loads(await websocket.recv())
        print("Poweron response:", response)
        
        # 2. Check Ready Status
        await websocket.send(json.dumps({
            "type": "ready",
            "data": {}
        }))
        response = json.loads(await websocket.recv())
        print("Ready response:", response)
        
        # 3. Execute Translation
        if response.get("data", {}).get("ready"):
            await websocket.send(json.dumps({
                "type": "compute",
                "data": {
                    "text": "Hello, world!",
                    "html": False
                }
            }))
            response = json.loads(await websocket.recv())
            print("Translation:", response.get("data", {}).get("translated_text"))
        
        # 4. Close Connection
        await websocket.send(json.dumps({
            "type": "poweroff",
            "data": {"time": 0, "force": True}
        }))

asyncio.run(translate())
```

### Use wscat (command line tool)

```bash
# Install wscat: npm install -g wscat
wscat -c ws://localhost:8988/ws

# Send messages after connection:

# 1. Load Engine
> {"type":"poweron","data":{"path":"models/enzh"}}

# 2. Check Ready Status
> {"type":"ready","data":{}}

# 3. Translate Text
> {"type":"compute","data":{"text":"Hello, world!","html":false}}

# 4. Shutdown Server
> {"type":"poweroff","data":{"time":0,"force":true}}
```

## Error Handling

All protocols use the same error code system. The `code` field in the response indicates the operation result:

- `0`: Success
- `1000-1099`: Poweron related errors
- `1100-1199`: Poweroff related errors  
- `1200-1299`: Compute related errors

Detailed error code list please refer to [REFERENCE.md](REFERENCE.md#error-codes).

### Error Response Examples

#### HTTP
```json
{
  "code": 1001,
  "message": "path does not exist: /invalid/path"
}
```

#### gRPC
```json
{
  "code": 1200,
  "message": "Engine is not ready. Please call poweron first"
}
```

#### WebSocket
```json
{
  "type": "compute",
  "code": 1201,
  "msg": "Translation failed: context deadline exceeded"
}
```

## Complete Workflow Examples

### HTTP API Complete Workflow

```bash
#!/bin/bash

# 1. Check Server Health Status
curl http://localhost:8988/health

# 2. Load Translation Model
curl -X POST http://localhost:8988/poweron \
  -H "Content-Type: application/json" \
  -d '{"path": "models/enzh"}'

# 3. Wait for Model Loading
while true; do
  ready=$(curl -s http://localhost:8988/ready | jq -r '.data.ready')
  if [ "$ready" = "true" ]; then
    echo "Engine is ready"
    break
  fi
  echo "Waiting for engine..."
  sleep 1
done

# 4. Execute Translation
result=$(curl -s -X POST http://localhost:8988/compute \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello, world!", "html": false}')

echo "Translation result: $result"

# 5. Extract Translated Text
translated=$(echo $result | jq -r '.data.translated_text')
echo "Translated text: $translated"

# 6. Graceful Shutdown Server
curl -X POST http://localhost:8988/poweroff \
  -H "Content-Type: application/json" \
  -d '{"time": 5, "force": false}'
```

## More Information

- Complete API Specification: [openapi.json](openapi.json)
- Protocol Detailed Documentation: [REFERENCE.md](REFERENCE.md)
- Project Homepage: [README.md](README.md)
