# ocr — DOX

## Purpose

Wraps external tools (`pdftoppm`, `llama-mtmd-cli`) and provides pipeline functions: `ExtractPages` (PDF → PNG), `RecognizePages` (PNG → HTML), `Assemble` (build `combined.html`). Bridges `layoutblock`/`ocr/cycle`/`output` to the CLI.

## Ownership

Owns: `ExtractPages`, `RecognizePages`, `Assemble`, `AssembleReport`, `pageName`, and page selection (`ParsePageSpec`, `SelectPages`, `PdfPageCount`). Config types (`ExtractConfig`, `RecognizeConfig`, `Config`) are type aliases to `internal/config/`.

Depends on: `internal/config/`, `layoutblock/`, `ocr/cycle/`.

## Local Contracts

- **Two-pass extraction is mandatory**: `ExtractConfig.OcrMaxWidth` and `OcrMaxHeight` must both be `> 0` (probe @scale 1000 → compute scale → final extraction preserving aspect ratio).
- Page filenames are zero-padded (`page_001.png`, `page_001.html`) via `pageName`.
- Sources land in `<dir>/pageSources/`; `Assemble` globs `pageSources/page_*.html`, keeps only those in the passed page set, sorts numerically, and silently skips pages with no recognized blocks.
- `ExtractPages`/`RecognizePages`/`Assemble` take an explicit `[]int` page list (not a `from`/`to` range). Page specs are parsed with `ParsePageSpec` (comma-separated pages/inclusive ranges) and filtered against PDF page count with `SelectPages`.
- **Retry on cycle**: `RecognizePages` orchestrates per-page recognition with automatic retry on cycle detection. Each page gets its own `RetryState` (max retries, temp adjustment). On cycle: restart OCR with `--temp <RetryTempArgs>`, log to stderr. If max retries exceeded, complete HTML with available content.
- `CycleConfig.MaxTokens == 0` disables detection. `cycle_detector.go` is a backwards-compat alias to `ocr/cycle.Detector` — keep it in sync.

## Work Guidance

- `ExtractPages` and `RecognizePages` report progress via `PipelineConfig.ProgressFn(current, total)`.
- Do not add parallelism to the sequential pipeline (known issue #11 in `doc/known_issues.md`) without an explicit request.

## Verification

```
go build ./...
go vet ./...
```

No unit tests in `ocr/` except `pages_test.go` (covers `ParsePageSpec` / `SelectPages`).

## Child DOX Index

- `ocr/cycle/` — pure cycle-detection algorithm and HTML completion.
