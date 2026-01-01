# MT - Command Line Translation Tool

A simple command line translation tool based on the Bergamot translation engine.

## Build

```bash
make build-mt
```

The compiled binary is located at `build/mt`.

## Usage

### Basic Usage

Use model directory (automatically discovers model files):

```bash
./build/mt -model ./models/enzh -text "Hello, world!"
```

### REPL Interactive Mode

Launch interactive translation interface for continuous text input and translation:

```bash
./build/mt -model ./models/enzh -r
```

In REPL mode:

- Enter text and press Enter to get translation
- Type `exit` or `quit` to exit
- Press `Ctrl+D` to exit
- Supports HTML mode: `./build/mt -model ./models/enzh -r -html`

### Using Separate Files

To manually specify each model file:

```bash
./build/mt \
  -model-file ./models/enzh/model.enzh.intgemm.alphas.bin \
  -shortlist ./models/enzh/lex.50.50.enzh.s2t.bin \
  -vocab-src ./models/enzh/srcvocab.enzh.spm \
  -vocab-trg ./models/enzh/trgvocab.enzh.spm \
  -text "Hello, world!"
```

### HTML Translation

Translate HTML content:

```bash
./build/mt -model ./models/enzh -html -text "<p>Hello, world!</p>"
```

### Read from Standard Input (with pipes)

```bash
echo "Hello, world!" | xargs -I {} ./build/mt -model ./models/enzh -text "{}"
```

## Command Line Options

| Option        | Description                                   | Default |
| ------------- | --------------------------------------------- | ------- |
| `-model`      | Model directory path (auto-discovers files)   | -       |
| `-model-file` | Model file path (.bin)                        | -       |
| `-shortlist`  | Vocabulary shortlist file path (lex\*.bin)    | -       |
| `-vocab-src`  | Source language vocabulary file path (.spm)   | -       |
| `-vocab-trg`  | Target language vocabulary file path (.spm)   | -       |
| `-vocab`      | Vocabulary file path (.spm)                   | -       |
| `-text`       | Text to translate (required in non-REPL mode) | -       |
| `-html`       | Treat input as HTML                           | false   |
| `-cache`      | Translation cache size                        | 1024    |
| `-repl`       | Start REPL (interactive) mode                 | false   |

## Examples

### English to Chinese

```bash
./build/mt -model ./models/enzh -text "Good morning!"
```

Output:

```
早上好！
```

### REPL Interactive Mode Example

```bash
./build/mt -model ./models/enzh -r
```

Interactive example:

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

### Batch Translation (Script Example)

Create a file `translate.sh`:

```bash
#!/bin/bash
MODEL_DIR="./models/enzh"

while IFS= read -r line; do
  ./build/mt -model "$MODEL_DIR" -text "$line"
done < input.txt > output.txt
```

## Notes

1. **Model Files**: Ensure the model directory contains the necessary files:

   - Model file: `*.bin` (not starting with lex)
   - Vocabulary shortlist: `lex*.bin`
   - Vocabulary files: at least one `*.spm` file

2. **Performance**:

   - **Single Translation Mode**: Model loads on each startup, suitable for quick testing and script calls
   - **REPL Mode**: Model loads only once, suitable for interactive scenarios requiring multiple translations, better performance
   - **Worker Service**: For high-performance batch translation or production environment use, the worker service mode is recommended

3. **Memory Usage**: The translation engine will consume memory depending on the model size.

## Differences from Worker Service

| Feature     | MT Tool (Single)         | MT Tool (REPL)                           | Worker Service                           |
| ----------- | ------------------------ | ---------------------------------------- | ---------------------------------------- |
| Purpose     | One-time CLI translation | Interactive translation                  | Long-running service                     |
| Interface   | CLI arguments            | CLI interaction                          | HTTP/gRPC/WebSocket                      |
| Performance | Model loads each startup | Model loads once                         | Model resides in memory                  |
| Use Cases   | Scripts, quick testing   | Local interaction, multiple translations | Production environment, high concurrency |
