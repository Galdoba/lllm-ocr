package ocr

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ParsePageSpec parses a comma-separated list of single pages and inclusive
// ranges (e.g. "1-3,5,6-8") into a sorted, de-duplicated slice of page numbers.
// Ranges are inclusive on both ends ("12-14" -> 12, 13, 14).
// Returns an error for malformed input: empty string, empty range components,
// non-numeric tokens, or reversed ranges ("5-2").
func ParsePageSpec(spec string) ([]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("empty page specification")
	}

	seen := make(map[int]bool)
	var pages []int
	for _, item := range strings.Split(spec, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("empty page in specification: %q", spec)
		}

		lo, hi, err := parsePageItem(item)
		if err != nil {
			return nil, err
		}
		if lo < 1 {
			return nil, fmt.Errorf("invalid page %d: pages are 1-indexed", lo)
		}
		for p := lo; p <= hi; p++ {
			if !seen[p] {
				seen[p] = true
				pages = append(pages, p)
			}
		}
	}

	sort.Ints(pages)
	return pages, nil
}

// parsePageItem parses one comma-separated item into (lo, hi). A lone integer
// maps to (n, n); "start-end" maps to (start, end) inclusive.
func parsePageItem(item string) (int, int, error) {
	if dash := strings.IndexByte(item, '-'); dash >= 0 {
		start := strings.TrimSpace(item[:dash])
		end := strings.TrimSpace(item[dash+1:])
		lo, err := strconv.Atoi(start)
		if err != nil || start == "" {
			return 0, 0, fmt.Errorf("invalid page range %q: page numbers must be integers", item)
		}
		hi, err := strconv.Atoi(end)
		if err != nil || end == "" {
			return 0, 0, fmt.Errorf("invalid page range %q: page numbers must be integers", item)
		}
		if lo > hi {
			return 0, 0, fmt.Errorf("invalid page range %q: start page cannot be greater than end page", item)
		}
		return lo, hi, nil
	}

	n, err := strconv.Atoi(item)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page %q: must be an integer or a start-end range", item)
	}
	return n, n, nil
}

// SelectPages parses spec and keeps only pages up to maxPage. Pages above
// maxPage are dropped. A maxPage <= 0 applies no upper bound.
func SelectPages(spec string, maxPage int) ([]int, error) {
	pages, err := ParsePageSpec(spec)
	if err != nil {
		return nil, err
	}
	if maxPage <= 0 {
		return pages, nil
	}

	kept := make([]int, 0, len(pages))
	for _, p := range pages {
		if p <= maxPage {
			kept = append(kept, p)
		} else {
			break
		}
	}
	return kept, nil
}

var pdfInfoPagesRegex = regexp.MustCompile(`(?im)^Pages:\s*(\d+)`)

// PdfPageCount runs `pdfinfo` on pdfPath and returns the number of pages.
func PdfPageCount(pdfPath string) (int, error) {
	out, err := exec.Command("pdfinfo", pdfPath).Output()
	if err != nil {
		return 0, fmt.Errorf("querying PDF page count: %w", err)
	}

	line := pdfInfoPagesRegex.FindSubmatch(out)
	if line == nil {
		return 0, fmt.Errorf("could not find page count in pdfinfo output for %s", pdfPath)
	}

	n, err := strconv.Atoi(string(line[1]))
	if err != nil {
		return 0, fmt.Errorf("parsing pdfinfo page count: %w", err)
	}
	return n, nil
}
