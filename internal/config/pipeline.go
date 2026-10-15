package config

import (
	"fmt"
)

// PipelineConfig is the top-level configuration struct for the OCR pipeline.
// It composes ExtractConfig, RecognizeConfig, CycleConfig, and runtime settings.
type PipelineConfig struct {
	Extract    ExtractConfig
	Recognize  RecognizeConfig
	Cycle      CycleConfig
	OutDir     string
	ProgressFn func(current, total int)
}

// CycleConfig holds the configuration for the cycle detector.
type CycleConfig struct {
	CheckInterval int
	MinTokens     int
	MaxTokens     int
	MaxPatternLen int
	MinRepeats    int
}

// DefaultCycleConfig returns a CycleConfig with sensible defaults for cycle detection.
func DefaultCycleConfig() CycleConfig {
	return CycleConfig{
		CheckInterval: 100,
		MinTokens:     400,
		MaxPatternLen: 120,
		MinRepeats:    3,
		MaxTokens:     10000,
	}
}

// NewPipelineConfig maps a flat CLIConfig into the nested PipelineConfig struct used by the pipeline.
func NewPipelineConfig(cliCfg CLIConfig, outDir string) PipelineConfig {
	cycleCfg := DefaultCycleConfig()
	cycleCfg.CheckInterval = cliCfg.CycleCheckInterval
	cycleCfg.MinTokens = cliCfg.CycleMinTokens
	cycleCfg.MaxTokens = cliCfg.CycleMaxTokens

	if !cliCfg.CycleEnabled {
		cycleCfg.MaxTokens = 0
	}

	recognizeCfg := RecognizeConfig{
		LlamaCliCmd:  cliCfg.LlamaCliCmd,
		LlamaCliArgs: cliCfg.LlamaCliArgs,
		MaxRetries:   cliCfg.RecognizeRetries,
		RetryTempArgs: func() []string {
			if cliCfg.RecognizeRetryTemp != "" {
				return []string{"--temp", cliCfg.RecognizeRetryTemp}
			}
			return nil
		}(),
	}

	pipelineCfg := PipelineConfig{
		Extract: ExtractConfig{
			PdftoppmCmd:  cliCfg.PdftoppmCmd,
			OcrMaxWidth:  cliCfg.OcrMaxWidth,
			OcrMaxHeight: cliCfg.OcrMaxHeight,
		},
		Recognize: recognizeCfg,
		Cycle:     cycleCfg,
		OutDir:    outDir,
		ProgressFn: func(current, total int) {
			fmt.Printf("\nProcessing page %d of %d...\n", current, total)
		},
	}

	return pipelineCfg
}

// BuildOCRConfig maps a flat CLIConfig into the nested PipelineConfig struct used by the pipeline.
// Deprecated: Use NewPipelineConfig instead. Kept for backward compatibility.
func BuildOCRConfig(cliCfg CLIConfig, outDir string) PipelineConfig {
	return NewPipelineConfig(cliCfg, outDir)
}
