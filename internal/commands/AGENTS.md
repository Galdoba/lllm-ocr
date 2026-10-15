# internal/commands — DOX

## Purpose

CLI subcommand definitions: each command is a separate file with its flags and action. Commands import `internal/config` for config types and helpers.

## Ownership

Owns: `NewConfigCommand`, `NewExtractCommand`, `NewOCRCommand`, `NewAssembleCommand`, `NewConvertCommand`, `GetCommands`.

Depends on: `internal/config/`, `layoutblock/`, `ocr/`, `ocr/cycle/`, `output/`.

## Local Contracts

- Each command file exports a single function: `NewXxxCommand() *cli.Command`.
- `GetCommands()` in `commands.go` returns all 5 commands as `[]*cli.Command`.
- Commands that need config import `internal/config` directly.
- `CLIConfig` and all config helpers live in `internal/config/`. All other config types (`ExtractConfig`, `RecognizeConfig`, `PipelineConfig`, `CycleConfig`, `FilterConfig`, `MergeConfig`, `ParserConfig`, `MarkdownConfig`) are also defined in `internal/config/`.
- **Workspace pattern**: Commands compute `workspacePath := config.WorkspacePath(sourcePath, explicitDir)` and create it with `os.MkdirAll(workspacePath, 0755)`. The workspace path is passed as `cfg.OutDir` to pipeline functions. Pipeline functions require zero changes.
- `assemble` command: `--input`/`-i` flag derives workspace from source PDF; `-d`/`--dir` overrides explicitly; neither defaults to current dir behavior.
- `convert` command: When `-o`/`--output` is not specified, defaults output to `<input_dir>/output.md`.

## Work Guidance

- Keep command Actions focused: validate flags, load config, call pipeline functions, print results.
- Do not move config logic into commands — it belongs in `internal/config/`.

## Verification

```
go build ./...
go vet ./...
```

No unit tests in `internal/commands/` yet.

## Child DOX Index

None.
