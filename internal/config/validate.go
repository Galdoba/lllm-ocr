package config

import (
	"fmt"
)

// ValidateExtractConfig checks that ExtractConfig has the required fields set.
func ValidateExtractConfig(cfg ExtractConfig) error {
	if cfg.PdftoppmCmd == "" {
		return fmt.Errorf("Extract.PdftoppmCmd is required")
	}
	if cfg.OcrMaxWidth <= 0 {
		return fmt.Errorf("Extract.OcrMaxWidth must be > 0")
	}
	if cfg.OcrMaxHeight <= 0 {
		return fmt.Errorf("Extract.OcrMaxHeight must be > 0")
	}
	return nil
}

// ValidateRecognizeConfig checks that RecognizeConfig has the required fields set.
func ValidateRecognizeConfig(cfg RecognizeConfig) error {
	if cfg.LlamaCliCmd == "" {
		return fmt.Errorf("Recognize.LlamaCliCmd is required")
	}
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("Recognize.MaxRetries must be >= 0")
	}
	return nil
}

// ValidateCycleConfig checks that CycleConfig has valid values.
func ValidateCycleConfig(cfg CycleConfig) error {
	if cfg.CheckInterval <= 0 {
		return fmt.Errorf("cycle.CheckInterval must be > 0")
	}
	if cfg.MinTokens <= 0 {
		return fmt.Errorf("cycle.MinTokens must be > 0")
	}
	if cfg.MaxTokens < 0 {
		return fmt.Errorf("cycle.MaxTokens must be >= 0 (0 disables detection)")
	}
	if cfg.MaxPatternLen <= 0 {
		return fmt.Errorf("cycle.MaxPatternLen must be > 0")
	}
	if cfg.MinRepeats <= 0 {
		return fmt.Errorf("cycle.MinRepeats must be > 0")
	}
	return nil
}

// ValidatePipelineConfig validates all config sections and returns the first error encountered.
func ValidatePipelineConfig(outDir string, extract ExtractConfig, recognize RecognizeConfig, cycleCfg CycleConfig) error {
	if outDir == "" {
		return fmt.Errorf("OutDir is required")
	}
	if err := ValidateExtractConfig(extract); err != nil {
		return fmt.Errorf("extract config: %w", err)
	}
	if err := ValidateRecognizeConfig(recognize); err != nil {
		return fmt.Errorf("recognize config: %w", err)
	}
	if err := ValidateCycleConfig(cycleCfg); err != nil {
		return fmt.Errorf("cycle config: %w", err)
	}
	return nil
}

// ValidatePipeline validates a PipelineConfig and returns the first error encountered.
func ValidatePipeline(cfg PipelineConfig) error {
	return ValidatePipelineConfig(cfg.OutDir, cfg.Extract, cfg.Recognize, cfg.Cycle)
}
