package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/Galdoba/lllm-ocr/internal/config"
	"github.com/Galdoba/lllm-ocr/layoutblock"
	"github.com/Galdoba/lllm-ocr/output"
)

// NewConvertCommand returns the "convert" CLI command.
func NewConvertCommand() *cli.Command {
	return &cli.Command{
		Name:  "convert",
		Usage: "Convert combined.html to Markdown",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Required: true,
				Usage:    "Path to input HTML file",
			},
			&cli.StringFlag{
				Name:     "format",
				Aliases:  []string{"f"},
				Required: true,
				Usage:    "Output format: markdown",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "Output file path (default: stdout)",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "",
				Usage:   "Path to config file (default: ~/.config/llllm-ocr/config.json)",
			},
			&cli.BoolFlag{
				Name:  "lenient",
				Value: false,
				Usage: "Skip unsupported labels during parsing",
			},
		},
		Action: func(cCtx *cli.Context) error {
			inputPath := cCtx.String("input")
			formatStr := cCtx.String("format")
			outputPath := cCtx.String("output")
			configPath := cCtx.String("config")
			lenient := cCtx.Bool("lenient")

			if outputPath == "" {
				outputPath = filepath.Join(filepath.Dir(inputPath), "output.md")
			}

			if configPath == "" {
				configPath = config.GetDefaultConfigPath()
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

			htmlContent, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("reading input file: %w", err)
			}

			parser := layoutblock.NewBlockParser(lenient)
			blocks, err := parser.Parse(string(htmlContent))
			if err != nil {
				return fmt.Errorf("parsing HTML: %w", err)
			}

			filterCfg := layoutblock.FilterConfig{
				RemoveLabels:    cliCfg.RemoveBlocks,
				KeepPageNumbers: false,
				KeepThinkTags:   false,
			}
			blocks = layoutblock.FilterBlocks(blocks, filterCfg)

			blocks, err = layoutblock.MergeCrossPageBlocks(blocks)
			if err != nil {
				return fmt.Errorf("merging blocks: %w", err)
			}

			layoutblock.NormalizeBlockLineBreaks(blocks)

			if formatStr != "markdown" {
				return fmt.Errorf("unsupported format: %s (only markdown is supported)", formatStr)
			}

			out := output.NewMarkdownOutput(output.MarkdownConfig{
				PageSeparator: "",
			})

			result := out.Convert(blocks)

			if outputPath != "" {
				dir := filepath.Dir(outputPath)
				if dir != "" && dir != "." {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return fmt.Errorf("creating output directory: %w", err)
					}
				}
				if err := os.WriteFile(outputPath, []byte(result), 0644); err != nil {
					return fmt.Errorf("writing output file: %w", err)
				}
				fmt.Printf("Output written to: %s\n", outputPath)
			} else {
				fmt.Print(result)
			}

			return nil
		},
	}
}
