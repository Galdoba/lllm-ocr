# lllm-ocr

**PDF → OCR → structured blocks → Markdown/Text/HTML/JSON**

A high-performance OCR pipeline using `pdftoppm` for image extraction and `llama-mtmd-cli` for text recognition, with automatic cycle detection and HTML→Markdown conversion.

![Go](https://img.shields.io/badge/go-1.25.1-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## Features

- **PDF Extraction**: Converts PDF pages to PNG images using `pdftoppm`
- **OCR Recognition**: Uses `llama-mtmd-cli` for text recognition with retry logic
- **Cycle Detection**: Detects and handles text recognition cycles for better accuracy
- **Block Parsing**: Parses recognized text into structured blocks with bounding boxes
- **Multi-format Output**: Convert to Markdown, Text, HTML, or JSON
- **Config-driven**: JSON configuration with platform-specific defaults
- **Page Selection**: Extract and process specific pages or page ranges

## Installation

### Prerequisites

1. **Go 1.25.1+**
   ```bash
   go version
   ```

2. **pdftoppm** (poppler-utils)
   - **Debian/Ubuntu**: `sudo apt-get install poppler-utils`
   - **macOS**: `brew install poppler`
   - **Arch Linux**: `sudo pacman -S poppler`

3. **llama-mtmd-cli**
   Download from: https://github.com/Mozilla-Ocho/llama-mtmd-cli
   Or build from source:
   ```bash
   git clone https://github.com/Mozilla-Ocho/llama-mtmd-cli.git
   cd llama-mtmd-cli && go build -o llama-mtmd-cli .
   ```

### Building

```bash
# Clone and build
git clone https://github.com/Galdoba/lllm-ocr.git
cd lllm-ocr
go build -o lllm-ocr ./cmd/lllm-ocr/

# Or install globally
go install ./cmd/lllm-ocr/
```

## Quick Start

```bash
# 1. Extract pages from PDF to PNG
./lllm-ocr extract -i input.pdf -p 1-10

# 2. Recognize text from PNG pages
./lllm-ocr ocr -i input.pdf -p 1-10

# 3. Assemble combined HTML
./lllm-ocr assemble -i input.pdf

# 4. Convert to Markdown
./lllm-ocr convert -i combined.html -f markdown -o output.md
```

## Command Reference

### extract

Extract pages from a PDF into PNG images.

```bash
lllm-ocr extract -i <input.pdf> [flags]
```

**Flags:**
- `-i, --input <path>`: Path to input PDF file (required)
- `-c, --config <path>`: Path to config file (default: `~/.config/llllm-ocr/config.json`)
- `-p, --pages <spec>`: Pages or ranges (default: `1-999`, e.g., `1-3,5,6-8`)
- `-d, --dir <path>`: Working directory for PNG output (default: `.`)

### ocr

Recognize text from extracted PNG pages and produce HTML sources.

```bash
lllm-ocr ocr -i <input.pdf> [flags]
```

**Flags:**
- `-i, --input <path>`: Path to input PDF file (required)
- `-c, --config <path>`: Path to config file
- `-p, --pages <spec>`: Pages or ranges to recognize
- `-d, --dir <path>`: Working directory containing PNG files
- `--debug`: Visualize recognized blocks on each page

### assemble

Build `combined.html` from page sources.

```bash
lllm-ocr assemble -i <input.pdf> [flags]
```

**Flags:**
- `-i, --input <path>`: Path to source PDF (for auto-deriving workspace)
- `-c, --config <path>`: Path to config file
- `-p, --pages <spec>`: Pages or ranges to include
- `-d, --dir <path>`: Working directory containing `pageSources/`

### convert

Convert `combined.html` to Markdown.

```bash
lllm-ocr convert -i <combined.html> -f markdown [flags]
```

**Flags:**
- `-i, --input <path>`: Path to input HTML file (required)
- `-f, --format <format>`: Output format (required, currently supports: `markdown`)
- `-o, --output <path>`: Output file path (default: `<input_dir>/output.md`)
- `-c, --config <path>`: Path to config file
- `--lenient`: Skip unsupported labels during parsing

### config

Create a default configuration file.

```bash
lllm-ocr config
```

This creates `~/.config/llllm-ocr/config.json` with default settings.

## Configuration

Default config location:
- **Linux**: `~/.config/llllm-ocr/config.json`
- **macOS**: `~/Library/Application Support/llllm-ocr/config.json`
- **Windows**: `%APPDATA%/llllm-ocr/config.json`

### Config Schema

```json
{
  "pdftoppm_cmd": "pdftoppm",
  "llama_cli_cmd": "llama-mtmd-cli",
  "llama_cli_args": [
    "-m", "/path/to/model.gguf",
    "--mmproj", "/path/to/mmproj.gguf",
    "other", "parameters"
  ],
  "remove_blocks": ["caption", "image", "page_header", "page_footer", "form"],
  "ocr_max_width": 2048,
  "ocr_max_height": 3072,
  "cycle_enabled": true,
  "cycle_check_interval": 100,
  "cycle_min_tokens": 400,
  "cycle_max_tokens": 10000,
  "recognize_retries": 3,
  "recognize_retry_temp": "1"
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `pdftoppm_cmd` | string | `pdftoppm` | Command to run pdftoppm |
| `llama_cli_cmd` | string | `llama-mtmd-cli` | Command to run llama-mtmd-cli |
| `llama_cli_args` | array | `["-m", "/path/to/model.gguf", ...]` | Arguments for llama-mtmd-cli |
| `remove_blocks` | array | `["caption", "image", ...]` | Block types to remove during parsing |
| `ocr_max_width` | int | `2048` | Maximum OCR width |
| `ocr_max_height` | int | `3072` | Maximum OCR height |
| `cycle_enabled` | boolean | `true` | Enable cycle detection |
| `cycle_check_interval` | int | `100` | Token interval for cycle checks |
| `cycle_min_tokens` | int | `400` | Minimum tokens before cycle detection |
| `cycle_max_tokens` | int | `10000` | Maximum tokens for cycle detection |
| `recognize_retries` | int | `3` | Number of retries on failure |
| `recognize_retry_temp` | string | `"1"` | Temperature for retry recognition |

**Note**: Edit `llama_cli_args[0]` and `llama_cli_args[2]` to set your actual model and mmproj paths.

## Workflow

```
PDF Input
    ↓
extract → PNG files (workspace/)
    ↓
ocr → HTML files (pageSources/)
    ↓
assemble → combined.html (workspace/)
    ↓
convert → output.md (workspace/)
```

### Example Workflow

```bash
# Step 1: Extract first 10 pages as PNG
./lllm-ocr extract -i "documents/book.pdf" -p 1-10

# Step 2: Recognize text from extracted pages
./lllm-ocr ocr -i "documents/book.pdf" -p 1-10

# Step 3: Assemble combined HTML
./lllm-ocr assemble -i "documents/book.pdf"

# Step 4: Convert to Markdown
./lllm-ocr convert -i "documents/book/combined.html" -f markdown
```

## Directory Structure

```
project-root/
├── config.json              # Configuration (created automatically)
├── pageSources/            # HTML from OCR (after 'ocr' command)
│   ├── page_001.html
│   ├── page_002.html
│   └── ...
├── combined.html           # Combined HTML (after 'assemble' command)
└── output.md               # Final Markdown (after 'convert' command)
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     lllm-ocr (CLI)                          │
├─────────────────────────────────────────────────────────────┤
│  extract  ──►  ocr  ──►  assemble  ──►  convert            │
│      │            │            │             │              │
│  pdftoppm   llama-mtmd  HTML merge   Block parser         │
└─────────────────────────────────────────────────────────────┘
```

### Components

- **layoutblock/**: Core domain logic for `LayoutBlock`, `BBox`, `LabelType`
- **ocr/**: External tool wrappers and pipeline orchestration
- **ocr/cycle/**: Cycle detection and HTML completion
- **output/**: LayoutBlock → Markdown conversion
- **internal/config/**: Configuration types and helpers
- **internal/commands/**: CLI subcommand definitions

## Testing

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run specific package
go test ./layoutblock/...
go test ./ocr/...
```

## Known Issues

See [`doc/known_issues.md`](doc/known_issues.md) for a comprehensive list of known issues and workarounds.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Please read the documentation in [`doc/`](doc/) for:
- [Pipeline Overview](doc/pipeline.md)
- [Specification](doc/specification.md)
- [Feedback & Suggestions](doc/feedback_suggestions.md)
- [Refactoring Suggestions](doc/refactoring_suggestions.md)

## Dependencies

- **pdftoppm** (poppler-utils)
- **llama-mtmd-cli**
- **Go** 1.25.1+

### Go Modules

```
github.com/JohannesKaufmann/html-to-markdown v1.6.0
github.com/urfave/cli/v2 v2.27.7
golang.org/x/image v0.45.0
golang.org/x/net v0.58.0
```
# lllm-ocr
