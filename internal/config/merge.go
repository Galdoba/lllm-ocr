package config

// MergeConfig controls cross-page text merging look-ahead window.
type MergeConfig struct {
	MaxLookAhead int
}

// DefaultMergeConfig returns a MergeConfig with default look-ahead limit.
func DefaultMergeConfig() MergeConfig {
	return MergeConfig{
		MaxLookAhead: 9,
	}
}
