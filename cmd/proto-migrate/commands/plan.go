package commands

import (
	"fmt"

	"github.com/jackchuka/proto-migrate/internal/config"
	"github.com/jackchuka/proto-migrate/internal/engine"
	"github.com/spf13/cobra"
)

func newPlanCommand() *cobra.Command {
	var jsonOpt bool

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Loads plan, prints summary (no writes)",
		Long:  "Analyzes the source directory and shows what changes would be made",
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

			if jsonOpt {
				return plan.PrintJSON()
			}
			return plan.Print()
		},
	}
	cmd.Flags().BoolVar(&jsonOpt, "json", false, "Plan format (default: plain text, use --json for JSON)")

	return cmd
}
