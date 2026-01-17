package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/suelio/millhouse/internal/builder"
	"github.com/suelio/millhouse/internal/display"
	"github.com/suelio/millhouse/internal/planner"
	"github.com/suelio/millhouse/internal/prd"
	"github.com/suelio/millhouse/internal/reviewer"
)

var runCmd = &cobra.Command{
	Use:   "run N",
	Short: "Execute N iterations autonomously",
	Long: `Run N iterations of the three-phase cycle:

Each iteration:
1. Planner selects an open PRD and creates an implementation plan
2. Builder executes the plan to implement the PRD
3. Reviewer verifies completion or updates plans for bailouts

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
		activePRDs := prdFile.GetActivePRDs()
		pendingPRDs := prdFile.GetPendingPRDs()

		if len(openPRDs) == 0 && len(activePRDs) == 0 && len(pendingPRDs) == 0 {
			d.Success("All PRDs complete! Nothing to do.")
			break
		}

		// ========================================
		// PHASE 1: PLANNER
		// ========================================
		if planner.ShouldRunPlanner(prdFile) {
			d.SubHeader("Phase 1: Planner")

			planResult, err := planner.Run(ctx, cwd, prdFile)
			if err != nil {
				d.Error(fmt.Sprintf("Planner error: %v", err))
				continue
			}

			if planResult.Skipped {
				d.Info(fmt.Sprintf("Planner skipped: %s", planResult.SkipReason))
			} else if planResult.PRDID != "" {
				d.Signal("PLAN_COMPLETE", planResult.PRDID)
			}

			// Handle planner signals
			for _, signal := range planResult.Signals {
				if signal.Type != "PLAN_COMPLETE" && signal.Type != "PLAN_SKIPPED" {
					d.Signal(signal.Type, signal.Details)
				}
			}

			// Reload PRD state after planner
			prdFile, err = prd.Load(cwd)
			if err != nil {
				return fmt.Errorf("failed to reload PRDs: %w", err)
			}
		} else if len(activePRDs) > 0 {
			d.Info(fmt.Sprintf("Planner skipped: active PRD exists (%s)", activePRDs[0].ID))
		} else if len(openPRDs) == 0 {
			d.Info("Planner skipped: no open PRDs")
		}

		// ========================================
		// PHASE 2: BUILDER
		// ========================================
		if builder.ShouldRunBuilder(prdFile) {
			d.SubHeader("Phase 2: Builder")

			activePRDs = prdFile.GetActivePRDs()
			if len(activePRDs) > 0 {
				d.Info(fmt.Sprintf("Executing plan for PRD: %s", activePRDs[0].ID))
			}

			buildResult, err := builder.Run(ctx, cwd, prdFile)
			if err != nil {
				d.Error(fmt.Sprintf("Builder error: %v", err))
			} else {
				// Handle builder signals
				for _, signal := range buildResult.Signals {
					d.Signal(signal.Type, signal.Details)
				}
			}

			// Reload PRD state after builder
			prdFile, err = prd.Load(cwd)
			if err != nil {
				return fmt.Errorf("failed to reload PRDs: %w", err)
			}
		} else {
			d.Info("Builder skipped: no active PRD")
		}

		// ========================================
		// PHASE 3: REVIEWER
		// ========================================
		if reviewer.ShouldRunReviewer(prdFile) {
			d.SubHeader("Phase 3: Reviewer")
			d.AnalysisStart()

			reviewResult, err := reviewer.Run(ctx, cwd, prdFile, i)
			if err != nil {
				d.Warning(fmt.Sprintf("Reviewer error: %v", err))
			} else {
				// Handle reviewer signals
				for _, id := range reviewResult.Verified {
					d.Signal("VERIFIED", id)
				}
				for _, id := range reviewResult.Rejected {
					d.Signal("REJECTED", id)
				}
				for _, id := range reviewResult.PlanUpdated {
					d.Signal("PLAN_UPDATED", id)
				}
				for _, id := range reviewResult.LoopRisk {
					d.Warning(fmt.Sprintf("Loop risk detected for PRD: %s", id))
				}
			}
		} else {
			d.Info("Reviewer skipped: no PRDs to review")
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
	active := prdFile.GetActivePRDs()
	pending := prdFile.GetPendingPRDs()
	complete := prdFile.GetCompletePRDs()

	d.SummaryExtended(len(open), len(active), len(pending), len(complete))

	if len(open) > 0 {
		d.Info(fmt.Sprintf("Open PRDs remaining: %d", len(open)))
	}
	if len(active) > 0 {
		d.Info(fmt.Sprintf("Active PRDs (with plans): %d", len(active)))
	}

	return nil
}
