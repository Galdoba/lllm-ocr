# cmd/lllm-ocr — DOX

## Purpose

CLI entry point (`github.com/urfave/cli/v2`). Thin bootstrap that wires `internal/commands.GetCommands()` into `cli.App`.

## Ownership

Owns: `main` function only. Delegates all subcommand definitions and config logic to `internal/commands/` and `internal/config/`.

Depends on: `internal/commands/`, `internal/config/`.

## Local Contracts

- `main()` creates a `cli.App` with `Name`, `Usage`, and `Commands` from `commands.GetCommands()`.
- No subcommand definitions, no config helpers, no pipeline logic in this file.

## Verification

```
go build ./cmd/lllm-ocr/
go vet ./...
```

No unit tests in `cmd/` yet.

## Child DOX Index

None.
