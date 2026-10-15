package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/Galdoba/lllm-ocr/internal/config"
	"github.com/Galdoba/lllm-ocr/layoutblock"
	"github.com/Galdoba/lllm-ocr/ocr"
)

// NewOCRCommand returns the "ocr" CLI command.
func NewOCRCommand() *cli.Command {
	return &cli.Command{
		Name:  "ocr",
		Usage: "Recognize text from extracted PNG pages and produce HTML sources",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Required: true,
				Usage:    "Path to input PDF file (used for page count filtering)",
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
				Usage:   "Pages or ranges to recognize, comma-separated (e.g. 1-3,5,6-8)",
			},
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Value:   ".",
				Usage:   "Working directory containing PNG files",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "Visualize recognized blocks on each page (outputs _page_NNN.png)",
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

			if err := config.ValidatePipeline(ocrCfg); err != nil {
				return fmt.Errorf("invalid pipeline config: %w", err)
			}

			fmt.Println("Starting OCR (recognition) pipeline...")
			if err := ocr.RecognizePages(ocrCfg, pages); err != nil {
				fmt.Fprintf(os.Stderr, "OCR recognition failed: %v\n", err)
				return err
			}

			if cCtx.Bool("debug") {
				parser := layoutblock.NewBlockParser(true)
				for _, p := range pages {
					pn := config.PageName(p)
					htmlPath := filepath.Join(workspacePath, pn+".html")
					pngPath := filepath.Join(workspacePath, pn+".png")
					debugPath := filepath.Join(workspacePath, "_"+pn+".png")

					if _, err := os.Stat(pngPath); os.IsNotExist(err) {
						fmt.Fprintf(os.Stderr, "debug: skipping page %d (%s not found)\n", p, pn+".png")
						continue
					}

					htmlContent, err := os.ReadFile(htmlPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "debug: skipping page %d (cannot read %s: %v)\n", p, pn+".html", err)
						continue
					}

					blocks, err := parser.Parse(string(htmlContent))
					if err != nil {
						fmt.Fprintf(os.Stderr, "debug: skipping page %d (parse error: %v)\n", p, err)
						continue
					}

					if err := layoutblock.VisualizeBlocks(pngPath, blocks, debugPath); err != nil {
						fmt.Fprintf(os.Stderr, "debug: page %d visualization failed: %v\n", p, err)
						continue
					}

					fmt.Printf("debug: page %d -> %s\n", p, debugPath)
				}
			}

			pageSources := filepath.Join(workspacePath, "pageSources")
			if err := os.MkdirAll(pageSources, 0755); err != nil {
				return fmt.Errorf("creating pageSources directory: %w", err)
			}

			for _, p := range pages {
				pn := config.PageName(p)
				htmlPath := filepath.Join(workspacePath, pn+".html")
				if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
					continue
				}
				if err := os.Rename(htmlPath, filepath.Join(pageSources, pn+".html")); err != nil {
					return fmt.Errorf("moving %s to pageSources: %w", pn+".html", err)
				}
			}

			fmt.Println("\nOCR complete.")
			fmt.Println("-> HTML sources saved in: " + pageSources)
			fmt.Println("-> Run 'assemble' to build combined.html, then 'convert'.")
			return nil
		},
	}
}
