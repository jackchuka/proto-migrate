package commands

import (
	"fmt"
	"os"

	"github.com/jackchuka/proto-migrate/internal/config"
	"github.com/jackchuka/proto-migrate/internal/engine"
	"github.com/spf13/cobra"
)

func newDiffCommand() *cobra.Command {
	var exitCode bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Shows a unified diff of all pending rewrites",
		Long:  "Displays a colorized unified diff showing what changes would be made",
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

			hasDiffs, err := plan.Diff(os.Stdout)
			if err != nil {
				return fmt.Errorf("generating diff: %w", err)
			}

			if exitCode && hasDiffs {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&exitCode, "exit-code", false, "Exit with code 1 if there are differences")
	return cmd
}
