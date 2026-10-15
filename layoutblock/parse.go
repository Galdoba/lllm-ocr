package layoutblock

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Galdoba/lllm-ocr/internal/config"
	"golang.org/x/net/html"
)

// ParserConfig is an alias for config.ParserConfig for backward compatibility.
type ParserConfig = config.ParserConfig

type BlockParser struct {
	cfg ParserConfig
}

// NewBlockParser creates a BlockParser with the given lenient mode.
func NewBlockParser(lenient bool) *BlockParser {
	return &BlockParser{
		cfg: ParserConfig{Lenient: lenient},
	}
}

// Parse parses HTML from a string or file path and extracts LayoutBlocks.
// If input contains '<', it is treated as raw HTML; otherwise it is treated as a file path.
func (bp *BlockParser) Parse(input string) ([]LayoutBlock, error) {
	htmlContent := input

	if !strings.Contains(input, "<") {
		content, err := os.ReadFile(input)
		if err != nil {
			return nil, fmt.Errorf("read html file %s: %w", input, err)
		}
		htmlContent = string(content)
	}

	return parse(htmlContent, bp.cfg.Lenient)
}

var pageRegex = regexp.MustCompile(`(?i)PAGE (\d+)`)

// getInnerHtml renders the inner HTML of an node as a string.
func getInnerHtml(n *html.Node) string {
	var buf bytes.Buffer
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		_ = html.Render(&buf, c)
	}
	return buf.String()
}

// parse parses an HTML string and extracts LayoutBlocks, optionally skipping errors in lenient mode.
func parse(input string, lenient bool) ([]LayoutBlock, error) {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var blocks []LayoutBlock
	var errs []string
	currentPage := 0

	var traverse func(*html.Node) error
	traverse = func(n *html.Node) error {
		if n.Type == html.CommentNode {
			if matches := pageRegex.FindStringSubmatch(n.Data); len(matches) > 1 {
				if p, err := strconv.Atoi(matches[1]); err == nil {
					currentPage = p
				}
			}
		}

		if n.Type == html.ElementNode && n.Data == "div" {
			var bboxStr, labelStr string
			for _, attr := range n.Attr {
				if attr.Key == "data-bbox" {
					bboxStr = attr.Val
				} else if attr.Key == "data-label" {
					labelStr = attr.Val
				}
			}

			if bboxStr != "" && labelStr != "" {
				label := LabelType(labelStr)
				if !supportedLabels[label] {
					err := fmt.Errorf("unsupported label at page %d: %s", currentPage, labelStr)
					if !lenient {
						return err
					}
					errs = append(errs, err.Error())
				} else {
					parts := strings.Fields(bboxStr)
					if len(parts) != 4 {
						err := fmt.Errorf("invalid bbox format at page %d: %s", currentPage, bboxStr)
						if !lenient {
							return err
						}
						errs = append(errs, err.Error())
					} else {
						x0, e1 := strconv.Atoi(parts[0])
						y0, e2 := strconv.Atoi(parts[1])
						x1, e3 := strconv.Atoi(parts[2])
						y1, e4 := strconv.Atoi(parts[3])
						if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
							err := fmt.Errorf("invalid bbox coordinates at page %d: %s", currentPage, bboxStr)
							if !lenient {
								return err
							}
							errs = append(errs, err.Error())
						} else {
							raw := getInnerHtml(n)
							blocks = append(blocks, LayoutBlock{
								BBox: BBox{
									X0: x0,
									Y0: y0,
									X1: x1,
									Y1: y1,
								},
								Label: label,
								Raw:   raw,
								Page:  currentPage,
							})
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := traverse(c); err != nil {
				return err
			}
		}
		return nil
	}

	if err := traverse(doc); err != nil {
		return nil, err
	}

	if len(errs) > 0 {
		return blocks, fmt.Errorf("errors encountered: %s", strings.Join(errs, "\n"))
	}

	return blocks, nil
}
