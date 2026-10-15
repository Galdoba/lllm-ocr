package commands

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/Galdoba/lllm-ocr/internal/config"
)

// NewConfigCommand returns the "new-config" CLI command.
func NewConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "new-config",
		Usage: "Copy default config to current directory for local editing",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "config.json",
				Usage:   "Output file path for the copied config",
			},
		},
		Action: func(cCtx *cli.Context) error {
			defaultPath := config.GetDefaultConfigPath()
			if err := config.EnsureDefaultConfigExists(defaultPath); err != nil {
				return fmt.Errorf("ensuring default config exists: %w", err)
			}

			mainData, err := os.ReadFile(defaultPath)
			if err != nil {
				return fmt.Errorf("reading default config: %w", err)
			}

			outputPath := cCtx.String("output")
			if err := os.WriteFile(outputPath, mainData, 0644); err != nil {
				return fmt.Errorf("writing config to %s: %w", outputPath, err)
			}

			fmt.Printf("Config copied to: %s\n", outputPath)
			fmt.Printf("Use with: -c %s\n", outputPath)
			return nil
		},
	}
}
