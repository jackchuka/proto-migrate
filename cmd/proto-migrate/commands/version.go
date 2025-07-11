package commands

import (
	"fmt"

	"github.com/jackchuka/proto-migrate/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version.Info())
			return nil
		},
	}
}
