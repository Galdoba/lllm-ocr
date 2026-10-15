package ocr

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// ExtractPages runs pdftoppm for each page in the given slice, producing per-page PNG files.
// Accepts Config, pdfPath (input PDF), and pages (page numbers to extract).
// Returns a slice of extracted page numbers and any error encountered.
func ExtractPages(cfg Config, pdfPath string, pages []int) ([]int, error) {
	if cfg.Extract.OcrMaxWidth <= 0 || cfg.Extract.OcrMaxHeight <= 0 {
		return nil, fmt.Errorf("two-pass extraction requires OcrMaxWidth and OcrMaxHeight to be set (> 0)")
	}

	var result []int
	for _, p := range pages {
		if cfg.ProgressFn != nil {
			cfg.ProgressFn(p, len(pages))
		}

		pngPath := fmt.Sprintf("%s/%s.png", cfg.OutDir, pageName(p))

		if err := extractPageTwoPass(cfg, pdfPath, p, pngPath); err != nil {
			return result, err
		}

		result = append(result, p)
	}

	return result, nil
}

// extractPageTwoPass probes the page dimensions at scale 1000, computes optimal scale to fit within max bounds, then extracts.
func extractPageTwoPass(cfg Config, pdfPath string, page int, pngPath string) error {
	probeScale := 1000
	probePrefix := fmt.Sprintf("probe_%s", pageName(page))

	probeArgs := []string{"-png", "-scale-to", strconv.Itoa(probeScale),
		"-f", strconv.Itoa(page), "-l", strconv.Itoa(page),
		pdfPath, filepath.Join(cfg.OutDir, probePrefix)}

	probeCmd := exec.Command(cfg.Extract.PdftoppmCmd, probeArgs...)
	probeCmd.Stderr = os.Stderr

	if err := probeCmd.Run(); err != nil {
		return fmt.Errorf("pdftoppm probe failed for page %d: %w", page, err)
	}

	probePattern := fmt.Sprintf("%s/%s-*.png", cfg.OutDir, probePrefix)
	probeMatches, err := filepath.Glob(probePattern)
	if err != nil {
		return fmt.Errorf("glob failed for probe: %w", err)
	}
	if len(probeMatches) == 0 {
		// Clean up any leftover probe files matching the pattern.
		matches, _ := filepath.Glob(probePattern)
		for _, m := range matches {
			os.Remove(m)
		}
		return fmt.Errorf("no probe file found for page %d", page)
	}

	probeFile := probeMatches[0]
	f, err := os.Open(probeFile)
	if err != nil {
		os.Remove(probeFile)
		return fmt.Errorf("open probe file: %w", err)
	}
	cfg2, _, err := image.DecodeConfig(f)
	f.Close()
	if err != nil {
		os.Remove(probeFile)
		return fmt.Errorf("read probe dimensions: %w", err)
	}
	probeW, probeH := cfg2.Width, cfg2.Height
	os.Remove(probeFile)

	if probeW == 0 || probeH == 0 {
		return fmt.Errorf("probe produced zero-dimension image for page %d", page)
	}

	targetScale := computeTargetScale(probeW, probeH, cfg.Extract.OcrMaxWidth, cfg.Extract.OcrMaxHeight)

	finalPrefix := pageName(page)
	finalArgs := []string{"-png", "-scale-to", strconv.Itoa(targetScale),
		"-f", strconv.Itoa(page), "-l", strconv.Itoa(page),
		pdfPath, fmt.Sprintf("%s/%s", cfg.OutDir, finalPrefix)}

	finalCmd := exec.Command(cfg.Extract.PdftoppmCmd, finalArgs...)
	finalCmd.Stderr = os.Stderr

	if err := finalCmd.Run(); err != nil {
		return fmt.Errorf("pdftoppm failed for page %d: %w", page, err)
	}

	pattern := fmt.Sprintf("%s/%s-*.png", cfg.OutDir, finalPrefix)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob failed: %w", err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("no output file for page %d", page)
	}
	pngFile := matches[0]

	if err := os.Rename(pngFile, pngPath); err != nil {
		return fmt.Errorf("rename %s to %s: %w", pngFile, pngPath, err)
	}

	info, err := os.Stat(pngPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", pngPath, err)
	}
	if info.Size() == 0 {
		os.Remove(pngPath)
		return fmt.Errorf("page %d produced zero-byte output", page)
	}

	return nil
}

// computeTargetScale calculates the scale factor to fit probeW×probeH within maxWidth×maxHeight while preserving aspect ratio.
func computeTargetScale(probeW, probeH, maxWidth, maxHeight int) int {
	factorW := float64(maxWidth) / float64(probeW)
	factorH := float64(maxHeight) / float64(probeH)
	factor := math.Min(factorW, factorH)

	finalW := int(math.Round(float64(probeW) * factor))
	finalH := int(math.Round(float64(probeH) * factor))

	if probeW >= probeH {
		return finalW
	}
	return finalH
}
