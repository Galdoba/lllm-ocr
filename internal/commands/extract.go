package commands

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/Galdoba/lllm-ocr/internal/config"
	"github.com/Galdoba/lllm-ocr/ocr"
)

// NewExtractCommand returns the "extract" CLI command.
func NewExtractCommand() *cli.Command {
	return &cli.Command{
		Name:  "extract",
		Usage: "Extract pages from a PDF into PNG images",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Required: true,
				Usage:    "Path to input PDF file",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "",
				Usage:   "Path to config file (default: ~/.config/llllm-ocr/config.json)",
			},
			&cli.StringFlag{
				Name:    "pages",
				Aliases: []string{"p"},
				Value:   "1-999",
				Usage:   "Pages or ranges to extract, comma-separated (e.g. 1-3,5,6-8)",
			},
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Value:   ".",
				Usage:   "Working directory for PNG output",
			},
		},
		Action: func(cCtx *cli.Context) error {
			pdfPath := cCtx.String("input")
			configPath := cCtx.String("config")
			pagesSpec := cCtx.String("pages")
			outDir := cCtx.String("dir")

			if configPath == "" {
				configPath = config.GetDefaultConfigPath()
			}

			if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
				return fmt.Errorf("PDF file '%s' does not exist", pdfPath)
			}

			pageCount, err := ocr.PdfPageCount(pdfPath)
			if err != nil {
				return err
			}

			pages, err := ocr.SelectPages(pagesSpec, pageCount)
			if err != nil {
				return err
			}

			if configPath == config.GetDefaultConfigPath() {
				if err := config.EnsureDefaultConfigExists(configPath); err != nil {
					return fmt.Errorf("ensuring default config exists: %w", err)
				}
			}

			cliCfg, err := config.LoadConfig(configPath)
			if err != nil {
				return err
			}

			workspacePath := config.WorkspacePath(pdfPath, outDir)
			if err := os.MkdirAll(workspacePath, 0755); err != nil {
				return fmt.Errorf("creating workspace directory: %w", err)
			}

			ocrCfg := config.BuildOCRConfig(cliCfg, workspacePath)

			fmt.Println("Starting extract pipeline...")
			extracted, err := ocr.ExtractPages(ocrCfg, pdfPath, pages)
			if err != nil {
				return fmt.Errorf("extract failed: %w", err)
			}
			if len(extracted) == 0 {
				return nil
			}

			fmt.Println("\nExtract complete.")
			fmt.Println("-> PNG files saved in: " + workspacePath)
			fmt.Println("-> Run 'ocr' to recognize text, then 'assemble' to build combined.html.")
			return nil
		},
	}
}
