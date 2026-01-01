> Merged into [MTranServer](https://github.com/xxnuo/MTranServer), this repository is no longer updated.

# MTranCore

A translator core library based on webassembly and Go for cross-platform applications.

## Quick Start

Auto-load translation model on startup:

```bash
# Using model directory
./worker -model-dir models/enzh

# Using individual file paths
./worker \
  -model-path models/enzh/model.enzh.intgemm.alphas.bin \
  -lexical-shortlist-path models/enzh/lex.50.50.enzh.s2t.bin \
  -vocabulary-path models/enzh/vocab.enzh.spm

# Translate text via HTTP
curl -X POST http://localhost:8988/trans \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello, world!", "html": false}'
```

### Documentation

- **Simplified API Guide**: [SLIM.md](SLIM.md) - Quick reference for downstream integration
- **OpenAPI Specification**: [openapi.json](openapi.json) - Complete API specification in OpenAPI 3.0 format
- **Examples**: [EXAMPLES.md](EXAMPLES.md) - Quick start guide with practical examples for all three protocols
- **Protocol Reference**: [REFERENCE.md](REFERENCE.md) - Detailed protocol documentation and error codes

## Thanks

[KSpaceer](https://github.com/KSpaceer) for the original golang [library](https://github.com/KSpaceer/gobergamot).

[Bergamot Project](https://browser.mt/) for awesome idea of local translation.

[Mozilla](https://github.com/mozilla) for the [models](https://github.com/mozilla/firefox-translations-models).