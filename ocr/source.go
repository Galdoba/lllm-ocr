package ocr

import (
	"fmt"

	"github.com/Galdoba/lllm-ocr/internal/config"
)

// Type aliases for centralized config types.
type ExtractConfig = config.ExtractConfig
type RecognizeConfig = config.RecognizeConfig
type Config = config.PipelineConfig

// pageName returns a zero-padded page name string (e.g., "page_001").
func pageName(p int) string {
	return fmt.Sprintf("page_%03d", p)
}
