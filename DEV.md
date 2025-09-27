# Worker 服务定义

服务器支持 gRPC 、HTTP 、WebSocket 三种协议。

# 环境变量

- `LOG_LEVEL`: 日志级别，默认 `info`，可选 `debug`、`info`、`warn`、`error`(不区分大小写)。
- `WROK_DIR`: 工作目录，默认 `./`，主要用于拼接在 poweron 传入的模型目录 path 路径前。
- `ENABLE_HTTP`: 是否启用 HTTP 服务，默认 `true`，可选值：`true`、`false`、0、1、'yes'、'no'(不区分大小写)。
- `ENABLE_WEBSOCKET`: 是否启用 WebSocket 服务，默认 `true`，可选值：`true`、`false`、0、1、'yes'、'no'(不区分大小写)。
- `ENABLE_GRPC`: 是否启用 gRPC 服务，默认 `true`，可选值：`true`、`false`、0、1、'yes'、'no'(不区分大小写)。
- `HTTP_HOST`: HTTP 服务主机，默认 `0.0.0.0`，仅在 `ENABLE_HTTP` 为真时生效。
- `HTTP_PORT`: HTTP 服务端口，默认 `8989`，仅在 `ENABLE_HTTP` 为 `true` 时生效。
- `WEBSOCKET_HOST`: WebSocket 服务主机，默认 `0.0.0.0`，仅在 `ENABLE_WEBSOCKET` 为真时生效。
- `WEBSOCKET_PORT`: WebSocket 服务端口，默认 `8990`，仅在 `ENABLE_WEBSOCKET` 为真时生效。
- `GRPC_HOST`: gRPC 服务主机，默认 `0.0.0.0`，仅在 `ENABLE_GRPC` 为真时生效。
- `GRPC_PORT`: gRPC 服务端口，默认 `8991`，仅在 `ENABLE_GRPC` 为真时生效。

## 主要接口定义

### health - 健康检查

### poweron - 启动

参数：

- path: 相对于程序工作目录的模型目录路径，需要包含 lex*.bin, model*.bin, vocab*.spm(或 srcvocab*.spm, trgvocab\*.spm) 文件。

返回值错误代码定义：

- 0: 成功
- 1000: 参数错误
- 1001: 模型目录不存在
- 1002: 模型文件不完整
- 1003: 内部错误
- 1009：未知错误

### poweroff - 关闭

参数：

- time: 多少秒后关闭服务器
- force: 是否强制关闭(默认 false，会等待所有请求处理完成)

返回值错误代码定义：

- 0: 成功
- 1100: 参数不合法
- 1101: 等待请求处理完成
- 1109: 内部错误

### ready - 获取状态

返回是否启动

### compute - 翻译文本

主要参数：

- text: 待翻译文本
- html: 是否为 HTML 文本(默认 false)

返回值错误代码定义：

- 0: 成功
- 1200: 参数不合法
- 1201: 翻译失败
- 1209: 内部错误
