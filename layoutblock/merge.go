package layoutblock

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"

	"github.com/Galdoba/lllm-ocr/internal/config"
	"golang.org/x/net/html"
)

// MergeConfig is an alias for config.MergeConfig for backward compatibility.
type MergeConfig = config.MergeConfig

// DefaultMergeConfig returns a MergeConfig with default look-ahead limit.
func DefaultMergeConfig() MergeConfig {
	return config.DefaultMergeConfig()
}

// MergeCrossPageBlocks merges adjacent Text blocks across pages using default configuration.
func MergeCrossPageBlocks(blocks []LayoutBlock) ([]LayoutBlock, error) {
	return MergeCrossPageBlocksWithConfig(blocks, DefaultMergeConfig())
}

// MergeCrossPageBlocksWithConfig merges adjacent Text blocks using the given configuration.
func MergeCrossPageBlocksWithConfig(blocks []LayoutBlock, cfg MergeConfig) ([]LayoutBlock, error) {
	if len(blocks) == 0 {
		return blocks, nil
	}

	if cfg.MaxLookAhead == 0 {
		cfg.MaxLookAhead = 9
	}

	var result []LayoutBlock
	merged := make([]bool, len(blocks))

	for i := 0; i < len(blocks); i++ {
		if merged[i] {
			continue
		}

		current := blocks[i]

		if current.Label == LabelTOC {
			nextIdx := -1
			limit := i + cfg.MaxLookAhead
			if limit > len(blocks) {
				limit = len(blocks)
			}
			for j := i + 1; j < limit; j++ {
				if merged[j] {
					continue
				}
				if blocks[j].Label == LabelTOC {
					nextIdx = j
					break
				}
			}
			if nextIdx != -1 {
				if mergedHTML, err := mergeTOCTables(current.Raw, blocks[nextIdx].Raw); err == nil {
					current.Raw = mergedHTML
					merged[nextIdx] = true
					continue
				}
			}
			result = append(result, current)
			continue
		}

		if current.Label != LabelText {
			result = append(result, current)
			continue
		}

		textA := strings.TrimSpace(ExtractText(current.Raw))
		last := lastVisibleChar(textA)

		if isSentenceEnd(last) {
			result = append(result, current)
			continue
		}

		needsMerge := false

		if last == '-' {
			modified := stripTrailingDash(current.Raw)
			if modified != "" {
				current.Raw = modified
			}
			needsMerge = true
		} else if textA != "" {
			needsMerge = true
		}

		if !needsMerge {
			result = append(result, current)
			continue
		}

		innerA := stripOuterTag(current.Raw)
		if innerA == "" {
			innerA = ExtractText(current.Raw)
		}

		limit := i + cfg.MaxLookAhead
		if limit > len(blocks) {
			limit = len(blocks)
		}

		for j := i + 1; j < limit; j++ {
			if merged[j] {
				continue
			}

			cand := blocks[j]
			if cand.Label != LabelText {
				continue
			}

			textB := strings.TrimSpace(ExtractText(cand.Raw))

			if startsWithUpper(textB) {
				continue
			}

			innerB := stripOuterTag(cand.Raw)
			if innerB == "" {
				innerB = ExtractText(cand.Raw)
			}

			innerA = innerA + " " + innerB
			merged[j] = true

			if isSentenceEnd(lastVisibleChar(innerA)) {
				break
			}
		}

		current.Raw = "<p>" + innerA + "</p>"

		result = append(result, current)
	}

	return result, nil
}

// mergeTOCTables merges two Table-Of-Contents blocks by appending the table rows of the
// second block onto the first block's table, producing a single continuous table.
func mergeTOCTables(rawA, rawB string) (string, error) {
	docA, err := html.Parse(strings.NewReader(rawA))
	if err != nil {
		return "", err
	}
	docB, err := html.Parse(strings.NewReader(rawB))
	if err != nil {
		return "", err
	}

	tableA := findNodeByTag(docA, "table")
	tableB := findNodeByTag(docB, "table")
	if tableA == nil || tableB == nil {
		return rawA + rawB, nil
	}

	var trNodes []*html.Node
	var walkB func(*html.Node)
	walkB = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			trNodes = append(trNodes, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkB(c)
		}
	}
	walkB(tableB)

	for _, tr := range trNodes {
		tr.Parent = nil
		tr.PrevSibling = nil
		tr.NextSibling = nil
		tableA.AppendChild(tr)
	}

	var buf bytes.Buffer
	bodyA := FindBody(docA)
	if bodyA != nil {
		for c := bodyA.FirstChild; c != nil; c = c.NextSibling {
			html.Render(&buf, c)
		}
	} else {
		html.Render(&buf, docA)
	}

	return buf.String(), nil
}

// brRe matches <br>, <br/>, <br />, and <BR/> style line-break tags regardless of case.
var brRe = regexp.MustCompile(`(?i)<br\s*/?>`)

// ReplaceBRWithSpaces replaces all line-break (<br>) tags in the HTML with a single space.
// Used for block types where a literal <br> would break markdown/text rendering.
func ReplaceBRWithSpaces(raw string) string {
	return brRe.ReplaceAllString(raw, " ")
}

// findNodeByTag returns the first element node with the given tag name in the subtree rooted at n.
func findNodeByTag(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNodeByTag(c, tag); found != nil {
			return found
		}
	}
	return nil
}

// stripOuterTag extracts the inner text from a <p>...</p> wrapped string.
func stripOuterTag(raw string) string {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)

	if !strings.HasPrefix(lower, "<p") || !strings.HasSuffix(lower, "</p>") {
		return ""
	}

	openEnd := strings.Index(raw, ">")
	if openEnd == -1 {
		return ""
	}

	closeStart := strings.LastIndex(raw, "</p>")
	if closeStart == -1 {
		return ""
	}

	return raw[openEnd+1 : closeStart]
}

// stripTrailingDash removes a trailing hyphen from the last text node of an HTML string.
func stripTrailingDash(raw string) string {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return ""
	}

	var lastTextNode *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			lastTextNode = n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if lastTextNode == nil {
		return ""
	}

	data := lastTextNode.Data
	runes := []rune(data)
	for i := len(runes) - 1; i >= 0; i-- {
		if !unicode.IsSpace(runes[i]) {
			if runes[i] == '-' {
				lastTextNode.Data = string(runes[:i])
				var buf bytes.Buffer
				html.Render(&buf, doc)
				return buf.String()
			}
			return raw
		}
	}
	return raw
}

// lastVisibleChar returns the last non-space rune in a string.
func lastVisibleChar(text string) rune {
	runes := []rune(text)
	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]
		if !unicode.IsSpace(r) {
			return r
		}
	}
	return 0
}

// isSentenceEnd returns true if the rune is a sentence-ending punctuation mark.
func isSentenceEnd(r rune) bool {
	return r == '.' || r == '!' || r == '?' || r == '…' || r == ':' || r == ';'
}

// startsWithUpper returns true if the first non-space rune in the string is uppercase.
func startsWithUpper(text string) bool {
	runes := []rune(text)
	for _, r := range runes {
		if !unicode.IsSpace(r) {
			return unicode.IsUpper(r)
		}
	}
	return false
}

// brBlockLabels are the block labels whose <br> tags must be normalized to spaces.
var brBlockLabels = map[LabelType]bool{
	LabelSectionHeader: true,
	LabelPageHeader:    true,
	LabelTable:         true,
	LabelTOC:           true,
}

// NormalizeBlockLineBreaks replaces <br> tags with spaces in the Raw HTML of blocks
// whose label is a header, table, or table-of-contents.
func NormalizeBlockLineBreaks(blocks []LayoutBlock) {
	for i := range blocks {
		if brBlockLabels[blocks[i].Label] {
			blocks[i].Raw = ReplaceBRWithSpaces(blocks[i].Raw)
		}
	}
}
