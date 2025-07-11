package commands

import (
	"fmt"

	"github.com/jackchuka/proto-migrate/internal/config"
	"github.com/jackchuka/proto-migrate/internal/engine"
	"github.com/spf13/cobra"
)

func newApplyCommand() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Executes the plan, rewrites files atomically",
		Long:  "Applies all transformations to the proto files and ensures atomicity",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			flags := GetGlobalFlags()

			cfg, err := config.Load(flags.Config)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			eng := engine.New(cfg, flags)

			plan, err := eng.Plan(ctx)
			if err != nil {
				return fmt.Errorf("planning: %w", err)
			}

			if dryRun {
				fmt.Println("Dry run mode - no changes will be made")
				return plan.Print()
			}

			if err := eng.Apply(ctx, plan); err != nil {
				return fmt.Errorf("applying changes: %w", err)
			}

			fmt.Println("Changes applied successfully")
			return nil
		},
	}
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	return cmd
}
