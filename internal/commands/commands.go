package commands

import "github.com/urfave/cli/v2"

// GetCommands returns all CLI subcommands.
func GetCommands() []*cli.Command {
	return []*cli.Command{
		NewConfigCommand(),
		NewExtractCommand(),
		NewOCRCommand(),
		NewAssembleCommand(),
		NewConvertCommand(),
	}
}
