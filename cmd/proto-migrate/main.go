package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackchuka/proto-migrate/cmd/proto-migrate/commands"
)

func main() {
	ctx := context.Background()
	rootCmd := commands.NewRootCommand()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
