package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const configDirName = "llllm-ocr"

// CLIConfig holds the flat configuration parsed from config.json.
type CLIConfig struct {
	PdftoppmCmd        string   `json:"pdftoppm_cmd"`
	LlamaCliCmd        string   `json:"llama_cli_cmd"`
	LlamaCliArgs       []string `json:"llama_cli_args"`
	RemoveBlocks       []string `json:"remove_blocks"`
	OcrMaxWidth        int      `json:"ocr_max_width"`
	OcrMaxHeight       int      `json:"ocr_max_height"`
	CycleEnabled       bool     `json:"cycle_enabled"`
	CycleCheckInterval int      `json:"cycle_check_interval"`
	CycleMinTokens     int      `json:"cycle_min_tokens"`
	CycleMaxTokens     int      `json:"cycle_max_tokens"`
	RecognizeRetries   int      `json:"recognize_retries"`
	RecognizeRetryTemp string   `json:"recognize_retry_temp"`
}

// GetDefaultConfigPath returns the platform-specific default path for config.json.
func GetDefaultConfigPath() string {
	return getConfigPath(runtime.GOOS)
}

func getConfigPath(goos string) string {
	var configDir string
	switch goos {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		configDir = filepath.Join(configDir, configDirName)
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", configDirName)
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			configDir = filepath.Join(xdg, configDirName)
		} else {
			configDir = filepath.Join(os.Getenv("HOME"), ".config", configDirName)
		}
	}
	return filepath.Join(configDir, "config.json")
}

// CreateDefaultConfig writes a default CLIConfig to the given path.
func CreateDefaultConfig(path string) error {
	cfg := CLIConfig{
		PdftoppmCmd: "pdftoppm",
		LlamaCliCmd: "llama-mtmd-cli",
		LlamaCliArgs: []string{
			"-m", "/path/to/model.gguf",
			"--mmproj", "/path/to/mmproj.gguf",
			"other", "parameters",
		},
		RemoveBlocks: []string{
			"caption",
			"image",
			"page_header",
			"page_footer",
			"form",
		},
		OcrMaxWidth:        2048,
		OcrMaxHeight:       3072,
		CycleEnabled:       true,
		CycleCheckInterval: 100,
		CycleMinTokens:     400,
		CycleMaxTokens:     10000,
		RecognizeRetries:   3,
		RecognizeRetryTemp: "1",
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// LoadConfig reads and parses a CLIConfig JSON file from the given path.
func LoadConfig(path string) (CLIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CLIConfig{}, fmt.Errorf("reading config file '%s': %w", path, err)
	}

	var cfg CLIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return CLIConfig{}, fmt.Errorf("parsing config file: %w", err)
	}

	return cfg, nil
}

// EnsureDefaultConfigExists creates a default config file at the given path if it does not already exist.
func EnsureDefaultConfigExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return CreateDefaultConfig(path)
	}
	return nil
}

// PageName returns a zero-padded page name string (e.g., "page_001").
func PageName(p int) string {
	return fmt.Sprintf("page_%03d", p)
}

// SourceStem returns the filename stem (without extension) from a path.
// "book.pdf" → "book"
// "/path/to/my-doc.pdf" → "my-doc"
// "file" → "file"  (no extension)
// ".hidden.pdf" → ".hidden"
func SourceStem(sourcePath string) string {
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}

// WorkspacePath derives the workspace output directory from a source path
// and an optional explicit directory override.
//
// Rules:
//   - If explicitDir is non-empty and not ".": return explicitDir (backward compat)
//   - Otherwise: return filepath.Join(filepath.Dir(sourcePath), SourceStem(sourcePath))
//
// Examples:
//
//	WorkspacePath("book.pdf", ".")          → "./book"
//	WorkspacePath("book.pdf", "")           → "./book"
//	WorkspacePath("book.pdf", "./custom")   → "./custom"
//	WorkspacePath("/abs/path/doc.pdf", ".") → "/abs/path/doc"
func WorkspacePath(sourcePath, explicitDir string) string {
	if explicitDir != "" && explicitDir != "." {
		return explicitDir
	}
	return filepath.Join(filepath.Dir(sourcePath), SourceStem(sourcePath))
}
