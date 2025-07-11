package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Scaffold a migration plan",
		Long:  "Generates a sample .proto-sync.yaml configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := `# .proto-sync.yaml
source: proto/oldpackage/v1
target: proto/newpackage/v1

excludes:
  - "*ignore*.proto"
  - "*private*.proto"

rules:
  - kind: package
    from: oldpackage.v1
    to: newpackage.v1

  - kind: service
    from: OldService
    to: NewService

  - kind: regexp
    pattern: "oldpackage\\.v1\\."
    replace: "newpackage.v1."
`
			fmt.Print(config)
			return nil
		},
	}
}
