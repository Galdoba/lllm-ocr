package output

import (
	"github.com/Galdoba/lllm-ocr/layoutblock"
)

// NewMarkdownOutput creates a MarkdownOutput with the given config.
func NewMarkdownOutput(cfg MarkdownConfig) Output {
	return &MarkdownOutput{Config: cfg}
}

type Output interface {
	Convert([]layoutblock.LayoutBlock) string
}
