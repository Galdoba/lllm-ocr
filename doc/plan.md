---
title: Issue Fix Plan — lllm-ocr
tags:
  - project/plan
  - project/issues
status: active
created: 2026-09-17
updated: 2026-09-17
---

# Issue Fix Plan — lllm-ocr

> Phased plan to resolve all 30 known issues. Phases are ordered by risk and dependency.

---

## Execution Rules

1. Work phases in order: Phase 1 → 2 → 3 → 4
2. Each issue gets a task entry in `tasks.md`; mark completed only when code + tests pass
3. After completing an issue, append a report to `issue_report.md`
4. Run `go vet ./...` and `-race` builds after every phase
5. If a fix reveals a new issue, add it to `known_issues.md` before proceeding

---

## Phase 1 — Pure-Function & Localized Fixes

> Single-file edits, no contract changes. Low risk, independently testable.

| # | Issue | Files | Effort |
|---|-------|-------|--------|
| 4 | `countRepeats` off-by-one | `ocr/cycle/detector.go` | small |
| 16 | `GetCleanTokens` returns all when cycling=true | `ocr/cycle/detector.go` | small |
| 6 | Fragile "killed" string match | `ocr/recognize.go` | small |
| 10 | Unchecked Write in `fixHTMLNaivelyPath` | `ocr/cycle/htmlfix.go` | small |
| 28 | `Parse()` swallows os.ReadFile errors | `layoutblock/parse.go` | small |
| 29 | Negative `MaxRetries` skips all attempts | `ocr/recognize.go` | small |
| 15 | `maxTokens==0` silent disable | `ocr/cycle/detector.go` | small |
| 12 | Progress reporting inconsistency | `ocr/extract.go` | small |
| 19 | Closure capture (Go 1.22+ safe) | `ocr/recognize.go` | cosmetic |

---

## Phase 2 — Config & CLI Wiring

> Medium effort, no concurrency changes. Sets up guards before Phase 4.

| # | Issue | Files | Effort |
|---|-------|-------|--------|
| 20 | Unvalidated CycleConfig → div-by-zero | `validate.go`, `commands/ocr.go`, `cycle/detector.go` | medium |
| 21 | Lenient flag inverted | `layoutblock/parse.go`, `commands/convert.go` | medium |
| 8+26 | Assemble regex + path mismatch | `ocr/assemble.go` | medium |
| 25 | OCR command swallows errors | `internal/commands/ocr.go` | small |
| 17 | Probe glob cleanup passes literal string | `ocr/extract.go` | small |
| 23 | Retry budget off-by-one | `ocr/recognize.go`, `recognize_retry.go` | small |
| 18 | Hardcoded "pageSources" directory | `ocr/assemble.go` | cosmetic |

---

## Phase 3 — Dead Code & Consolidation

> Removes divergence before the big concurrency rewrite.

| # | Issue | Files | Effort |
|---|-------|-------|--------|
| 14+27 | Dead `StartRetry`, duplicate temp-arg logic | `recognize_retry.go` | medium |
| 11 | Dead `detector` param in `CompleteHTMLWithContent` | `cycle/htmlfix.go`, call sites | small |

---

## Phase 4 — Concurrency & Reliability Rework

> Most invasive. All changes touch the same goroutine lifecycle in `recognize.go` / `recognize_retry.go`. Verify under `-race`.

| # | Issue | Files | Effort |
|---|-------|-------|--------|
| 30+3 | Detector async spawn / flaky test | `cycle/detector.go`, `detector_test.go` | large |
| 5 | Kill error handling | `recognize.go`, `recognize_retry.go` | small |
| 2 | Missing `recover()` on goroutines | `recognize.go`, `recognize_retry.go` | small |
| 1+24 | Goroutine/close race + dropped cycle signal | `recognize.go`, `recognize_retry.go` | large |
| 7 | HTML completion on non-cycle errors | `recognize.go` | medium |

---

## Dependencies

- Phase 2 (#20) must land before Phase 4 (#30) — both touch `Detector.AddToken`
- Phase 4 (#7) depends on (#1, #2, #24) — lifecycle rework first
- Phase 4 (#3) unlocks (#30) — fix test then fix detector

## Verification Gates

- After each phase: `go vet ./...` + `go test -race ./...`
- Phase 4 additionally: run full OCR pipeline on test books under `-race`
