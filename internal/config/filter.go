package config

// FilterConfig controls which layout blocks are filtered out during post-processing.
type FilterConfig struct {
	RemoveLabels     []string
	KeepPageNumbers  bool
	KeepThinkTags    bool
	MinArtifactHeight int
	MinArtifactY     int
}

// DefaultFilterConfig returns a FilterConfig with default artifact detection thresholds.
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		MinArtifactHeight: 50,
		MinArtifactY:      900,
	}
}
