# ocr/cycle — DOX

## Purpose

Pure, dependency-free generation/cycle detection for streaming vision-LLM output, plus HTML completion after a cycle is cut. No external tool or filesystem dependency in the core algorithm.

## Ownership

Owns: `Detector` (`cycle_detector.go`), `IsCycling`, `GetCleanTokens`, `GetCycleInfo`, and `htmlfix.go` (`completeHTMLWithContent`, `fixHTMLNaivelyPath`). `Config` is a type alias to `internal/config/CycleConfig`; `DefaultConfig()` delegates to `internal/config.DefaultCycleConfig()`.

Depends on: `internal/config/`. Nothing else internal. Consumed by `ocr.RecognizePages`.

## Local Contracts

- Detects repeating patterns of **3–120 tokens** with **≥3 consecutive repeats**; configurable via `CycleConfig{CheckInterval, MinTokens, MaxTokens}`.
- Thread-safe via `sync.RWMutex`; fed token-by-token through `AddToken` (never poll files).
- `MaxTokens` bounds the rolling buffer; feeding `ocr` sets it to `0` to disable.
- `completeHTMLWithContent` parses and wraps bare elements in `<body>`; falls back to `fixHTMLNaivelyPath` (append `</body></html>`) only if parsing fails.

## Work Guidance

- Keep the detection core free of I/O so it stays unit-testable in isolation.
- `detector_test.go` exercises `countRepeats` directly — preserve that testable surface.

## Verification

```
go test ./ocr/cycle/...
go vet ./...
```

## Child DOX Index

None.
