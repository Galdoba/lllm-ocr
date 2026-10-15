# internal/config — DOX

## Purpose

Canonical home for ALL configuration types and config-related utilities. No internal package dependencies.

## Ownership

Owns: `CLIConfig`, `ExtractConfig`, `RecognizeConfig`, `PipelineConfig`, `CycleConfig`, `FilterConfig`, `MergeConfig`, `ParserConfig`, `MarkdownConfig`, `GetDefaultConfigPath`, `CreateDefaultConfig`, `LoadConfig`, `EnsureDefaultConfigExists`, `NewPipelineConfig`, `BuildOCRConfig`, `ValidateExtractConfig`, `ValidateRecognizeConfig`, `ValidateCycleConfig`, `ValidatePipelineConfig`, `ValidatePipeline`, `PageName`, `SourceStem`, `WorkspacePath`.

Depends on: nothing internal.

## Local Contracts

- Config path is platform-specific: Linux `$XDG_CONFIG_HOME`/`~/.config`, macOS `~/Library/Application Support`, Windows `%APPDATA%`; filename `config.json`; const `configDirName = "llllm-ocr"`.
- `NewPipelineConfig` maps flat `CLIConfig` → nested `PipelineConfig`; `CycleEnabled=false` forces `CycleConfig.MaxTokens = 0`.
- `RecognizeRetries` (default 3) and `RecognizeRetryTemp` (default "1") control OCR restart behavior on cycle detection.
- `RemoveBlocks` (CLI `remove_blocks`) is always forwarded to `FilterConfig.RemoveLabels`.
- `BuildOCRConfig` is a deprecated alias for `NewPipelineConfig` (backward compatibility).
- `SourceStem(path)` returns the filename stem without extension (e.g., "book.pdf" → "book").
- `WorkspacePath(sourcePath, explicitDir)` derives workspace directory: if explicitDir is non-empty and not ".", returns explicitDir; otherwise returns `<source_dir>/<stem>`.
- All functions are exported for use by `internal/commands/` and other packages.

## Work Guidance

- The default config embeds **hardcoded machine-specific model paths** (`CreateDefaultConfig`). Do not "fix" this without an explicit request — users edit `config.json` per machine.

## Verification

```
go build ./...
go vet ./...
go test ./internal/config/... -v
```

## Child DOX Index

None.
