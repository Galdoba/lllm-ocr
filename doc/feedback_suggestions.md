# Process feedback & logging guide for lllm-ocr

Goal: give the user a clear, consistent picture of what the pipeline is doing, in two visibility modes (`minimum`, `verbose`), without flooding output. This is a guide + suggested change list, aligned with the fix plan in `known_issues.md` (progress inconsistency #12, swallowed errors #25) and the suppression of incidental stderr token echo.

## Principles

1. **Two channels, two purposes.**
   - `stdout` — user-facing progress, results, artifacts (paths/counts), and final summaries.
   - `stderr` — warnings, errors, and all verbose/debug diagnostics.
2. **Two modes, one knob.** Add a single `--verbose`/`-v` flag at the `cli.App` level so every command inherits it:
   - **minimum** (default): milestones the user must see to follow the run and act on failures.
   - **verbose**: adds everything else — exact commands, derived values, per-item detail.
3. **Errors are not optional.** Every failure path must produce a message and (where a phase failed) a nonzero exit (#25). Noise is what verbose mode is for; errors are not noise.
4. **No per-token logging.** The current `io.MultiWriter(htmlFile, os.Stderr)` in `recognize.go:70` streams every emitted token to stderr — remove the `os.Stderr` sink entirely (or gate it behind an explicit `debug` flag). Report an aggregate token count instead.
5. **One progress format.** `ExtractPages` reports `i+1`, `RecognizePages` reports the real page number (#12). Use one format everywhere: `page 3/42 (origin 42)` or `page 42/42` — pick one and apply to both phases.
6. Use the standard library `log/slog` as the backend (Go 1.25, no new deps). Progress/results go to `fmt.Fprintf(os.Stdout, …)`; diagnostics go through `slog` (Info/Warn/Error to stderr; Debug enabled by `--verbose`).

## Minimum mode — what the user always sees

Progress (stdout), one short line per milestone endpoint:

| # | Message | Stage |
|---|---|---|
| 1 | `config: <path> loaded` (only if custom `-c` given; otherwise silent) | CLI |
| 2 | `pdf: <path> — 412 pages` | CLI |
| 3 | `pages selected: 412 (spec 1-412)` | CLI |
| 4 | `extracting page 7/412 …` (live progress, one line per page) | extract |
| 5 | `extract: wrote page_007.png (3.1 MB)` | extract |
| 6 | `recognizing page 7/412 …` (live progress, same format as #4) | recognize |
| 7 | `recognize: page 7 done (843 tokens, 482 ms)` | recognize |
| 8 | `recognize: page 7 cycle detected — retry 1/3` (then 2/3 …) | recognize |
| 9 | `recognize: page 7 max retries (3) exceeded` — warning | recognize |
| 10 | `recognize: page 7 command failed: <err>` — error, page NOT completed (#7/#25) | recognize |
| 11 | `assemble: including page 7` | assemble |
| 12 | `assemble: skipping page 7 (no recognized blocks)` | assemble |
| 13 | `assemble: wrote combined.html (302 pages, 1.2 MB)` | assemble |
| 14 | `convert: parsed 3 412 blocks (2 unsupported labels skipped)` | convert |
| 15 | `convert: wrote output.md (128 KB)` | convert |
| 16 | Phase summaries: `extract complete → <dir>`, `ocr complete → <dir>/pageSources`, `assemble complete → <dir>/combined.html` | all |
| 17 | Every validation error, config error, missing PDF/image, glob failure — as an error line that also aborts with nonzero exit | all |

## Verbose mode — what it adds

Everything below is `Debug` level (stderr), shown only with `--verbose`:

- Exact commands spawned: `pdftoppm -png -scale-to 1000 -f 7 -l 7 …`, `llama-mtmd-cli -m … --image page_007.png`, `pdfinfo <path>`.
- Extract internals: probe dimensions `1000px @ scale 1000`, computed `target_scale = 2048`, final rename `page_007-1.png → page_007.png`, per-file size.
- Recognize internals: detector config (`min_tokens=400`, `check_interval=100`, `max_tokens=10000`), token count at cycle detection, cycle details (`pattern_len=10 repeats=5 cycle_start=370`), temp-arg change per retry (`--temp 0 → 1`), HTML completion path + byte count.
- Assemble internals: glob matched `N` files in `pageSources/`, pages dropped as out-of-range, empty vs present `data-bbox` reasoning.
- Convert internals: block count after parse/filter/merge (`parsed 3412 → filtered 3011 → merged 2887`), the exact unsupported label string (e.g. `unsupported label at page 7: "chart"`), `html-to-markdown` fallbacks (`convertHeader` → text on parse failure).
- Per-page timing (wall clock) for extract, recognize, convert.

Errors from subprocesses always go to **minimum** (`slog.Error`/`Warn`), and in verbose the failing command's args are repeated in the error context so logs are self-contained.

## Feedback point map — every touchpoint the pipeline should report

| Stage | Touchpoint | Minimum? | Verbose-only? | Example |
|---|---|---|---|---|
| CLI | config load, validation | ✔ (errors only) | ✔ (path + resolved fields) | `validating config: cycle.check_interval must be > 0` |
| CLI | pdfinfo page count | ✔ | ✔ | `pdf: 412 pages` |
| CLI | page spec resolution | — | ✔ | `pages: [1 2 3 5 …]` |
| CLI | workspace mkdir | — | ✔ | `mkdir ~/book` |
| extract | probe run | — | ✔ | `pdftoppm probe scale=1000` |
| extract | probe dims → scale | — | ✔ | `probe 2480x3508 → target_scale 2048` |
| extract | final render | ✔ (progress) | ✔ (argv) | `extracting page 7/412` |
| extract | rename + size check | ✔ (result) | ✔ (sizes) | `wrote page_007.png` |
| recognize | image existence | ✔ (error only) | — | `image page_007.png not found` |
| recognize | spawn llama | — | ✔ | `argv: llama-mtmd-cli …` |
| recognize | per-page output | ✔ (summary) | ✔ (token count) | `page 7 done (843 tokens)` |
| recognize | cycle + kill | ✔ (warning) | ✔ (pattern details) | `cycle detected, retrying 1/3` |
| recognize | retry budget spent | ✔ (warning) | — | `max retries exceeded — keeping partial page` |
| recognize | non-cycle command error | ✔ (error, abort) | — | `page 7 command failed: exit 1` |
| recognize | HTML completion | — | ✔ | `completed output 84.2 KB` |
| assemble | glob + selection | — | ✔ | `matched 414 sources, selected 412` |
| assemble | per-page skip | ✔ | — | `skipping page 8 (no recognized blocks)` |
| assemble | combined write | ✔ | ✔ (bytes) | `wrote combined.html (302 pages)` |
| convert | parse/filter/merge | ✔ (counts) | ✔ (per-label) | `3 412 → 3 011 blocks` |
| convert | markdown write | ✔ | ✔ (bytes) | `wrote output.md` |
| any | phase failure = error + nonzero exit (#25) | ✔ | — | `ocr: recognizing page 3: …` |

## Implementation sketch (kept small — the project is not expected to grow)

1. Add `verbose` (and optionally `quiet`) bool flags to the `cli.App` in `cmd/lllm-ocr/main.go`; they propagate to every command via the shared `cli.Context`.
2. Add a tiny `internal/feedback` package (single file) exposing `Progressf` (→ stdout), and `Debugf/Infof/Warnf/Errorf` (→ slog, stderr), with the level fixed once from the flag. It is the single place that decides minimum vs verbose.
3. In `internal/commands/`: create `loadConfig(cCtx)` / `workspaceFor(cCtx, input)` helpers (see `refactoring_suggestions.md` §4) and route every `fmt.Print*` through `feedback.*`.
4. In `ocr/`:
   - Replace the default `ProgressFn` in `config.NewPipelineConfig` (move printing out of `internal/config`; per `refactoring_suggestions.md` §2) and standardize extract/recognize progress to one format (#12).
   - Delete the `os.Stderr` sink in `recognize.go:70`'s `io.MultiWriter`; report aggregate token counts instead.
   - Emit cycle/retry/max-retries logs via `feedback.Warnf` instead of raw `log.Printf`.
5. In `assemble.go` / `convert` command: emit skip reasons and counts via `feedback.Infof`.

## Out of scope (deliberately)

- No metrics/tracing services, no log files/rotation, no structured JSON logs on disk, no per-token streaming.
- No banner/art ascii output; progress is plain lines that scripts can grep.
- Minimum mode must not drop the existing `-> …` artifact lines that users already rely on — they become the #5/#7/#13/#15 summaries above.
- Keep `log/slog` defaults (no new dependencies); if a dedicated debug stream ever becomes necessary, wrap the flag name `--debug` in `internal/feedback` and leave the pipeline untouched.