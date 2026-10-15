package output

import (
	"regexp"
	"strings"
)

// cleanMarkdown applies post-conversion cleanup to fix common markdown artifacts
// produced while converting OCR'd HTML: redundant list markers, empty table
// headers, and split tables that share an identical header.
func cleanMarkdown(md string) string {
	md = cleanListMarkers(md)
	md = cleanTables(md)
	return md
}

// ---- redundant list markers ----

// unorderedMarkerDupRe matches a bullet (U+2022) that is used as a redundant
// marker immediately after a markdown list marker, e.g. "- • item".
var unorderedMarkerDupRe = regexp.MustCompile(`(?m)^([ \t]*[-*+][ \t]*)\x{2022}[ \t]*`)

// orderedMarkerDupRe matches a redundant "N\. " prefix that duplicates the
// ordered-list number, e.g. "1. 1\. item".
var orderedMarkerDupRe = regexp.MustCompile(`(?m)^([ \t]*(\d+)\.[ \t]+)(\d+)\\?\.[ \t]+`)

func cleanListMarkers(md string) string {
	md = unorderedMarkerDupRe.ReplaceAllString(md, "${1}")

	return orderedMarkerDupRe.ReplaceAllStringFunc(md, func(match string) string {
		sub := orderedMarkerDupRe.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		// Only strip the duplicated number when it matches the list's own number,
		// to avoid altering legitimately numbered content.
		if sub[2] == sub[3] {
			return sub[1]
		}
		return match
	})
}

// ---- table cleanup ----

var blockSplitRe = regexp.MustCompile(`\n[ \t]*\n`)

// tableSeparatorRe matches a markdown table separator row such as
// "| --- | --- |" or "-----|-----".
var tableSeparatorRe = regexp.MustCompile(`^[ \t]*\|?[ \t]*:?-{2,}:?[ \t]*(\|[ \t]*:?-{2,}:?[ \t]*)*\|?[ \t]*$`)

func splitBlocks(md string) []string {
	return blockSplitRe.Split(md, -1)
}

// tableSeparatorIndex returns the index of the first separator row in a block,
// and true if the block looks like a markdown table.
func tableSeparatorIndex(lines []string) (int, bool) {
	for i, line := range lines {
		if tableSeparatorRe.MatchString(line) {
			return i, true
		}
	}
	return -1, false
}

// isEmptyHeaderLine reports whether every pipe-delimited cell in the line is empty.
func isEmptyHeaderLine(line string) bool {
	for _, cell := range strings.Split(line, "|") {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// tableHeader returns the trimmed header line (the row above the first separator)
// and true if the block is a markdown table with a header.
func tableHeader(block string) (string, bool) {
	lines := strings.Split(block, "\n")
	sepIdx, ok := tableSeparatorIndex(lines)
	if !ok || sepIdx == 0 {
		return "", false
	}
	return strings.TrimSpace(lines[sepIdx-1]), true
}

func cleanTables(md string) string {
	blocks := splitBlocks(md)

	for i, b := range blocks {
		lines := strings.Split(b, "\n")
		sepIdx, ok := tableSeparatorIndex(lines)
		if !ok || sepIdx != 1 {
			continue
		}
		// An entirely empty header row means the real header was mis-parsed as the
		// first data row: promote that row to the header, keeping the separator below it.
		if isEmptyHeaderLine(lines[0]) && len(lines) >= 3 {
			promoted := make([]string, 0, len(lines))
			promoted = append(promoted, lines[sepIdx+1])
			promoted = append(promoted, lines[sepIdx])
			promoted = append(promoted, lines[sepIdx+2:]...)
			blocks[i] = strings.Join(promoted, "\n")
		}
	}

	return strings.Join(mergeIdenticalHeaderTables(blocks), "\n\n")
}

// mergeIdenticalHeaderTables concatenates two adjacent tables that share the same
// header row, dropping the second table's header and separator so they become one.
func mergeIdenticalHeaderTables(blocks []string) []string {
	var result []string
	for _, b := range blocks {
		if merged, ok := tryMergeIntoLast(result, b); ok {
			result[len(result)-1] = merged
			continue
		}
		result = append(result, b)
	}
	return result
}

// tryMergeIntoLast reports whether block b can be appended to the last block in
// out (they must be tables with identical, non-empty headers). On success it
// returns the merged block and true.
func tryMergeIntoLast(out []string, b string) (string, bool) {
	if len(out) == 0 {
		return "", false
	}
	headerLast, okA := tableHeader(out[len(out)-1])
	headerB, okB := tableHeader(b)
	if !okA || !okB || headerLast == "" || headerLast != headerB {
		return "", false
	}

	linesA := strings.Split(out[len(out)-1], "\n")
	var dataB []string
	sepSeen := false
	for _, line := range strings.Split(b, "\n") {
		if !sepSeen {
			if tableSeparatorRe.MatchString(line) {
				sepSeen = true
			}
			continue
		}
		dataB = append(dataB, line)
	}

	merged := make([]string, 0, len(linesA)+len(dataB))
	merged = append(merged, linesA...)
	merged = append(merged, dataB...)
	return strings.Join(merged, "\n"), true
}
