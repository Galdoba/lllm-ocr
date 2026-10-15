---
title: Issue Fix Tasks — lllm-ocr
tags:
  - project/tasks
  - project/issues
status: active
created: 2026-09-17
updated: 2026-09-17
---

# Issue Fix Tasks — lllm-ocr

> Concrete actionable tasks mapped to each known issue. Status: pending → in_progress → done.

---

## Phase 1 — Pure-Function & Localized Fixes

### Issue #4: `countRepeats` off-by-one
**Status:** done  
**File:** `ocr/cycle/detector.go:103-106`  
**Task list:**
- [x] Read `countRepeats` function — verify loop breaks at `i < patternLen` skipping first chunk
- [x] Change loop condition to iterate over all chunks including index 0
- [x] Add bounds check: ensure `len(tokens) >= patternLen` before accessing
- [x] Write regression test: short token list (e.g. 6 tokens, patternLen=3) that previously returned 0
- [x] Run `go test -race ./ocr/cycle/...`

### Issue #16: `GetCleanTokens` returns all when `cycleStart==0`
**Status:** done  
**File:** `ocr/cycle/detector.go:156-169`  
**Task list:**
- [x] Read `GetCleanTokens` — confirm it returns all tokens when `cycleStart == 0` even if cycling
- [x] Decide semantics: if `cycling==true` and `cycleStart==0`, return empty slice (all tokens are cycle)
- [x] Implement the fix
- [x] Add test: detector with full-buffer cycle → `GetCleanTokens` returns empty
- [x] Run `go test -race ./ocr/cycle/...`

### Issue #6: Fragile "killed" string match
**Status:** done  
**File:** `ocr/recognize.go:119`  
**Task list:**
- [x] Read the error handling block around line 119
- [x] Replace `strings.Contains(errStr, "killed")` with proper checks:
  - `errors.Is(err, os.ErrProcessDone)` for process-exit detection
  - Check exit status code (137 for SIGKILL) if err is `*exec.ExitError`
- [x] Update non-Unix fallback if needed
- [x] Run `go vet ./ocr/...`

### Issue #10: Unchecked Write in `fixHTMLNaivelyPath`
**Status:** done  
**File:** `ocr/cycle/htmlfix.go:86`  
**Task list:**
- [x] Read `fixHTMLNaivelyPath` — confirm `f.Write()` and `f.WriteString()` return values ignored
- [x] Check each `Write`/`WriteString` return value
- [x] Return wrapped error if any write fails
- [x] Add test: mock file system where write fails → verify error propagated
- [x] Run `go test ./ocr/cycle/...`

### Issue #28: `Parse()` swallows `os.ReadFile` errors
**Status:** done  
**File:** `layoutblock/parse.go:35-38`  
**Task list:**
- [x] Read `Parse` — confirm it checks for `<` char and falls through to `os.ReadFile(path)` on no match
- [x] Check the `os.ReadFile` error; return it instead of treating path string as HTML
- [x] Add test: nonexistent file path → expect error, not empty blocks
- [x] Run `go test ./layoutblock/...`

### Issue #29: Negative `MaxRetries` skips all attempts
**Status:** done  
**File:** `ocr/recognize.go:44`  
**Task list:**
- [x] Read `CanRetry()` — `CurrentRetry < MaxRetries` returns false when MaxRetries <= 0
- [x] Clamp negative values to 0 in `RecognizeSinglePage` (no retries, but initial attempt runs)
- [x] Add test: negative MaxRetries → clamp to 0, initial attempt runs
- [x] Run `go test -race ./ocr/...`

### Issue #15: `maxTokens==0` silent disable
**Status:** done  
**File:** `ocr/cycle/detector.go:41`  
**Task list:**
- [x] Read `AddToken` — confirm `maxTokens==0` silently prevents token storage
- [x] Add a one-time log warning when detection is disabled due to zero config
- [x] Add test: zero maxTokens → verify log warning appears
- [x] Run `go test ./ocr/cycle/...`

### Issue #12: Progress reporting inconsistency
**Status:** done  
**File:** `ocr/extract.go:24-25`  
**Task list:**
- [x] Read `ExtractPages` — confirm it reports `i+1` (iteration index) instead of actual page number `p`
- [x] Change progress to report `p` (actual page number) like `RecognizePages` does
- [x] Run `go test ./ocr/...`

### Issue #19: Closure capture (Go 1.22+ safe)
**Status:** done  
**File:** `ocr/recognize.go:18`  
**Task list:**
- [x] Verify Go version >= 1.22 (per-iteration loop variables)
- [x] Add a comment noting Go 1.22+ safety; no code change needed
- [x] Run `go vet ./ocr/...`

---

## Phase 2 — Config & CLI Wiring

### Issue #20: Unvalidated CycleConfig → divide-by-zero
**Status:** done  
**Files:** `validate.go`, `commands/ocr.go:91`, `cycle/detector.go:47`  
**Task list:**
- [x] Read `ValidatePipeline` / `ValidatePipelineConfig` — confirm they exist but are never called
- [x] Call `ValidatePipeline` in `commands/ocr.go` before creating pipeline
- [x] Add defensive guard in `Detector.AddToken`: if `checkInterval <= 0`, return without cycling (log warning)
- [x] Add test: zero CheckInterval → expect no panic, warning logged
- [x] Run `go test -race ./...`

### Issue #21: Lenient flag inverted
**Status:** done  
**Files:** `layoutblock/parse.go:89-116`, `commands/convert.go:83-96`  
**Task list:**
- [x] Read `parse` and `traverse` functions — confirm lenient mode returns error on first unsupported label
- [x] Fix lenient mode: collect unsupported labels into errs slice, continue processing (don't return early)
- [x] Fix strict mode: return immediately on first error (current behavior)
- [x] Update `convert.go`: if lenient mode, use collected blocks + errors; if strict, fail on error
- [x] Add test: HTML with mixed supported/unsupported labels, lenient=true → partial blocks returned
- [x] Run `go test ./layoutblock/...`

### Issue #8 + #26: Assemble regex + path mismatch
**Status:** done  
**File:** `ocr/assemble.go:30,43-59`  
**Task list:**
- [x] Change `pageNumRegex` from `(\d+)$` to `^page_(\d+)\.html$`
- [x] Fix path derivation: use the actually-matched filename, not re-derived zero-padded path
- [x] Treat unmatched files as skips (log warning), not run-aborting errors
- [x] Add test: file `page_42.html` in pageSources → correctly read that file
- [x] Add test: file `backup_001.html` → correctly skipped (doesn't match regex)
- [x] Run `go test ./ocr/...`

### Issue #25: OCR command swallows errors
**Status:** done  
**File:** `internal/commands/ocr.go:94-96`  
**Task list:**
- [x] Read error handling around `RecognizePages` call
- [x] Print error to stderr (not stdout)
- [x] Return the error so CLI exits non-zero
- [x] Keep pageSources move only on success (current: happens regardless)
- [x] Run `go test ./internal/commands/...`

### Issue #17: Probe glob cleanup passes literal string
**Status:** done  
**File:** `ocr/extract.go:62`  
**Task list:**
- [x] Read `os.RemoveAll(probePattern)` — confirm it passes glob pattern literally
- [x] Replace with proper glob expansion + file-by-file removal
- [x] Add test: verify probe files are actually cleaned up
- [x] Run `go test ./ocr/...`

### Issue #23: Retry budget off-by-one
**Status:** done  
**Files:** `ocr/recognize.go:44`, `recognize_retry.go:127-130`  
**Task list:**
- [x] Read `CanRetry`: `CurrentRetry < MaxRetries` gives exactly MaxRetries total attempts
- [x] Change `CanRetry` to `CurrentRetry <= MaxRetries` (MaxRetries=3 → 4 total: 1 initial + 3 retries)
- [x] Align log messages with new semantics
- [x] Add test: MaxRetries=2 → verify 3 total attempts (1 initial + 2 retries)
- [x] Run `go test -race ./ocr/...`

### Issue #18: Hardcoded "pageSources" directory
**Status:** done  
**File:** `ocr/assemble.go:30`  
**Task list:**
- [x] Read where "pageSources" is used
- [x] Add a doc comment documenting the hardcoded coupling as a contract
- [x] Run `go vet ./ocr/...`

---

## Phase 3 — Dead Code & Consolidation

### Issue #14 + #27: Dead `StartRetry`, duplicate temp-arg logic
**Status:** done  
**File:** `ocr/recognize_retry.go:25-117`  
**Task list:**
- [x] Read `StartRetry` — confirm it's exported but never called
- [x] Verify `recognize.go` has its own inline retry loop with duplicate temp-arg logic
- [x] Delete `StartRetry` and all its helper functions
- [x] Keep `getArgsForAttempt` (or equivalent) as the single source of truth
- [x] Move temp-arg logic into a shared helper if not already
- [x] Add unit test for temp-arg replacement function
- [x] Run `go test -race ./ocr/...`

### Issue #11: Dead `detector` param in `CompleteHTMLWithContent`
**Status:** done  
**Files:** `ocr/cycle/htmlfix.go:12`, call sites in `recognize.go:129,145`  
**Task list:**
- [x] Read `CompleteHTMLWithContent` — confirm `detector` parameter unused
- [x] Remove `*cycle.Detector` from function signature
- [x] Update all call sites to drop the detector argument
- [x] Update any tests that pass a detector
- [x] Run `go test ./ocr/cycle/...`

---

## Phase 4 — Concurrency & Reliability Rework

> **WARNING:** High-risk phase. All changes touch the same goroutine lifecycle. Test under `-race` after each sub-task.

### Issue #30 + #3: Detector async spawn / flaky test
**Status:** done  
**Files:** `ocr/cycle/detector.go:47-51`, `detector_test.go:141-166`  
**Task list:**
- [x] Read `AddToken` — confirm `go d.detectCycling()` spawns goroutine per CheckInterval
- [x] Replace goroutine-per-check with synchronous detection in `AddToken` (lock held, fast path)
- [x] For async: use a single worker goroutine with a channel, not one per check
- [x] Expose `Detect() bool` for synchronous detection (used by tests)
- [x] Fix `TestDetector_GetCleanTokens`: use `Detect()` + `sync.WaitGroup` instead of sleep
- [x] Verify O(n²) churn is eliminated (no deep-copy per check)
- [x] Run `go test -race -count=10 ./ocr/cycle/...` to verify no flakiness

### Issue #5: Kill error handling
**Status:** done  
**Files:** `ocr/recognize.go:93`, `recognize_retry.go:96`  
**Task list:**
- [x] Read `syscall.Kill` error handling — confirm errors are discarded with `_ =`
- [x] Check the error; if non-nil, log warning with process PID
- [x] If process survived (not killed), consider a second kill signal or mark page as failed
- [x] Run `go test -race ./ocr/...`

### Issue #2: Missing `recover()` on goroutines
**Status:** done  
**Files:** `ocr/recognize.go:82-99`, `recognize_retry.go:88-102`  
**Task list:**
- [x] Add `defer recover()` at the start of each stdout-reading goroutine
- [x] On panic, log the panic value and signal failure via channel (set `cycleDetected <- false` or use error channel)
- [x] Ensure `cmd.Wait()` is never blocked on a dead goroutine
- [x] Run `go test -race ./ocr/...`

### Issue #1 + #24: Goroutine/close race + dropped cycle signal
**Status:** done  
**Files:** `ocr/recognize.go:70,104-113`, `recognize_retry.go:59,82-109`  
**Task list:**
- [x] Replace non-blocking `select` with blocking `<-cycleDetected` receive
- [x] Add `sync.WaitGroup` to wait for goroutine before closing `htmlFile`
- [x] Ensure goroutine finishes (HTML writes complete) before file close
- [x] Verify cycle signal is never lost (blocking receive guarantees delivery)
- [x] Run `go test -race -count=20 ./ocr/...` under stress

### Issue #7: HTML completion on non-cycle errors
**Status:** done  
**File:** `ocr/recognize.go:115-129`  
**Task list:**
- [x] Read HTML completion logic — confirm it completes HTML on any error path
- [x] Add a marker/flag to distinguish "intentional completion" from "crash recovery"
- [x] Only complete HTML when page is being kept (not when OOM/crash)
- [x] Use a marker file or header flag that post-process can detect as "partial"
- [x] Run `go test -race ./ocr/...`

---

## Cross-Cutting Verification

After each phase completes:
- [ ] `go vet ./...`
- [ ] `go test -race ./...`
- [ ] If applicable: run full pipeline on `test_books/` PDFs
- [ ] Update `known_issues.md` to mark resolved issues
- [ ] Append report to `issue_report.md`
