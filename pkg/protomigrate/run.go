package protosync

import (
	"context"
	"fmt"

	"github.com/jackchuka/proto-migrate/internal/config"
	"github.com/jackchuka/proto-migrate/internal/engine"
	"github.com/jackchuka/proto-migrate/internal/types"
)

type Options struct {
	Config        string
	ProtoPath     []string
	VendorDeps    bool
	UpdateOptions bool
	LogJSON       bool
	Concurrency   int
	DryRun        bool
}

func Run(ctx context.Context, opts Options) error {
	cfg, err := config.Load(opts.Config)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	flags := &types.GlobalFlags{
		Config:      opts.Config,
		VendorDeps:  opts.VendorDeps,
		Concurrency: opts.Concurrency,
	}

	eng := engine.New(cfg, flags)

	plan, err := eng.Plan(ctx)
	if err != nil {
		return fmt.Errorf("planning: %w", err)
	}

	if opts.DryRun {
		return plan.Print()
	}

	if err := eng.Apply(ctx, plan); err != nil {
		return fmt.Errorf("applying: %w", err)
	}

	return nil
}
