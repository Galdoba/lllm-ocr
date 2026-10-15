package ocr

// This file keeps backward-compatible type aliases for code that may still
// reference the old source package names.  All implementations live in
// ocr/cycle.

import (
	"github.com/Galdoba/lllm-ocr/ocr/cycle"
)

// CycleDetector is an alias for cycle.Detector for backward compatibility.
type CycleDetector = cycle.Detector

// NewCycleDetector creates a detector with default configuration.
func NewCycleDetector() *CycleDetector {
	return cycle.NewDetector(cycle.DefaultConfig())
}

// NewCycleDetectorWithConfig creates a detector with custom check interval, min tokens, and max tokens.
func NewCycleDetectorWithConfig(checkInterval, minTokens, maxTokens int) *CycleDetector {
	cfg := cycle.DefaultConfig()
	cfg.CheckInterval = checkInterval
	cfg.MinTokens = minTokens
	cfg.MaxTokens = maxTokens
	return cycle.NewDetector(cfg)
}
