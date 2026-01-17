package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/suelio/millhouse/internal/analyzer"
	"github.com/suelio/millhouse/internal/display"
	"github.com/suelio/millhouse/internal/executor"
	"github.com/suelio/millhouse/internal/prd"
)

var runCmd = &cobra.Command{
	Use:   "run N",
	Short: "Execute N iterations autonomously",
	Long: `Run N iterations of the executor-analyzer loop:

Each iteration:
1. Executor picks the highest priority open PRD
2. Executor implements the PRD and signals completion
3. Analyzer verifies pending PRDs and updates state

The loop continues until N iterations complete or no open PRDs remain.`,
	Args: cobra.ExactArgs(1),
	RunE: runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	iterations, err := strconv.Atoi(args[0])
	if err != nil || iterations < 1 {
		return fmt.Errorf("N must be a positive integer")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create display instance with color settings
	d := display.NewWithOptions(GetNoColor())

	if !prd.MillhouseExists(cwd) {
		d.Error(".millhouse/ directory not found")
		d.Info("Run 'mill init' to initialize")
		return fmt.Errorf("not initialized")
	}

	// Create context for the run
	ctx := context.Background()

	d.Header(fmt.Sprintf("Millhouse Run (%d iterations)", iterations))

	for i := 1; i <= iterations; i++ {
		d.IterationHeader(i, iterations)

		// Load fresh PRD state at start of each iteration
		prdFile, err := prd.Load(cwd)
		if err != nil {
			return fmt.Errorf("failed to load PRDs: %w", err)
		}

		// Check if there's work to do
		openPRDs := prdFile.GetOpenPRDs()
		pendingPRDs := prdFile.GetPendingPRDs()

		if len(openPRDs) == 0 && len(pendingPRDs) == 0 {
			d.Success("All PRDs complete! Nothing to do.")
			break
		}

		// Run pre-iteration analyzer if there are pending PRDs (skip first iteration)
		if i > 1 && analyzer.ShouldRunAnalyzer(prdFile) {
			d.SubHeader("Pre-iteration Analysis")
			d.AnalysisStart()
			analyzerResult, err := analyzer.Run(ctx, cwd, prdFile, i)
			if err != nil {
				d.Warning(fmt.Sprintf("Analyzer error: %v", err))
			} else {
				for _, id := range analyzerResult.Verified {
					d.Signal("VERIFIED", id)
				}
				for _, id := range analyzerResult.Rejected {
					d.Signal("REJECTED", id)
				}
				for _, id := range analyzerResult.LoopRisk {
					d.Warning(fmt.Sprintf("Loop risk detected for PRD: %s", id))
				}
			}

			// Reload PRD state after analyzer
			prdFile, err = prd.Load(cwd)
			if err != nil {
				return fmt.Errorf("failed to reload PRDs: %w", err)
			}
		}

		// Check if there are open PRDs to work on
		if len(openPRDs) == 0 {
			// No open PRDs, but might have pending ones
			if len(prdFile.GetPendingPRDs()) > 0 {
				d.Info("No open PRDs. Running analyzer to verify pending PRDs...")
				d.AnalysisStart()
				analyzerResult, err := analyzer.Run(ctx, cwd, prdFile, i)
				if err != nil {
					d.Warning(fmt.Sprintf("Analyzer error: %v", err))
				} else {
					for _, id := range analyzerResult.Verified {
						d.Signal("VERIFIED", id)
					}
					for _, id := range analyzerResult.Rejected {
						d.Signal("REJECTED", id)
					}
				}
				continue
			}
			d.Success("All PRDs complete!")
			break
		}

		// Run executor - it will choose which PRD to work on
		d.SubHeader("Executor")
		d.Info("Executor will choose from open PRDs...")

		execResult, err := executor.RunExecutor(ctx, cwd, prdFile)
		if err != nil {
			d.Error(fmt.Sprintf("Executor error: %v", err))
			continue
		}

		// Handle executor signals
		for _, signal := range execResult.Signals {
			d.Signal(signal.Type, signal.Details)
		}

		// Run post-iteration analyzer
		prdFile, err = prd.Load(cwd)
		if err != nil {
			return fmt.Errorf("failed to reload PRDs: %w", err)
		}

		if analyzer.ShouldRunAnalyzer(prdFile) {
			d.SubHeader("Post-iteration Analysis")
			d.AnalysisStart()
			analyzerResult, err := analyzer.Run(ctx, cwd, prdFile, i)
			if err != nil {
				d.Warning(fmt.Sprintf("Analyzer error: %v", err))
			} else {
				for _, id := range analyzerResult.Verified {
					d.Signal("VERIFIED", id)
				}
				for _, id := range analyzerResult.Rejected {
					d.Signal("REJECTED", id)
				}
			}
		}

		d.Divider()
	}

	// Final status
	d.Header("Final Status")
	prdFile, err := prd.Load(cwd)
	if err != nil {
		return fmt.Errorf("failed to load final PRD state: %w", err)
	}

	open := prdFile.GetOpenPRDs()
	pending := prdFile.GetPendingPRDs()
	complete := prdFile.GetCompletePRDs()

	d.Summary(len(open), len(pending), len(complete))

	if len(open) > 0 {
		d.Info(fmt.Sprintf("Open PRDs remaining: %d", len(open)))
	}

	return nil
}
