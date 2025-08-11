# pytextcodec

Go package for text encoding/decoding compatible with Python's codecs

## Overview

This package provides text encoding/decoding functionality compatible with Python's codecs. It can be used in the same way as `golang.org/x/text/transform`, allowing you to simply replace the Transformer.

## Installation

```bash
go get github.com/akihiroy/pytextcodec
```

## Usage

### CP932 Encoding

```go
import (
    "github.com/akihiroy/pytextcodec/japanese"
    "golang.org/x/text/transform"
)

// Using encoder
encoder := japanese.CP932.NewEncoder()
encoded, _, err := transform.Bytes(encoder, []byte("こんにちは"))

// Using decoder
decoder := japanese.CP932.NewDecoder()
decoded, _, err := transform.Bytes(decoder, encoded)

// Combining with transform.NewReader
reader := transform.NewReader(sjisFile, japanese.CP932.NewDecoder())
```

## Development

### Running Tests

First, generate test data:

```bash
uv run japanese/testdata/generate_testdata.py
```

Then run the tests:

```bash
make test
```

### Linting

```bash
make lint
```

### Viewing Test Coverage

```bash
make test-coverage
```

## Supported Encodings

- CP932 (Shift_JIS)

## License

MIT License
