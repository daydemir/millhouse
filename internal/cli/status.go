package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/daydemir/milhouse/internal/display"
	"github.com/daydemir/milhouse/internal/prd"
)

var verboseFlag bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show PRD status summary",
	Long:  `Display the current status of all PRDs, grouped by status (open, pending, complete).`,
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show full PRD details")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !prd.MillhouseExists(cwd) {
		display.Error(".milhouse/ directory not found")
		display.Info("Run 'mil init' to initialize")
		return fmt.Errorf("not initialized")
	}

	prdFile, err := prd.Load(cwd)
	if err != nil {
		return fmt.Errorf("failed to load PRDs: %w", err)
	}

	open := prdFile.GetOpenPRDs()
	pending := prdFile.GetPendingPRDs()
	complete := prdFile.GetCompletePRDs()

	// Sort each by priority
	sort.Slice(open, func(i, j int) bool { return open[i].Priority < open[j].Priority })
	sort.Slice(pending, func(i, j int) bool { return pending[i].Priority < pending[j].Priority })
	sort.Slice(complete, func(i, j int) bool { return complete[i].Priority < complete[j].Priority })

	if len(prdFile.PRDs) == 0 {
		display.Info("No PRDs defined yet")
		display.Info("Run 'mil discuss' to add PRDs")
		return nil
	}

	// Verbose mode shows full details
	if verboseFlag {
		display.Header("Millhouse Status")

		// Show open PRDs
		if len(open) > 0 {
			display.SubHeader(fmt.Sprintf("Open (%d)", len(open)))
			for _, p := range open {
				display.PRDStatus(p)
			}
		}

		// Show pending PRDs
		if len(pending) > 0 {
			display.SubHeader(fmt.Sprintf("Pending Verification (%d)", len(pending)))
			for _, p := range pending {
				display.PRDStatus(p)
			}
		}

		// Show complete PRDs
		if len(complete) > 0 {
			display.SubHeader(fmt.Sprintf("Complete (%d)", len(complete)))
			for _, p := range complete {
				display.PRDStatus(p)
			}
		}

		// Summary
		display.Summary(len(open), len(pending), len(complete))
	} else {
		// Compact mode (default)
		display.SummaryCompact(len(open), len(pending), len(complete))

		// Show open PRDs (compact)
		if len(open) > 0 {
			fmt.Println("\nOpen:")
			maxShow := 20
			for i, p := range open {
				if i >= maxShow {
					fmt.Printf("  + %d more...\n", len(open)-maxShow)
					break
				}
				display.PRDStatusCompact(p)
			}
		}

		// Show complete PRDs (compact)
		if len(complete) > 0 {
			fmt.Println("\nComplete:")
			maxShow := 2
			for i, p := range complete {
				if i >= maxShow {
					fmt.Printf("  + %d more...\n", len(complete)-maxShow)
					break
				}
				display.PRDStatusCompact(p)
			}
		}
	}

	// Next action hint
	if len(open) > 0 || len(pending) > 0 {
		display.Info("Run 'mil run N' to execute N iterations")
	} else if len(complete) > 0 && len(open) == 0 && len(pending) == 0 {
		display.Success("All PRDs complete!")
	}

	return nil
}
