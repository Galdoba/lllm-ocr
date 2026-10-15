# Refactoring suggestions for lllm-ocr

Context: the project is stable and is NOT expected to grow beyond small features and expanding `layoutblock` types. Therefore the guiding principle here is **deletion over abstraction** — reduce duplication, remove dead surface, tighten package purity, and keep the existing layered structure. No new frameworks, no new layers.

Overlaps with `known_issues.md` are referenced by issue number where relevant.

## Current architecture (keep this)

```
internal/config   flat CLIConfig (JSON) + nested configs + validation + paths   (no internal deps)
layoutblock       LayoutBlock/BBox/LabelType + parse/filter/merge transforms    (pure)
ocr/cycle         cycle detector kernel + HTML completion                        (mixing pure + I/O)
ocr               pdftoppm / llama-mtmd-cli wrappers + pipeline orchestrations  (I/O)
output            LayoutBlock -> Markdown                                        (pure)
internal/commands CLI subcommands,  cmd/lllm-ocr bootstrap
```

Dependency chain to preserve: `internal/config` ← `layoutblock`, `ocr`, `ocr/cycle`, `output` ← `internal/commands` ← `cmd/lllm-ocr`.

## 1. Delete first (dead / duplicated surface, lowest risk)

- **One retry implementation**: `ocr/recognize_retry.go`'s `StartRetry` + `RetryResult` are never called (#27). Delete them and keep only `RetryState`; `getArgsForAttempt` in `recognize.go` becomes the single temp-arg logic (#14). This also removes the redundant `os.Truncate` (#13).
- **Drop compat aliases nobody uses**: `ocr/cycle_detector.go` (`CycleDetector`, `NewCycleDetector`, `NewCycleDetectorWithConfig`) has no callers.
- **Remove the dead parser**: `layoutblock.ParsePage` + helpers (`renderInnerHTML`, `getAttr`, `parseBBox`) in `layoutblock.go` are unused — `BlockParser.Parse` is the live path (used by `convert` and `ocr --debug`). Keep one parsing entry point. `FindBody`/`getCoord` stay (used by merge/sort).
- **Remove the deprecated `BuildOCRConfig`**: replace its 4 command call-sites and `config_test.go` uses with `NewPipelineConfig`, then delete the alias.
- **One `pageName`**: `ocr/pageName` duplicates `config.PageName`; delete the local copy (`ocr/source.go`) and use `config.PageName` everywhere.

## 2. Purity & package responsibility (do after the deletions)

- **Make `ocr/cycle` truly pure** by moving HTML completion (`htmlfix.go`: `CompleteHTMLWithContent`, `fixHTMLNaivelyPath`, `cloneNode`) into `ocr/` (or an `ocr/html` file). Cycle detection then has no `os`/`io` dependency and its tests run in isolation. Move `detector_test.go`'s `TestCompleteHTMLWithContent_*` cases with the code. Then `ocr/cycle` shrinks to: config alias, `Detector`, `countRepeats`, `findCycle`, `GetCleanTokens`, `GetCycleInfo`.
- **Remove side effects from config construction**: `NewPipelineConfig` bakes in a `ProgressFn` that prints to stdout. Leave that defaulting to the commands layer so `internal/config` stays side-effect-free. Do not design a progress abstraction — a default closure in `internal/commands` is enough.
- End state: only `ocr/` and `internal/commands/` touch the filesystem/processes; `layoutblock`, `ocr/cycle`, `output`, `internal/config` are pure and unit-testable.

## 3. Simplifications that reduce concurrency noise

- **Make cycle detection deterministic** (#3 + #30): don't spawn a goroutine per check interval with a full buffer copy (`cycle/detector.go:47-51`). Run `detectCycling()` synchronously inside `AddToken` at check boundaries — it's a pure function over a token slice, so there is no reason for async. This removes the 300 ms sleeps in `detector_test.go`, the `-race` flakes, and the O(n²) snapshot growth. Keep the `sync.RWMutex` only if you keep concurrent access; with synchronous detection the mutex can also be dropped.
- **Model a per-page outcome explicitly**: have `RecognizeSinglePage` return a small result (status, attempts, whether HTML was intentionally completed) instead of `nil` + side-effected file. That gives the `ocr` command a real signal to exit nonzero on failed pages (#25) and to distinguish crash-recovery from intentional completion (#7) — without inventing an error-type hierarchy.

## 4. Config & CLI simplification

- **Actually validate**: call `config.ValidatePipeline` once after `LoadConfig` in every command action (fixes #20 divide-by-zero; the validator already exists but is dead).
- **One shared command helper**: `extract`, `ocr`, `assemble`, `convert` all re-implement "ensure default config exists + load config + derive workspace `MkdirAll`". Extract a `loadConfig(cCtx)` and `workspaceFor(cCtx, inputPath)` in `internal/commands`. Small, safe, cuts ~4x copy-paste.
- **Trim re-export noise (optional)**: `layoutblock` re-defines `FilterConfig`/`MergeConfig`/`Default*` wrappers around `internal/config`. Since `layoutblock` already imports `internal/config`, either use `config.*` directly or keep the short aliases deliberately — but don't maintain two default-factory locations (`DefaultFilterConfig` exists in both packages).

## 5. Growth path — this is where future work happens

`layoutblock` is the only area expected to grow. Keep it single-source and friction-free:

- One parser entry (`BlockParser.Parse`); the file-vs-HTML heuristic and the (currently inverted) lenient semantics from #21 get fixed once, here.
- One `LabelType` registry: a new label = string constant + `supportedLabels` entry + optional filter/merge rule + an `output` rendering case. Document this checklist in `layoutblock/AGENTS.md` next to `supportedLabels` (the two-step rule is already noted).
- Prefer new pure transforms in `layoutblock`/`output` over new CLI commands. New external tools = a wrapper method + a flag in the existing pipeline (as `pdftoppm`/`llama-mtmd-cli` are today), not a new package.
- Once `ocr/cycle` is pure, leave it alone except for config toggles (`MaxTokens=0` disables detection).

## 6. Verification budget

- After the detector change: `go test -race ./ocr/cycle/... ./layoutblock/...` — must be deterministic, no sleeps.
- Add regression tests for the pure functions touched in the staged plan: `countRepeats` (#4), TOC merge (#22), parser lenient semantics (#21).
- Keep `go build ./...` + `go vet ./...` as the repo gate; make `go test ./... -race` a habit for PRs touching `ocr/` or `ocr/cycle`.

## Explicitly NOT recommended (matches "not expected to grow")

- No converter/detector plugin interfaces, no DI framework, no event bus.
- No config-schema/validation library, no YAML/DB config migration — `CLIConfig` JSON is fine.
- No command-framework swap (keep `urfave/cli`), no multi-module split.
- No general-purpose concurrency framework in `ocr/` — the pipeline stays sequential on purpose (`ocr/AGENTS.md`).

## Suggested sequencing

1. Section 1 (deletions) + section 4 (validation + command helpers) — smallest, lowest risk, no behavior change.
2. Section 3 (deterministic detector + per-page result model) — unblocks known issues #3/#24/#7 and the `-race` clean-up.
3. Section 2 (purity moves: htmlfix relocation, progress defaulting) — cosmetic; do last so renames don't churn earlier work.

Each step must keep `go build ./... && go vet ./... && go test ./...` green, and must be followed by a DOX pass (update the owning `AGENTS.md`) whenever identifiers move between packages.