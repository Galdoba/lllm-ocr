# output — DOX

## Purpose

Converts `layoutblock.LayoutBlock` slices into Markdown output.

## Ownership

Owns: `Output` interface, `NewMarkdownOutput`, `MarkdownOutput`, `MarkdownConfig`, `Convert`, and all post-processing in `markdown_clean.go` (list markers, table headers, identical-header table merging). Config type (`MarkdownConfig`) is a type alias to `internal/config/`.

Depends on: `internal/config/`, `layoutblock/` + `github.com/JohannesKaufmann/html-to-markdown`.

## Local Contracts

- Only the **markdown** format is supported (the `format != "markdown"` guard lives in `cmd/lllm-ocr/main.go`, not here).
- `MarkdownConfig`: `PageSeparator`, `MathBlockMode`, `HideTables`.
- A single `md.Converter` is created once per `Convert()` call and reused across all blocks.
- Complex tables (containing `colspan`/`rowspan`) are rendered as callouts (`convertTableToCallout`).
- `<math>` content is preserved via `processMath`.

## Work Guidance

- Post-processing lives in `markdown_clean.go`; keep regex-based cleanup isolated there.
- Do not add new output formats without adding a parallel converter and a format guard in the CLI.

## Verification

```
go test ./output/...
go build ./...
```

## Child DOX Index

None.
