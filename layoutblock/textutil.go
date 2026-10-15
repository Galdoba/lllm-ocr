package layoutblock

import (
	"strings"

	"golang.org/x/net/html"
)

// ExtractText extracts plain text from an HTML string by parsing and traversing text nodes.
func ExtractText(raw string) string {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return raw
	}
	return ExtractTextFromNode(doc)
}

// ExtractTextFromNode extracts plain text from an HTML node and all its descendants.
func ExtractTextFromNode(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(ExtractTextFromNode(c))
	}
	return sb.String()
}

// FindBody searches an HTML document tree for the first body element node.
func FindBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := FindBody(c); found != nil {
			return found
		}
	}
	return nil
}
