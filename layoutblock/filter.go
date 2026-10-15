package layoutblock

import (
	"strings"
	"unicode"

	"github.com/Galdoba/lllm-ocr/internal/config"
)

// FilterConfig is an alias for config.FilterConfig for backward compatibility.
type FilterConfig = config.FilterConfig

// DefaultFilterConfig returns a FilterConfig with default artifact detection thresholds.
func DefaultFilterConfig() FilterConfig {
	return config.DefaultFilterConfig()
}

// isNumeric returns true if the string contains only digit characters.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// FilterBlocks removes blocks based on FilterConfig rules: label removal, think tags, empty text, and page-number artifacts.
func FilterBlocks(blocks []LayoutBlock, cfg FilterConfig) []LayoutBlock {
	if cfg.MinArtifactHeight == 0 {
		cfg.MinArtifactHeight = 50
	}
	if cfg.MinArtifactY == 0 {
		cfg.MinArtifactY = 900
	}

	var result []LayoutBlock

	removeLabels := make(map[LabelType]bool)
	for _, l := range cfg.RemoveLabels {
		removeLabels[LabelType(l)] = true
	}

	for _, b := range blocks {
		if removeLabels[b.Label] {
			continue
		}

		if !cfg.KeepThinkTags && strings.Contains(b.Raw, "<think") {
			continue
		}

		text := strings.TrimSpace(ExtractText(b.Raw))

		if text == "" {
			continue
		}

		if !cfg.KeepPageNumbers {
			if b.BBox.Height() < cfg.MinArtifactHeight && b.BBox.Y1 > cfg.MinArtifactY && isNumeric(text) {
				continue
			}
		}

		result = append(result, b)
	}
	return result
}
