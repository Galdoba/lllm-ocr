package config

// ExtractConfig holds settings for the pdftoppm extraction phase.
type ExtractConfig struct {
	PdftoppmCmd  string
	OcrMaxWidth  int
	OcrMaxHeight int
}

// RecognizeConfig holds settings for the llama-mtmd-cli recognition phase.
type RecognizeConfig struct {
	LlamaCliCmd   string
	LlamaCliArgs  []string
	MaxRetries    int
	RetryTempArgs []string
}
