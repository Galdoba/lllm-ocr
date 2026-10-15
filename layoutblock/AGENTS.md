# layoutblock — DOX

## Purpose

Core layout-block domain: the `LayoutBlock`/`BBox`/`LabelType` types plus the parse → filter → merge transformation that turns page HTML into a clean block list. This is the foundational package; `ocr/` and `output/` both import it.

## Ownership

Owns: `BBox`, `LabelType`, `LayoutBlock`, `BlockParser`/`ParserConfig`, `FilterBlocks`/`FilterConfig`, `MergeCrossPageBlocks`/`MergeConfig`, `ExtractText`/`ExtractTextFromNode`/`FindBody`, `VisualizeBlocks`, `NormalizeBlockLineBreaks`, `SortByPosition`, `BlocksIntersecting`, `BlocksContainingPoint`, `GroupByVerticalPosition`. Config types (`FilterConfig`, `MergeConfig`, `ParserConfig`) are type aliases to `internal/config/`.

Depends on: `internal/config/`. Nothing else internal. Imported by `ocr/` (convert stage) and `output/`.

## Local Contracts

- Coordinates are normalized integers in `[0, 1000]` (`BBox`).
- Adding a `LabelType` requires TWO changes: the string constant AND an entry in the `supportedLabels` map (`layoutblock.go`).
- Page-number artifact thresholds live in `FilterConfig.MinArtifactHeight` (default 50) and `FilterConfig.MinArtifactY` (default 900) — never hardcode magic numbers in `FilterBlocks`.
- Cross-page text merge look-ahead is bounded by `MergeConfig.MaxLookAhead` (default 9) via `MergeCrossPageBlocksWithConfig`.
- Cross-page merge only handles `LabelText` and `LabelTOC`. Table/text merges for other labels are intentionally skipped (see `doc/known_issues.md` #2) — do not expand this without an explicit request.

## Work Guidance

- Keep pure functions pure (no filesystem/`os`/`log`) except `VisualizeBlocks` (debug helper in `debug.go`).
- Preserve `Merged` flag semantics on merged blocks.
- `Parse` distinguishes HTML string vs file path by the presence of `<`.

## Verification

```
go test ./layoutblock/...
go build ./...
```

## Child DOX Index

None.
