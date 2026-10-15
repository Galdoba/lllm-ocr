package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	commands "github.com/Galdoba/lllm-ocr/internal/commands"
)

func main() {
	app := &cli.App{
		Name:     "lllm-ocr",
		Usage:    "PDF → OCR → structured blocks → Markdown/Text/HTML/JSON",
		Commands: commands.GetCommands(),
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
