package cli

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var noColor bool

var rootCmd = &cobra.Command{
	Use:   "mil",
	Short: "Millhouse - Autonomous PRD-driven development",
	Long: `Millhouse is a CLI tool for autonomous PRD-driven development.

It manages a list of Product Requirements Documents (PRDs) and uses
Claude to implement them iteratively, with an analyzer agent ensuring
quality and progress between iterations.

Commands:
  init     Create .milhouse/ folder with starter files
  discuss  Interactive Claude session for PRD management
  status   Show PRD status summary
  run N    Execute N iterations autonomously`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Disable colors globally if --no-color flag is set
		if noColor {
			color.NoColor = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

// GetNoColor returns the no-color flag value
func GetNoColor() bool {
	return noColor
}

func Execute() error {
	return rootCmd.Execute()
}
