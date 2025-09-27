# API Examples

This document provides quick usage examples for the three protocols of MTranCore.

## HTTP/REST API Examples

### 1. Health Check

```bash
curl http://localhost:8080/health
```

### 2. Load Translation Engine

```bash
curl -X POST http://localhost:8080/poweron \
  -H "Content-Type: application/json" \
  -d '{"path": "models/en-zh"}'
```

### 3. Check Ready Status

```bash
curl http://localhost:8080/ready
```

### 4. Translate Text

```bash
curl -X POST http://localhost:8080/compute \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello, world!", "html": false}'
```

### 5. Shutdown Server

```bash
# Graceful Shutdown (wait for requests to complete)
curl -X POST http://localhost:8080/poweroff \
  -H "Content-Type: application/json" \
  -d '{"time": 0, "force": false}'

# Force Shutdown
curl -X POST http://localhost:8080/poweroff \
  -H "Content-Type: application/json" \
  -d '{"time": 0, "force": true}'
```

## gRPC API Examples

Use `grpcurl` tool to test gRPC API:

### 1. Health Check

```bash
grpcurl -plaintext localhost:8081 translator.TranslatorService/Health
```

### 2. Load Translation Engine

```bash
grpcurl -plaintext -d '{"path": "models/en-zh"}' \
  localhost:8081 translator.TranslatorService/Poweron
```

### 3. Check Ready Status

```bash
grpcurl -plaintext localhost:8081 translator.TranslatorService/Ready
```

### 4. Translate Text

```bash
grpcurl -plaintext -d '{"text": "Hello, world!", "html": false}' \
  localhost:8081 translator.TranslatorService/Compute
```

### 5. Streaming Translation

```bash
grpcurl -plaintext -d @ localhost:8081 translator.TranslatorService/ComputeStream <<EOF
{"text": "Hello"}
{"text": "World"}
{"text": "Goodbye"}
EOF
```

### 6. Shutdown Server

```bash
grpcurl -plaintext -d '{"time": 0, "force": false}' \
  localhost:8081 translator.TranslatorService/Poweroff
```

## WebSocket API Examples

### Use JavaScript (Browser/Node.js)

```javascript
const ws = new WebSocket('ws://localhost:8082/ws');

ws.onopen = () => {
  console.log('Connected');
  
  // 1. Load Engine
  ws.send(JSON.stringify({
    type: 'poweron',
    data: { path: 'models/en-zh' }
  }));
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
    uri = "ws://localhost:8082/ws"
    async with websockets.connect(uri) as websocket:
        # 1. Load Engine
        await websocket.send(json.dumps({
            "type": "poweron",
            "data": {"path": "models/en-zh"}
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
wscat -c ws://localhost:8082/ws

# Send messages after connection:

# 1. Load Engine
> {"type":"poweron","data":{"path":"models/en-zh"}}

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
curl http://localhost:8080/health

# 2. Load Translation Model
curl -X POST http://localhost:8080/poweron \
  -H "Content-Type: application/json" \
  -d '{"path": "models/en-zh"}'

# 3. Wait for Model Loading
while true; do
  ready=$(curl -s http://localhost:8080/ready | jq -r '.data.ready')
  if [ "$ready" = "true" ]; then
    echo "Engine is ready"
    break
  fi
  echo "Waiting for engine..."
  sleep 1
done

# 4. Execute Translation
result=$(curl -s -X POST http://localhost:8080/compute \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello, world!", "html": false}')

echo "Translation result: $result"

# 5. Extract Translated Text
translated=$(echo $result | jq -r '.data.translated_text')
echo "Translated text: $translated"

# 6. Graceful Shutdown Server
curl -X POST http://localhost:8080/poweroff \
  -H "Content-Type: application/json" \
  -d '{"time": 5, "force": false}'
```

## More Information

- Complete API Specification: [openapi.json](openapi.json)
- Protocol Detailed Documentation: [REFERENCE.md](REFERENCE.md)
- Project Homepage: [README.md](README.md)
