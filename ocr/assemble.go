package ocr

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var pageNumRegex = regexp.MustCompile(`^page_(\d+)\.html$`)

// AssembleReport contains statistics about the assembly operation.
type AssembleReport struct {
	TotalPages  int
	Processed   int
	Skipped     int
	CombinedLen int
}

// Assemble discovers pageSources/page_*.html, keeps those in the given page set,
// sorts numerically, concatenates with PAGE markers, wraps in html/body, and writes combined.html.
//
// Contract: page sources must be in <OutDir>/pageSources/ directory. This is a documented
// coupling — the directory name is not configurable.
//
// Accepts Config and pages (page numbers to include).
// Returns AssembleReport with page counts and combined file length, or an error.
func Assemble(cfg Config, pages []int) (AssembleReport, error) {
	pageSources := filepath.Join(cfg.OutDir, "pageSources") // contract: "pageSources" directory name

	matches, err := filepath.Glob(filepath.Join(pageSources, "page_*.html"))
	if err != nil {
		return AssembleReport{}, fmt.Errorf("globbing page sources: %w", err)
	}

	allowed := make(map[int]bool, len(pages))
	for _, p := range pages {
		allowed[p] = true
	}

	// Map from page number to the actually-matched file path.
	selected := make(map[int]string)
	for _, m := range matches {
		stem := filepath.Base(m)
		if hit := pageNumRegex.FindStringSubmatch(stem); hit != nil {
			if p, err := strconv.Atoi(hit[1]); err == nil && allowed[p] {
				selected[p] = m
			}
		}
	}

	var pageNumbers []int
	for p := range selected {
		pageNumbers = append(pageNumbers, p)
	}
	sort.Ints(pageNumbers)

	report := AssembleReport{
		TotalPages: len(pageNumbers),
	}

	var combined strings.Builder
	for _, p := range pageNumbers {
		htmlPath := selected[p]
		htmlContent, err := os.ReadFile(htmlPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "assemble: warning - skipping page %d (cannot read %s: %v)\n", p, htmlPath, err)
			report.Skipped++
			continue
		}

		if !hasRecognizedBlocks(string(htmlContent)) {
			fmt.Fprintf(os.Stderr, "assemble: skipping page %d (no recognized blocks)\n", p)
			report.Skipped++
			continue
		}

		combined.WriteString(fmt.Sprintf("<!-- PAGE %d -->\n", p))
		combined.WriteString(string(htmlContent))
		combined.WriteString("\n")
		report.Processed++
	}

	finalHTML := "<html>\n  <body>\n" + combined.String() + "\n  </body>\n</html>\n"

	combinedPath := filepath.Join(cfg.OutDir, "combined.html")
	if err := os.WriteFile(combinedPath, []byte(finalHTML), 0644); err != nil {
		return report, fmt.Errorf("writing combined.html: %w", err)
	}

	report.CombinedLen = len(finalHTML)
	return report, nil
}

// hasRecognizedBlocks checks if an HTML string contains at least one div with both data-bbox and data-label attributes.
func hasRecognizedBlocks(htmlStr string) bool {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return false
	}

	var found bool
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && n.Data == "div" {
			var hasBbox, hasLabel bool
			for _, a := range n.Attr {
				if a.Key == "data-bbox" {
					hasBbox = true
				}
				if a.Key == "data-label" {
					hasLabel = true
				}
			}
			if hasBbox && hasLabel {
				found = true
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return found
}
