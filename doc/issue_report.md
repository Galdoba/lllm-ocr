---
title: Issue Fix Reports — lllm-ocr
tags:
  - project/reports
  - project/issues
status: active
created: 2026-09-17
updated: 2026-09-17
---

# Issue Fix Reports — lllm-ocr

> Reports appended after each issue is resolved. Format: issue number, date, summary, verification.

---

## Reports

<!-- Append reports in reverse chronological order (newest first). -->

<!--
### Issue #XX: <title>
**Date:** YYYY-MM-DD  
**Phase:** N  
**Summary:** What was changed  
**Files:** list of files modified  
**Tests:** tests added/updated  
**Verification:** `go test -race ./...` passed / failed  
**Notes:** Any additional context, follow-ups needed
-->

### Issue #4: `countRepeats` off-by-one
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Removed `if i < patternLen { break }` guard in `countRepeats` that skipped the first chunk when `i == 0`. The loop now evaluates all chunks including index 0.  
**Files:** `ocr/cycle/detector.go:103-105` (removed 3-line guard)  
**Tests:** Added `TestCountRepeats_FirstChunkEvaluated` and `TestCountRepeats_FirstChunkNonMatching`; fixed `TestDetectCycling_BelowMinTokens` expectation (49→50); fixed `TestDetectCycling_MaxPatternLen` expectation  
**Verification:** `go test -race ./ocr/cycle/...` passed; `go vet ./...` clean  
**Notes:** The fix exposed that `TestDetectCycling_MaxPatternLen` was relying on the bug to pass. Updated test to document correct behavior.

### Issue #3: `TestDetector_GetCleanTokens` flaky test
**Date:** 2026-09-17  
**Phase:** 4 (partial)  
**Summary:** Added synchronous `Detect()` method to `Detector` so tests don't depend on async goroutine timing. Updated `TestDetector_GetCleanTokens` to use `Detect()` instead of `time.Sleep(300ms)`.  
**Files:** `ocr/cycle/detector.go` (added `Detect()` method); `ocr/cycle/detector_test.go` (updated test)  
**Tests:** Verified deterministic with `-count=5`  
**Verification:** `go test -race -count=5 ./ocr/cycle/...` passed 5/5 times  
**Notes:** Issue #30 (O(n²) goroutine churn) is still pending. The `Detect()` method is a stepping stone that also helps tests.

### Issue #16: `GetCleanTokens` returns all when `cycleStart==0`
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Fixed `GetCleanTokens` condition from `!d.cycling || d.cycleStart == 0` to `!d.cycling` only. When cycling=true and cycleStart=0, all tokens are part of the cycle, so return empty slice.  
**Files:** `ocr/cycle/detector.go:160`  
**Tests:** Added `TestGetCleanTokens_AllTokensAreCycle`  
**Verification:** `go test -race ./ocr/cycle/...` passed; `go test -race ./...` passed  
**Notes:** None.

### Issue #6: Fragile "killed" string match
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Replaced `strings.Contains(errStr, "killed")` with `isProcessKilledError()` that checks `os.ErrProcessDone` and exit code 137/SIGKILL via `*exec.ExitError`.  
**Files:** `ocr/recognize.go:119` (replaced fragile string matching); added `isProcessKilledError()` helper  
**Tests:** No new tests (behavior verified by existing tests)  
**Verification:** `go build ./...` passed; `go vet ./...` passed; `go test -race ./...` passed  
**Notes:** Removed unused `strings` import.

### Issue #10: Unchecked Write in `fixHTMLNaivelyPath`
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Added return value checks for `f.Write()` and `f.WriteString()` in `fixHTMLNaivelyPath`. Returns wrapped errors on write failure.  
**Files:** `ocr/cycle/htmlfix.go:86-87`  
**Tests:** Existing tests still pass; no regression  
**Verification:** `go test -race ./...` passed  
**Notes:** None.

### Issue #28: `Parse()` swallows `os.ReadFile` errors
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Fixed `Parse()` to return an error when `os.ReadFile` fails on a file path input, instead of silently treating the path string as HTML.  
**Files:** `layoutblock/parse.go:35-39`  
**Tests:** Existing tests pass; behavior change is backward-compatible (previously silent failure now returns error)  
**Verification:** `go test -race ./...` passed  
**Notes:** None.

### Issue #29: Negative `MaxRetries` skips all attempts
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Clamped negative `MaxRetries` to 0 in `RecognizeSinglePage` so the initial attempt runs (no retries, but no silent skip).  
**Files:** `ocr/recognize.go:42`  
**Tests:** Existing tests pass; behavior change prevents silent no-op on negative config  
**Verification:** `go test -race ./...` passed  
**Notes:** None.

### Issue #15: `maxTokens==0` silent disable
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Added one-time log warning when `maxTokens==0` disables cycle detection. Previously silent.  
**Files:** `ocr/cycle/detector.go:41-43` (added sync.Once warning before early return)  
**Tests:** Existing tests pass; warning fires once per program run  
**Verification:** `go test -race ./...` passed  
**Notes:** Uses package-level `sync.Once` to ensure warning fires exactly once.

### Issue #12: Progress reporting inconsistency
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Changed `ExtractPages` to report actual page number (`p`) in progress callback, matching `RecognizePages` behavior. Previously reported iteration index (`i+1`).  
**Files:** `ocr/extract.go:25`  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Removed unused `i` variable from loop.

### Issue #19: Closure capture (Go 1.22+ safe)
**Date:** 2026-09-17  
**Phase:** 1  
**Summary:** Added comment documenting Go 1.22+ per-iteration loop variable safety. No code change needed.  
**Files:** `ocr/recognize.go:18`  
**Tests:** N/A (cosmetic change)  
**Verification:** `go vet ./...` passed  
**Notes:** Go 1.25.1 is well above 1.22, so per-iteration loop variables are guaranteed.

### Issue #20: Unvalidated CycleConfig → divide-by-zero
**Date:** 2026-09-17  
**Phase:** 2  
**Summary:** Added validation call in OCR command before pipeline execution; added defensive guard in `Detector.AddToken` for `checkInterval <= 0`.  
**Files:** `internal/commands/ocr.go:91` (added `ValidatePipeline` call); `ocr/cycle/detector.go:57-61` (added checkInterval guard with one-time warning)  
**Tests:** Existing tests pass; guard prevents panic on bad config  
**Verification:** `go test -race ./...` passed  
**Notes:** Two-layer defense: CLI validation (user-friendly error) + runtime guard (prevents panic).

### Issue #21: Lenient flag inverted
**Date:** 2026-09-17  
**Phase:** 2  
**Summary:** Fixed lenient mode in `parse()`: now collects errors and continues instead of returning immediately. Strict mode still returns on first error.  
**Files:** `layoutblock/parse.go:57-153` (inverted lenient logic: `!lenient` for early return, lenient collects into `errs`)  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** The `convert.go` error handling at line 83-86 already returns on error, so lenient mode now returns partial blocks + error, while strict mode fails fast.

### Issue #8 + #26: Assemble regex + path mismatch
**Date:** 2026-09-17  
**Phase:** 2  
**Summary:** Tightened `pageNumRegex` to `^page_(\d+)\.html$`; changed path derivation to use the actually-matched filename instead of re-derived zero-padded path; treat read errors as skips, not failures.  
**Files:** `ocr/assemble.go:15` (regex), `ocr/assemble.go:37-57` (path mapping), `ocr/assemble.go:64-71` (error handling)  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** `backup_001.html` now correctly skipped (doesn't match `^page_\d+\.html$`); `page_42.html` (non-zero-padded) correctly read.

### Issue #25: OCR command swallows errors
**Date:** 2026-09-17  
**Phase:** 2  
**Summary:** Changed error handling in OCR command: print to stderr and return error (nonzero exit) instead of silently continuing.  
**Files:** `internal/commands/ocr.go:98-100`  
**Tests:** No unit tests in commands package  
**Verification:** `go build ./...` passed; `go vet ./...` passed  
**Notes:** Failed recognition now exits non-zero and skips pageSources move.

### Issue #17: Probe glob cleanup passes literal string
**Date:** 2026-09-17  
**Phase:** 2  
**Summary:** Fixed probe file cleanup to use `filepath.Glob` expansion + `os.Remove` per file instead of passing glob pattern literally to `os.RemoveAll`.  
**Files:** `ocr/extract.go:61-66`  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Previously `os.RemoveAll(".../probe_page_1-*.png")` passed literal glob string; now expands and removes each file.

### Issue #23: Retry budget off-by-one
**Date:** 2026-09-17  
**Phase:** 2  
**Summary:** Changed `CanRetry` from `CurrentRetry < MaxRetries` to `CurrentRetry <= MaxRetries` so MaxRetries=N means N retries after first attempt (N+1 total attempts).  
**Files:** `ocr/recognize_retry.go:128-130`  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Log message "max retries (N) exceeded" is now accurate: N retries after first attempt.

### Issue #18: Hardcoded "pageSources" directory
**Date:** 2026-09-17  
**Phase:** 2  
**Summary:** Added doc comment documenting "pageSources" as a hardcoded contract. No code change.  
**Files:** `ocr/assemble.go:25-32`  
**Tests:** N/A  
**Verification:** `go vet ./...` passed  
**Notes:** Directory name is not configurable per user decision.

### Issue #14 + #27: Dead `StartRetry`, duplicate temp-arg logic
**Date:** 2026-09-17  
**Phase:** 3  
**Summary:** Deleted dead `StartRetry` function and `RetryResult` type (93 lines). Only `RetryState` and helpers remain in `recognize_retry.go`.  
**Files:** `ocr/recognize_retry.go` (removed 117 lines, kept 30 lines)  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** `getArgsForAttempt` in `recognize.go` is now the single source of temp-arg logic.

### Issue #11: Dead `detector` param in `CompleteHTMLWithContent`
**Date:** 2026-09-17  
**Phase:** 3  
**Summary:** Removed unused `*cycle.Detector` parameter from `CompleteHTMLWithContent` signature and all call sites.  
**Files:** `ocr/cycle/htmlfix.go:12` (removed parameter); `ocr/recognize.go:129,145` (updated calls); `ocr/cycle/detector_test.go` (updated calls)  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Also removed unused `detector := cycle.NewDetector(...)` from recognize.go line 144.

### Issue #30 + #3: Detector async spawn / flaky test
**Date:** 2026-09-17  
**Phase:** 4  
**Summary:** Replaced `go d.detectCycling()` with synchronous `d.detectCycling()` in `AddToken`. Fixed deadlock (AddToken held write lock, detectCycling needed RLock). Updated `TestDetector_DetectionTimeout` to measure sync performance.  
**Files:** `ocr/cycle/detector.go:42-75` (manual lock/unlock around sync detect); `ocr/cycle/detector_test.go:391-404` (updated test)  
**Tests:** All cycle tests pass; O(n²) goroutine churn eliminated  
**Verification:** `go test -race -count=3 ./ocr/cycle/...` passed; `go test -race ./...` passed  
**Notes:** Detection is now synchronous and fast. Lock is released during detection to avoid deadlock.

### Issue #5: Kill error handling
**Date:** 2026-09-17  
**Phase:** 4  
**Summary:** Added error check for `syscall.Kill` in cycle detection goroutine. Logs warning if kill fails.  
**Files:** `ocr/recognize.go:98`  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Process group kill now surfaces errors instead of silently ignoring them.

### Issue #2: Missing `recover()` on goroutines
**Date:** 2026-09-17  
**Phase:** 4  
**Summary:** Added `defer recover()` to stdout-reading goroutine in `RecognizeSinglePage`. On panic, logs and signals failure via channel.  
**Files:** `ocr/recognize.go:87-93`  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Prevents deadlock if goroutine panics — main goroutine still receives signal on `cycleDetected` channel.

### Issue #1 + #24: Goroutine/close race + dropped cycle signal
**Date:** 2026-09-17  
**Phase:** 4  
**Summary:** Added `sync.WaitGroup` to wait for goroutine before closing `htmlFile`. Replaced non-blocking select with blocking receive. Cycle signal now never lost.  
**Files:** `ocr/recognize.go:78-127`  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Key changes: (1) `wg.Add(1)` before goroutine, `wg.Done()` in goroutine; (2) `wg.Wait()` before `htmlFile.Close()`; (3) non-blocking sends in goroutine to avoid blocking; (4) blocking receive after `wg.Wait()`.

### Issue #7: HTML completion on non-cycle errors
**Date:** 2026-09-17  
**Phase:** 4  
**Summary:** Added `crashRecovery` parameter to `CompleteHTMLWithContent`. When true, prepends `<!-- OCR crash recovery: ... -->` marker so post-processing can detect incomplete files.  
**Files:** `ocr/cycle/htmlfix.go:13,28-30` (added crashRecovery param + marker); `ocr/recognize.go:155,169` (pass true for crash recovery)  
**Tests:** Existing tests pass  
**Verification:** `go test -race ./...` passed  
**Notes:** Crash-recovered HTML files now have a detectable marker comment at the top.
