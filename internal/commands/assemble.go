package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/Galdoba/lllm-ocr/internal/config"
	"github.com/Galdoba/lllm-ocr/ocr"
)

// NewAssembleCommand returns the "assemble" CLI command.
func NewAssembleCommand() *cli.Command {
	return &cli.Command{
		Name:  "assemble",
		Usage: "Build combined.html from pageSources/ (skips pages with no recognized blocks)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Value:    "",
				Usage:    "Path to source PDF (for auto-deriving workspace directory)",
			},
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Value:   ".",
				Usage:   "Working directory containing pageSources/",
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
				Usage:   "Pages or ranges to include, comma-separated (e.g. 1-3,5,6-8)",
			},
		},
		Action: func(cCtx *cli.Context) error {
			configPath := cCtx.String("config")
			explicitDir := cCtx.String("dir")
			inputPath := cCtx.String("input")
			pagesSpec := cCtx.String("pages")

			if configPath == "" {
				configPath = config.GetDefaultConfigPath()
			}

			var workspacePath string
			if explicitDir != "" && explicitDir != "." {
				workspacePath = explicitDir
			} else if inputPath != "" {
				workspacePath = config.WorkspacePath(inputPath, ".")
			} else {
				workspacePath = "."
			}

			if err := os.MkdirAll(workspacePath, 0755); err != nil {
				return fmt.Errorf("creating workspace directory: %w", err)
			}

			if configPath == config.GetDefaultConfigPath() {
				if err := config.EnsureDefaultConfigExists(configPath); err != nil {
					return fmt.Errorf("ensuring default config exists: %w", err)
				}
			}

			pages, err := ocr.ParsePageSpec(pagesSpec)
			if err != nil {
				return err
			}

			cliCfg, err := config.LoadConfig(configPath)
			if err != nil {
				return err
			}

			ocrCfg := config.BuildOCRConfig(cliCfg, workspacePath)

			report, err := ocr.Assemble(ocrCfg, pages)
			if err != nil {
				return fmt.Errorf("assembling combined.html: %w", err)
			}

			fmt.Println("Assemble complete.")
			fmt.Printf("-> Combined file saved as: %s\n", filepath.Join(workspacePath, "combined.html"))
			fmt.Printf("-> Pages processed: %d, skipped: %d\n", report.Processed, report.Skipped)
			if report.TotalPages == 0 {
				fmt.Println("-> No page sources found in pageSources/.")
			}
			fmt.Println("-> Convert with: convert -i combined.html -f markdown")
			return nil
		},
	}
}
