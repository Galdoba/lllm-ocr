package layoutblock

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type BBox struct {
	X0 int `json:"x0"`
	Y0 int `json:"y0"`
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
}

type LabelType string

const (
	LabelText          LabelType = "Text"
	LabelSectionHeader LabelType = "Section-Header"
	LabelTable         LabelType = "Table"
	LabelListGroup     LabelType = "List-Group"
	LabelPageHeader    LabelType = "Page-Header"
	LabelPageFooter    LabelType = "Page-Footer"
	LabelImage         LabelType = "Image"
	LabelCaption       LabelType = "Caption"
	LabelFigure        LabelType = "Figure"
	LabelDiagram       LabelType = "Diagram"
	LabelFootnote      LabelType = "Footnote"
	LabelTOC           LabelType = "Table-Of-Contents"
	LabelForm          LabelType = "Form"
	LabelEquationBlock LabelType = "Equation-Block"
)

var supportedLabels = map[LabelType]bool{
	LabelText:          true,
	LabelSectionHeader: true,
	LabelTable:         true,
	LabelListGroup:     true,
	LabelPageHeader:    true,
	LabelPageFooter:    true,
	LabelImage:         true,
	LabelCaption:       true,
	LabelFigure:        true,
	LabelDiagram:       true,
	LabelFootnote:      true,
	LabelTOC:           true,
	LabelForm:          true,
	LabelEquationBlock: true,
}

type LayoutBlock struct {
	BBox   BBox      `json:"bbox"`
	Label  LabelType `json:"label"`
	Raw    string    `json:"raw"`
	Page   int       `json:"page"`
	Merged bool      `json:"merged,omitempty"`
}

// Width returns the width of the bounding box.
func (b BBox) Width() int { return b.X1 - b.X0 }

// Height returns the height of the bounding box.
func (b BBox) Height() int { return b.Y1 - b.Y0 }

// Area returns the area of the bounding box.
func (b BBox) Area() int { return b.Width() * b.Height() }

// IsZero returns true if all coordinates are zero.
func (b BBox) IsZero() bool { return b.X0 == 0 && b.Y0 == 0 && b.X1 == 0 && b.Y1 == 0 }

// Intersects returns true if this bounding box intersects with another.
func (b BBox) Intersects(other BBox) bool {
	return b.X0 < other.X1 && b.X1 > other.X0 && b.Y0 < other.Y1 && b.Y1 > other.Y0
}

// Contains returns true if the point (x, y) is inside the bounding box.
func (b BBox) Contains(x, y int) bool {
	return x >= b.X0 && x <= b.X1 && y >= b.Y0 && y <= b.Y1
}

// OverlapArea returns the area of intersection between this bounding box and another.
func (b BBox) OverlapArea(other BBox) int {
	if !b.Intersects(other) {
		return 0
	}
	w := min(b.X1, other.X1) - max(b.X0, other.X0)
	h := min(b.Y1, other.Y1) - max(b.Y0, other.Y0)
	return w * h
}

// SortByPosition sorts blocks by two coordinate axes (e.g., "y0" then "x0").
func SortByPosition(blocks []LayoutBlock, primary, secondary string) {
	sort.SliceStable(blocks, func(i, j int) bool {
		pi := getCoord(blocks[i].BBox, primary)
		pj := getCoord(blocks[j].BBox, primary)
		if pi != pj {
			return pi < pj
		}
		si := getCoord(blocks[i].BBox, secondary)
		sj := getCoord(blocks[j].BBox, secondary)
		return si < sj
	})
}

// BlocksIntersecting returns all blocks whose bounding boxes intersect with the given bbox.
func BlocksIntersecting(blocks []LayoutBlock, bbox BBox) []LayoutBlock {
	var result []LayoutBlock
	for _, b := range blocks {
		if b.BBox.Intersects(bbox) {
			result = append(result, b)
		}
	}
	return result
}

// BlocksContainingPoint returns all blocks whose bounding boxes contain the point (x, y).
func BlocksContainingPoint(blocks []LayoutBlock, x, y int) []LayoutBlock {
	var result []LayoutBlock
	for _, b := range blocks {
		if b.BBox.Contains(x, y) {
			result = append(result, b)
		}
	}
	return result
}

// GroupByVerticalPosition groups blocks into columns based on X coordinate threshold.
func GroupByVerticalPosition(blocks []LayoutBlock, threshold int) [][]LayoutBlock {
	if len(blocks) == 0 {
		return nil
	}
	sort.SliceStable(blocks, func(i, j int) bool {
		return blocks[i].BBox.X0 < blocks[j].BBox.X0
	})
	var columns [][]LayoutBlock
	current := []LayoutBlock{blocks[0]}
	lastX := blocks[0].BBox.X0
	for _, b := range blocks[1:] {
		if b.BBox.X0-lastX > threshold {
			columns = append(columns, current)
			current = []LayoutBlock{b}
		} else {
			current = append(current, b)
		}
		lastX = b.BBox.X0
	}
	columns = append(columns, current)
	return columns
}

// ParsePage parses HTML from an io.Reader and extracts LayoutBlocks from divs with data-bbox and data-label attributes.
func ParsePage(r io.Reader) ([]LayoutBlock, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	var blocks []LayoutBlock
	var walk func(*html.Node) error
	walk = func(n *html.Node) error {
		if n.Type == html.ElementNode && n.Data == "div" {
			bboxAttr := getAttr(n, "data-bbox")
			labelAttr := getAttr(n, "data-label")
			if bboxAttr != "" && labelAttr != "" {
				bbox, err := parseBBox(bboxAttr)
				if err != nil {
					return fmt.Errorf("invalid bbox %q: %w", bboxAttr, err)
				}
				label := LabelType(labelAttr)
				if !supportedLabels[label] {
					return fmt.Errorf("unsupported label %q", labelAttr)
				}
				rawContent, err := renderInnerHTML(n)
				if err != nil {
					return fmt.Errorf("failed to render inner HTML: %w", err)
				}
				blocks = append(blocks, LayoutBlock{
					BBox:  bbox,
					Label: label,
					Raw:   rawContent,
				})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(doc); err != nil {
		return nil, err
	}
	return blocks, nil
}

// getAttr returns the value of the first attribute with the given name on an HTML node.
func getAttr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

// parseBBox parses a space-separated "x0 y0 x1 y1" string into a BBox.
func parseBBox(s string) (BBox, error) {
	parts := strings.Fields(s)
	if len(parts) != 4 {
		return BBox{}, fmt.Errorf("invalid bbox format: %q", s)
	}
	coords := make([]int, 4)
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return BBox{}, fmt.Errorf("invalid bbox coordinate %q: %w", p, err)
		}
		coords[i] = v
	}
	return BBox{X0: coords[0], Y0: coords[1], X1: coords[2], Y1: coords[3]}, nil
}

// renderInnerHTML renders the inner HTML of an node as a string.
func renderInnerHTML(n *html.Node) (string, error) {
	var buf bytes.Buffer
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&buf, c); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

// getCoord returns the value of a named coordinate (x0, y0, x1, y1) from a BBox.
func getCoord(b BBox, name string) int {
	switch name {
	case "x0":
		return b.X0
	case "y0":
		return b.Y0
	case "x1":
		return b.X1
	case "y1":
		return b.Y1
	default:
		return 0
	}
}
