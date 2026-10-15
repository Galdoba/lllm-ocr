package cycle

import "github.com/Galdoba/lllm-ocr/internal/config"

// Config is an alias for config.CycleConfig for backward compatibility.
type Config = config.CycleConfig

// DefaultConfig returns a Config with sensible defaults for cycle detection.
func DefaultConfig() Config {
	return config.DefaultCycleConfig()
}
