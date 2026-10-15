package cycle

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/net/html"
)

// CompleteHTMLWithContent parses an HTML file, wraps bare elements in <body> if missing, and writes the result back.
// If crashRecovery is true, prepends a marker comment indicating the file was completed after a crash.
func CompleteHTMLWithContent(htmlPath string, crashRecovery bool) error {
	raw, err := os.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("read html file %s: %w", htmlPath, err)
	}

	if len(raw) == 0 {
		return nil
	}

	// Prepend crash marker if this is a crash recovery.
	if crashRecovery {
		raw = append([]byte("<!-- OCR crash recovery: file completed with partial content -->\n"), raw...)
	}

	doc, parseErr := html.Parse(bytes.NewReader(raw))
	if parseErr != nil {
		return fixHTMLNaivelyPath(htmlPath, raw)
	}

	hasBody := false
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "body" {
			hasBody = true
			break
		}
	}

	if !hasBody {
		body := &html.Node{Type: html.ElementNode, Data: "body"}

		for c := doc.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				body.AppendChild(cloneNode(c))
			}
		}
		newDoc := &html.Node{Type: html.DocumentNode}
		newDoc.AppendChild(body)
		doc = newDoc
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return fixHTMLNaivelyPath(htmlPath, raw)
	}

	return os.WriteFile(htmlPath, buf.Bytes(), 0644)
}

// cloneNode creates a deep copy of an HTML node and all its children.
func cloneNode(n *html.Node) *html.Node {
	clone := &html.Node{
		Type:        n.Type,
		DataAtom:    n.DataAtom,
		Data:        n.Data,
		Namespace:   n.Namespace,
		Attr:        make([]html.Attribute, len(n.Attr)),
		Parent:      nil,
		FirstChild:  nil,
		LastChild:   nil,
		PrevSibling: nil,
		NextSibling: nil,
	}
	copy(clone.Attr, n.Attr)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		clone.AppendChild(cloneNode(c))
	}

	return clone
}

// fixHTMLNaivelyPath appends </body></html> to raw HTML bytes as a fallback when parsing fails.
func fixHTMLNaivelyPath(htmlPath string, raw []byte) error {
	f, err := os.Create(htmlPath)
	if err != nil {
		return fmt.Errorf("create html file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(raw); err != nil {
		return fmt.Errorf("write html content: %w", err)
	}
	if _, err := f.WriteString("\n</body>\n</html>\n"); err != nil {
		return fmt.Errorf("write html closing tags: %w", err)
	}
	return nil
}
