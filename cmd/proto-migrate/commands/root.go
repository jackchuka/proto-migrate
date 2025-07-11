package commands

import (
	"github.com/jackchuka/proto-migrate/internal/types"
	"github.com/spf13/cobra"
)

var globalFlags types.GlobalFlags

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proto-migrate",
		Short: "Protocol Buffer migrations done right - AST-aware, safe, and blazing fast",
		Long: `proto-migrate rewrites package names, imports, service names, language-specific
options, and more—then validates that the resulting graph still compiles and
is backward-compatible.`,
		SilenceUsage: true,
	}

	cmd.PersistentFlags().StringVar(&globalFlags.Config, "config", "", "Path to proto-migrate.yaml (default: auto-detect)")
	cmd.PersistentFlags().BoolVar(&globalFlags.VendorDeps, "vendor-deps", false, "Copy missing externals to vendor/")
	cmd.PersistentFlags().IntVar(&globalFlags.Concurrency, "concurrency", 0, "Parallel file visits (default: #CPU)")

	cmd.AddCommand(
		newInitCommand(),
		newPlanCommand(),
		newDiffCommand(),
		newApplyCommand(),
		newVersionCommand(),
	)

	return cmd
}

func GetGlobalFlags() *types.GlobalFlags {
	return &globalFlags
}
