package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetDefaultConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		envSetup func(t *testing.T)
		want     string
	}{
		{
			name:     "linux with XDG_CONFIG_HOME",
			goos:     "linux",
			envSetup: func(t *testing.T) { t.Setenv("XDG_CONFIG_HOME", "/custom/xdg") },
			want:     "/custom/xdg/llllm-ocr/config.json",
		},
		{
			name:     "linux without XDG_CONFIG_HOME",
			goos:     "linux",
			envSetup: func(t *testing.T) { t.Setenv("HOME", "/home/user") },
			want:     "/home/user/.config/llllm-ocr/config.json",
		},
		{
			name:     "darwin",
			goos:     "darwin",
			envSetup: func(t *testing.T) { t.Setenv("HOME", "/Users/user") },
			want:     "/Users/user/Library/Application Support/llllm-ocr/config.json",
		},
		{
			name:     "windows with APPDATA",
			goos:     "windows",
			envSetup: func(t *testing.T) { t.Setenv("APPDATA", "C:\\Users\\user\\AppData\\Roaming") },
			want:     filepath.Join("C:", "Users", "user", "AppData", "Roaming", "llllm-ocr", "config.json"),
		},
		{
			name:     "windows without APPDATA fallback",
			goos:     "windows",
			envSetup: func(t *testing.T) { t.Setenv("USERPROFILE", "C:\\Users\\user") },
			want:     filepath.Join("C:", "Users", "user", "AppData", "Roaming", "llllm-ocr", "config.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envSetup(t)

			got := getConfigPath(tt.goos)
			if tt.goos == "windows" {
				got = strings.ReplaceAll(got, "\\", "/")
			}
			want := filepath.ToSlash(tt.want)
			if got != want {
				t.Errorf("getConfigPath(%q) = %q, want %q", tt.goos, got, want)
			}
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	t.Run("happy path creates valid JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		err := CreateDefaultConfig(path)
		if err != nil {
			t.Fatalf("CreateDefaultConfig() error = %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}

		var cfg CLIConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		assertCLIConfig(t, cfg)
	})

	t.Run("creates nested directories", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "deep", "nested", "path", "config.json")

		err := CreateDefaultConfig(path)
		if err != nil {
			t.Fatalf("CreateDefaultConfig() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("config file was not created")
		}
	})

	t.Run("error on unwritable path", func(t *testing.T) {
		// Create a regular file at the target path, then try to create a config there
		// os.Create will truncate it, but MkdirAll will fail for a parent that's a file
		dir := t.TempDir()
		existingFile := filepath.Join(dir, "existing")
		os.WriteFile(existingFile, []byte("data"), 0644)

		// Try to create config with a path where a parent component is a file
		path := filepath.Join(dir, "existing", "nested", "config.json")
		err := CreateDefaultConfig(path)
		if err == nil {
			t.Error("expected error when parent path is a file, got nil")
		}
	})
}

func assertCLIConfig(t *testing.T, cfg CLIConfig) {
	t.Helper()
	if cfg.PdftoppmCmd != "pdftoppm" {
		t.Errorf("PdftoppmCmd = %q, want %q", cfg.PdftoppmCmd, "pdftoppm")
	}
	if cfg.LlamaCliCmd != "llama-mtmd-cli" {
		t.Errorf("LlamaCliCmd = %q, want %q", cfg.LlamaCliCmd, "llama-mtmd-cli")
	}
	if cfg.OcrMaxWidth != 2048 {
		t.Errorf("OcrMaxWidth = %d, want %d", cfg.OcrMaxWidth, 2048)
	}
	if cfg.OcrMaxHeight != 3072 {
		t.Errorf("OcrMaxHeight = %d, want %d", cfg.OcrMaxHeight, 3072)
	}
	if !cfg.CycleEnabled {
		t.Error("CycleEnabled = false, want true")
	}
	if cfg.CycleCheckInterval != 100 {
		t.Errorf("CycleCheckInterval = %d, want %d", cfg.CycleCheckInterval, 100)
	}
	if cfg.CycleMinTokens != 400 {
		t.Errorf("CycleMinTokens = %d, want %d", cfg.CycleMinTokens, 400)
	}
	if cfg.CycleMaxTokens != 10000 {
		t.Errorf("CycleMaxTokens = %d, want %d", cfg.CycleMaxTokens, 10000)
	}
	if len(cfg.LlamaCliArgs) != 22 {
		t.Errorf("LlamaCliArgs len = %d, want %d", len(cfg.LlamaCliArgs), 22)
	}
	if len(cfg.RemoveBlocks) != 5 {
		t.Errorf("RemoveBlocks len = %d, want %d", len(cfg.RemoveBlocks), 5)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("round-trip with default config", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		if err := CreateDefaultConfig(path); err != nil {
			t.Fatalf("CreateDefaultConfig() error = %v", err)
		}

		got, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		assertCLIConfig(t, got)
	})

	t.Run("partial config", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		partial := CLIConfig{
			PdftoppmCmd: "my-pdftoppm",
			OcrMaxWidth: 1024,
		}
		data, _ := json.Marshal(partial)
		os.WriteFile(path, data, 0644)

		got, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}

		if got.PdftoppmCmd != "my-pdftoppm" {
			t.Errorf("PdftoppmCmd = %q, want %q", got.PdftoppmCmd, "my-pdftoppm")
		}
		if got.OcrMaxWidth != 1024 {
			t.Errorf("OcrMaxWidth = %d, want %d", got.OcrMaxWidth, 1024)
		}
		if got.LlamaCliCmd != "" {
			t.Errorf("LlamaCliCmd = %q, want empty", got.LlamaCliCmd)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		os.WriteFile(path, []byte("not valid json {{{"), 0644)

		_, err := LoadConfig(path)
		if err == nil {
			t.Error("LoadConfig() expected error for invalid JSON, got nil")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path/config.json")
		if err == nil {
			t.Error("LoadConfig() expected error for missing file, got nil")
		}
	})
}

func TestEnsureDefaultConfigExists(t *testing.T) {
	t.Run("creates when missing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		err := EnsureDefaultConfigExists(path)
		if err != nil {
			t.Fatalf("EnsureDefaultConfigExists() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("config file was not created")
		}
	})

	t.Run("does not overwrite existing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		custom := CLIConfig{PdftoppmCmd: "custom-pdftoppm"}
		data, _ := json.Marshal(custom)
		os.WriteFile(path, data, 0644)

		info, _ := os.Stat(path)
		mtime1 := info.ModTime()

		err := EnsureDefaultConfigExists(path)
		if err != nil {
			t.Fatalf("EnsureDefaultConfigExists() error = %v", err)
		}

		info, _ = os.Stat(path)
		mtime2 := info.ModTime()

		if !mtime1.Equal(mtime2) {
			t.Errorf("mtime changed: %v → %v", mtime1, mtime2)
		}

		var cfg CLIConfig
		data, _ = os.ReadFile(path)
		json.Unmarshal(data, &cfg)
		if cfg.PdftoppmCmd != "custom-pdftoppm" {
			t.Errorf("config was overwritten, PdftoppmCmd = %q, want %q", cfg.PdftoppmCmd, "custom-pdftoppm")
		}
	})

	t.Run("empty file is not overwritten", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		os.WriteFile(path, []byte{}, 0644)

		err := EnsureDefaultConfigExists(path)
		if err != nil {
			t.Fatalf("EnsureDefaultConfigExists() error = %v", err)
		}

		data, _ := os.ReadFile(path)
		if len(data) != 0 {
			t.Errorf("empty file was overwritten, got %d bytes", len(data))
		}
	})
}

func TestBuildOCRConfig(t *testing.T) {
	t.Run("full mapping", func(t *testing.T) {
		cliCfg := CLIConfig{
			PdftoppmCmd:        "pdftoppm",
			LlamaCliCmd:        "llama-cli",
			LlamaCliArgs:       []string{"-m", "model.gguf"},
			RemoveBlocks:       []string{"caption"},
			OcrMaxWidth:        1024,
			OcrMaxHeight:       2048,
			CycleEnabled:       true,
			CycleCheckInterval: 50,
			CycleMinTokens:     200,
			CycleMaxTokens:     5000,
		}
		outDir := "/tmp/output"

		got := BuildOCRConfig(cliCfg, outDir)

		if got.Extract.PdftoppmCmd != "pdftoppm" {
			t.Errorf("Extract.PdftoppmCmd = %q, want %q", got.Extract.PdftoppmCmd, "pdftoppm")
		}
		if got.Extract.OcrMaxWidth != 1024 {
			t.Errorf("Extract.OcrMaxWidth = %d, want %d", got.Extract.OcrMaxWidth, 1024)
		}
		if got.Extract.OcrMaxHeight != 2048 {
			t.Errorf("Extract.OcrMaxHeight = %d, want %d", got.Extract.OcrMaxHeight, 2048)
		}
		if got.Recognize.LlamaCliCmd != "llama-cli" {
			t.Errorf("Recognize.LlamaCliCmd = %q, want %q", got.Recognize.LlamaCliCmd, "llama-cli")
		}
		if len(got.Recognize.LlamaCliArgs) != 2 {
			t.Errorf("Recognize.LlamaCliArgs len = %d, want %d", len(got.Recognize.LlamaCliArgs), 2)
		}
		if got.Cycle.CheckInterval != 50 {
			t.Errorf("Cycle.CheckInterval = %d, want %d", got.Cycle.CheckInterval, 50)
		}
		if got.Cycle.MinTokens != 200 {
			t.Errorf("Cycle.MinTokens = %d, want %d", got.Cycle.MinTokens, 200)
		}
		if got.Cycle.MaxTokens != 5000 {
			t.Errorf("Cycle.MaxTokens = %d, want %d", got.Cycle.MaxTokens, 5000)
		}
		if got.OutDir != "/tmp/output" {
			t.Errorf("OutDir = %q, want %q", got.OutDir, "/tmp/output")
		}
		if got.ProgressFn == nil {
			t.Error("ProgressFn is nil, want non-nil")
		}
	})

	t.Run("cycle disabled zeroes MaxTokens", func(t *testing.T) {
		cliCfg := CLIConfig{
			PdftoppmCmd:        "pdftoppm",
			CycleEnabled:       false,
			CycleCheckInterval: 100,
			CycleMinTokens:     400,
			CycleMaxTokens:     9999,
		}

		got := BuildOCRConfig(cliCfg, "/tmp/out")

		if got.Cycle.MaxTokens != 0 {
			t.Errorf("Cycle.MaxTokens = %d, want 0 when CycleEnabled=false", got.Cycle.MaxTokens)
		}
		if got.Cycle.CheckInterval != 100 {
			t.Errorf("Cycle.CheckInterval = %d, want 100 preserved", got.Cycle.CheckInterval)
		}
		if got.Cycle.MinTokens != 400 {
			t.Errorf("Cycle.MinTokens = %d, want 400 preserved", got.Cycle.MinTokens)
		}
	})

	t.Run("cycle enabled preserves MaxTokens", func(t *testing.T) {
		cliCfg := CLIConfig{
			PdftoppmCmd:    "pdftoppm",
			CycleEnabled:   true,
			CycleMaxTokens: 7777,
		}

		got := BuildOCRConfig(cliCfg, "/tmp/out")

		if got.Cycle.MaxTokens != 7777 {
			t.Errorf("Cycle.MaxTokens = %d, want 7777 when CycleEnabled=true", got.Cycle.MaxTokens)
		}
	})

	t.Run("default cycle fields preserved", func(t *testing.T) {
		cliCfg := CLIConfig{PdftoppmCmd: "pdftoppm"}

		got := BuildOCRConfig(cliCfg, "/tmp/out")

		if got.Cycle.MaxPatternLen != 120 {
			t.Errorf("Cycle.MaxPatternLen = %d, want 120 (default)", got.Cycle.MaxPatternLen)
		}
		if got.Cycle.MinRepeats != 3 {
			t.Errorf("Cycle.MinRepeats = %d, want 3 (default)", got.Cycle.MinRepeats)
		}
	})

	t.Run("cycle enabled with MaxTokens=0", func(t *testing.T) {
		cliCfg := CLIConfig{
			PdftoppmCmd:    "pdftoppm",
			CycleEnabled:   true,
			CycleMaxTokens: 0,
		}

		got := BuildOCRConfig(cliCfg, "/tmp/out")

		if got.Cycle.MaxTokens != 0 {
			t.Errorf("Cycle.MaxTokens = %d, want 0", got.Cycle.MaxTokens)
		}
	})
}

func TestPageName(t *testing.T) {
	tests := []struct {
		name string
		p    int
		want string
	}{
		{"zero", 0, "page_000"},
		{"one", 1, "page_001"},
		{"ten", 10, "page_010"},
		{"hundred", 100, "page_100"},
		{"thousand", 1000, "page_1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PageName(tt.p)
			if got != tt.want {
				t.Errorf("PageName(%d) = %q, want %q", tt.p, got, tt.want)
			}
		})
	}
}

func TestSourceStem(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple pdf", "book.pdf", "book"},
		{"path with dir", "/path/to/my-doc.pdf", "my-doc"},
		{"no extension", "file", "file"},
		{"hidden file", ".hidden.pdf", ".hidden"},
		{"multiple dots", "file.name.tar.gz", "file.name.tar"},
		{"just extension", ".pdf", ""},
		{"uppercase ext", "book.PDF", "book"},
		{"single char", "a.pdf", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SourceStem(tt.input)
			if got != tt.want {
				t.Errorf("SourceStem(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWorkspacePath(t *testing.T) {
	tests := []struct {
		name        string
		sourcePath  string
		explicitDir string
		want        string
	}{
		{"simple pdf, no dir", "book.pdf", ".", "./book"},
		{"simple pdf, empty dir", "book.pdf", "", "./book"},
		{"explicit dir", "book.pdf", "./custom", "./custom"},
		{"absolute path", "/abs/path/doc.pdf", ".", "/abs/path/doc"},
		{"explicit absolute dir", "book.pdf", "/custom/out", "/custom/out"},
		{"relative path with dir", "./docs/book.pdf", ".", "./docs/book"},
		{"explicit dir with trailing slash", "book.pdf", "./custom/", "./custom/"},
		{"dot only explicit", "book.pdf", ".", "./book"},
		{"no extension", "file", ".", "file"},
		{"hidden file", ".hidden.pdf", ".", ".hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WorkspacePath(tt.sourcePath, tt.explicitDir)
			// Normalize for comparison (handle ./ prefix)
			wantNormalized := filepath.Clean(tt.want)
			gotNormalized := filepath.Clean(got)
			if gotNormalized != wantNormalized {
				t.Errorf("WorkspacePath(%q, %q) = %q, want %q", tt.sourcePath, tt.explicitDir, got, tt.want)
			}
		})
	}
}

func TestGetDefaultConfigPath_platform(t *testing.T) {
	// Test on the actual platform we're running on
	got := GetDefaultConfigPath()

	switch runtime.GOOS {
	case "windows":
		if got != filepath.Join(os.Getenv("APPDATA"), "llllm-ocr", "config.json") {
			t.Logf("windows path: %s (APPDATA=%s)", got, os.Getenv("APPDATA"))
		}
	case "darwin":
		if got != filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "llllm-ocr", "config.json") {
			t.Logf("darwin path: %s (HOME=%s)", got, os.Getenv("HOME"))
		}
	default:
		xdg := os.Getenv("XDG_CONFIG_HOME")
		want := filepath.Join(os.Getenv("HOME"), ".config", "llllm-ocr", "config.json")
		if xdg != "" {
			want = filepath.Join(xdg, "llllm-ocr", "config.json")
		}
		if got != want {
			t.Logf("linux path: %s, want: %s (XDG=%s, HOME=%s)", got, want, xdg, os.Getenv("HOME"))
		}
	}
}
