# MT - 命令行翻译工具

一个简单的命令行翻译工具，基于 Bergamot 翻译引擎。

## 构建

```bash
make build-mt
```

编译后的二进制文件位于 `build/mt`。

## 使用方法

### 基本用法

使用模型目录（自动发现模型文件）：

```bash
./build/mt -model ./models/enzh -text "Hello, world!"
```

### REPL 交互模式

启动交互式翻译界面，可以持续输入文本并获得翻译结果：

```bash
./build/mt -model ./models/enzh -r
```

在 REPL 模式下：
- 输入文本后按回车即可获得翻译
- 输入 `exit` 或 `quit` 退出
- 按 `Ctrl+D` 也可退出
- 支持 HTML 模式：`./build/mt -model ./models/enzh -r -html`

### 使用独立文件

如果需要手动指定每个模型文件：

```bash
./build/mt \
  -model-file ./models/enzh/model.enzh.intgemm.alphas.bin \
  -shortlist ./models/enzh/lex.50.50.enzh.s2t.bin \
  -vocab-src ./models/enzh/srcvocab.enzh.spm \
  -vocab-trg ./models/enzh/trgvocab.enzh.spm \
  -text "Hello, world!"
```

### HTML 翻译

翻译 HTML 内容：

```bash
./build/mt -model ./models/enzh -html -text "<p>Hello, world!</p>"
```

### 从标准输入读取（配合管道）

```bash
echo "Hello, world!" | xargs -I {} ./build/mt -model ./models/enzh -text "{}"
```

## 命令行选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-model` | 模型目录路径（自动发现文件） | - |
| `-model-file` | 模型文件路径 (.bin) | - |
| `-shortlist` | 词汇表快捷列表文件路径 (lex*.bin) | - |
| `-vocab-src` | 源语言词汇表文件路径 (.spm) | - |
| `-vocab-trg` | 目标语言词汇表文件路径 (.spm) | - |
| `-vocab` | 词汇表文件路径 (.spm) | - |
| `-text` | 要翻译的文本（非 REPL 模式下必需） | - |
| `-html` | 将输入视为 HTML | false |
| `-cache` | 翻译缓存大小 | 1024 |
| `-repl` | 启动 REPL（交互式）模式 | false |

## 示例

### 英译中

```bash
./build/mt -model ./models/enzh -text "Good morning!"
```

输出：
```
早上好！
```

### REPL 交互模式示例

```bash
./build/mt -model ./models/enzh -r
```

交互示例：
```
Bergamot Translator REPL Mode
Enter text to translate (type 'exit', 'quit', or press Ctrl+D to exit)

> Hello, world!
你好，世界！

> How are you?
你好吗？

> Good morning!
早上好！

> exit
Goodbye!
```

### 批量翻译（脚本示例）

创建一个文件 `translate.sh`：

```bash
#!/bin/bash
MODEL_DIR="./models/enzh"

while IFS= read -r line; do
  ./build/mt -model "$MODEL_DIR" -text "$line"
done < input.txt > output.txt
```

## 注意事项

1. **模型文件**：确保模型目录包含必要的文件：
   - 模型文件：`*.bin`（不以 lex 开头）
   - 词汇表快捷列表：`lex*.bin`
   - 词汇表文件：至少一个 `*.spm` 文件

2. **性能**：
   - **单次翻译模式**：每次启动都需要加载模型，适合快速测试和脚本调用
   - **REPL 模式**：模型只加载一次，适合需要多次翻译的交互场景，性能更优
   - **Worker 服务**：如果需要高性能批量翻译或生产环境使用，建议使用 worker 服务模式

3. **内存使用**：翻译引擎会占用一定内存，根据模型大小而定。

## 与 Worker 服务的区别

| 特性 | MT 工具（单次） | MT 工具（REPL） | Worker 服务 |
|------|----------------|-----------------|------------|
| 用途 | 单次命令行翻译 | 交互式翻译 | 长期运行服务 |
| 接口 | 命令行参数 | 命令行交互 | HTTP/gRPC/WebSocket |
| 性能 | 每次启动加载模型 | 模型加载一次 | 模型常驻内存 |
| 适用场景 | 脚本、快速测试 | 本地交互、多次翻译 | 生产环境、高并发 |
