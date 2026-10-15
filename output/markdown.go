package output

import (
	"regexp"
	"strings"

	"github.com/Galdoba/lllm-ocr/internal/config"
	"github.com/Galdoba/lllm-ocr/layoutblock"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"golang.org/x/net/html"
)

// MarkdownConfig is an alias for config.MarkdownConfig for backward compatibility.
type MarkdownConfig = config.MarkdownConfig

type MarkdownOutput struct {
	Config MarkdownConfig
}

// Convert renders blocks as Markdown using html-to-markdown with table, strikethrough, and task list plugins.
func (o *MarkdownOutput) Convert(blocks []layoutblock.LayoutBlock) string {
	converter := md.NewConverter("", true, nil).Use(
		plugin.Table(),
		plugin.Strikethrough("-"),
		plugin.TaskListItems(),
	)

	var sb strings.Builder
	var lastPage int

	for i, block := range blocks {
		if i > 0 && block.Page != lastPage {
			sb.WriteString(o.Config.PageSeparator)
		}
		lastPage = block.Page

		switch block.Label {
		case layoutblock.LabelPageFooter:
			continue
		case layoutblock.LabelSectionHeader:
			sb.WriteString(convertHeader(block.Raw, converter))
			sb.WriteString("\n\n")
		case layoutblock.LabelText:
			content := convertText(block.Raw, converter)
			if o.Config.MathBlockMode {
				content = processMath(content)
			}
			sb.WriteString(content)
			sb.WriteString("\n\n")
		case layoutblock.LabelTable:
			if o.Config.HideTables {
				sb.WriteString(convertTableToCallout(block.Raw))
				sb.WriteString("\n\n")
			} else {
				content := convertTable(block.Raw, converter)
				sb.WriteString(content)
				sb.WriteString("\n\n")
			}
		case layoutblock.LabelTOC:
			content := convertTable(block.Raw, converter)
			sb.WriteString(content)
			sb.WriteString("\n\n")
		case layoutblock.LabelListGroup:
			content := convertList(block.Raw, converter)
			sb.WriteString(content)
			sb.WriteString("\n\n")
		default:
			content := convertText(block.Raw, converter)
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}

	return strings.TrimRight(cleanMarkdown(sb.String()), "\n")
}

// convertHeader converts an HTML header element to Markdown heading syntax.
func convertHeader(rawHTML string, converter *md.Converter) string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return convertText(rawHTML, converter)
	}

	tag, text, found := extractTagAndText(doc)
	if !found || (tag != "h1" && tag != "h2" && tag != "h3" && tag != "h4" && tag != "h5" && tag != "h6") {
		return convertText(rawHTML, converter)
	}

	level := 2
	switch tag {
	case "h1":
		level = 1
	case "h2":
		level = 2
	case "h3":
		level = 3
	case "h4":
		level = 4
	case "h5":
		level = 5
	case "h6":
		level = 6
	}

	prefix := strings.Repeat("#", level) + " "
	return prefix + text
}

// extractTagAndText returns the tag name, text content, and success flag for the first element node with text.
func extractTagAndText(n *html.Node) (string, string, bool) {
	if n.Type == html.ElementNode {
		text := layoutblock.ExtractTextFromNode(n)
		if strings.TrimSpace(text) != "" {
			return n.Data, text, true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		tag, text, found := extractTagAndText(c)
		if found {
			return tag, text, true
		}
	}
	return "", "", false
}

// convertText converts an HTML string to Markdown using the provided converter.
func convertText(rawHTML string, converter *md.Converter) string {
	result, err := converter.ConvertString(rawHTML)
	if err != nil {
		return layoutblock.ExtractText(rawHTML)
	}
	return strings.TrimSpace(result)
}

// convertTable converts an HTML table to Markdown table syntax.
func convertTable(rawHTML string, converter *md.Converter) string {
	result, err := converter.ConvertString(rawHTML)
	if err != nil {
		return "<table>\n" + layoutblock.ExtractText(rawHTML) + "\n</table>"
	}
	return strings.TrimSpace(result)
}

// convertList converts an HTML list to Markdown list syntax.
func convertList(rawHTML string, converter *md.Converter) string {
	result, err := converter.ConvertString(rawHTML)
	if err != nil {
		return layoutblock.ExtractText(rawHTML)
	}
	return strings.TrimSpace(result)
}

var mathTagRe = regexp.MustCompile(`(?s)<math[^>]*>(.*?)</math>`)

// processMath converts <math>...</math> tags to LaTeX $$...$$ blocks.
func processMath(text string) string {
	return mathTagRe.ReplaceAllStringFunc(text, func(match string) string {
		latex := mathTagRe.FindSubmatch([]byte(match))
		if len(latex) > 1 {
			content := strings.TrimSpace(string(latex[1]))
			return "\n$$\n" + content + "\n$$\n"
		}
		return match
	})
}

// convertTableToCallout wraps table text in a Markdown callout block.
func convertTableToCallout(rawHTML string) string {
	text := layoutblock.ExtractText(rawHTML)
	return ">![TABLE]\n" + text + "\n>"
}

var complexTableRe = regexp.MustCompile(`<table[\s>].*?(colspan|rowspan).*?</table>`)

// isComplexTable returns true if the HTML contains a table with colspan or rowspan attributes.
func isComplexTable(rawHTML string) bool {
	return complexTableRe.MatchString(rawHTML)
}
