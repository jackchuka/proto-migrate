package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/jackchuka/proto-migrate/internal/config"
	"github.com/jackchuka/proto-migrate/internal/loader"
	"github.com/jackchuka/proto-migrate/internal/resolve"
	"github.com/jackchuka/proto-migrate/internal/transform"
	"github.com/jackchuka/proto-migrate/internal/types"
	"github.com/jackchuka/proto-migrate/internal/vendor"
)

type Engine struct {
	config *config.Config
	flags  *types.GlobalFlags
	loader *loader.Loader
}

func New(cfg *config.Config, flags *types.GlobalFlags) *Engine {
	_ = flags.Concurrency
	if flags.Concurrency <= 0 {
		_ = runtime.NumCPU()
	}

	return &Engine{
		config: cfg,
		flags:  flags,
		loader: loader.New(cfg.Excludes),
	}
}

func (e *Engine) Plan(ctx context.Context) (*Plan, error) {
	files, err := e.loader.LoadDirectory(e.config.Source)
	if err != nil {
		return nil, fmt.Errorf("loading source directory: %w", err)
	}

	graph := resolve.NewGraph()
	for _, file := range files {
		graph.AddFile(file)
	}

	if err := graph.ResolveImports(e.config.Source); err != nil {
		return nil, fmt.Errorf("resolving imports: %w", err)
	}

	plan := &Plan{
		Changes:   make([]Change, 0),
		SourceDir: e.config.Source,
		TargetDir: e.config.Target,
		Files:     files,
		Graph:     graph,
	}

	// Apply user-defined rules first
	var appliedRules []transform.Rule
	for _, ruleConfig := range e.config.Rules {
		rule, err := transform.CreateRule(ruleConfig)
		if err != nil {
			return nil, fmt.Errorf("creating rule: %w", err)
		}
		appliedRules = append(appliedRules, rule)

		for _, file := range files {
			changed, err := rule.Apply(file)
			if err != nil {
				return nil, fmt.Errorf("applying rule %s to %s: %w", rule.ID(), file.Path, err)
			}
			if changed {
				plan.Changes = append(plan.Changes, Change{
					File:        file.Path,
					Type:        "transform",
					Description: fmt.Sprintf("Applied rule: %s", rule.ID()),
				})
			}
		}
	}

	// Generate and apply automatic import rules
	autoImportRules := transform.GenerateAutoImportRules(e.config, appliedRules)
	for _, rule := range autoImportRules {
		for _, file := range files {
			changed, err := rule.Apply(file)
			if err != nil {
				return nil, fmt.Errorf("applying auto-import rule %s to %s: %w", rule.ID(), file.Path, err)
			}
			if changed {
				plan.Changes = append(plan.Changes, Change{
					File:        file.Path,
					Type:        "auto-import",
					Description: fmt.Sprintf("Applied auto-rule: %s", rule.ID()),
				})
			}
		}
	}

	return plan, nil
}

func (e *Engine) Apply(ctx context.Context, plan *Plan) error {
	if e.flags.VendorDeps {
		v := vendor.New(e.config.Target)
		if err := v.VendorExternalDeps(plan.Graph); err != nil {
			return fmt.Errorf("vendoring dependencies: %w", err)
		}
	}

	tmpDir, err := os.MkdirTemp("", "proto-migrate-")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	for _, file := range plan.Files {
		relPath, err := filepath.Rel(plan.SourceDir, file.Path)
		if err != nil {
			return fmt.Errorf("calculating relative path: %w", err)
		}

		tmpPath := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(tmpPath), 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		if err := os.WriteFile(tmpPath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("writing temp file: %w", err)
		}
	}

	if err := os.MkdirAll(plan.TargetDir, 0755); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}

	for _, file := range plan.Files {
		relPath, err := filepath.Rel(plan.SourceDir, file.Path)
		if err != nil {
			return fmt.Errorf("calculating relative path: %w", err)
		}

		tmpPath := filepath.Join(tmpDir, relPath)
		targetPath := filepath.Join(plan.TargetDir, relPath)

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("creating target directory: %w", err)
		}

		if err := os.Rename(tmpPath, targetPath); err != nil {
			content, _ := os.ReadFile(tmpPath)
			if writeErr := os.WriteFile(targetPath, content, 0644); writeErr != nil {
				return fmt.Errorf("moving file: %w", err)
			}
		}
	}

	return nil
}

type Plan struct {
	Changes   []Change
	SourceDir string
	TargetDir string
	Files     []*loader.ProtoFile
	Graph     *resolve.Graph
}

type Change struct {
	File        string
	Type        string
	Description string
}

func (p *Plan) Print() error {
	fmt.Printf("\nPlan Summary:\n")
	fmt.Printf("  Source: %s\n", p.SourceDir)
	fmt.Printf("  Target: %s\n", p.TargetDir)
	fmt.Printf("  Files: %d\n", len(p.Files))
	fmt.Printf("  Changes: %d\n\n", len(p.Changes))

	if len(p.Changes) > 0 {
		fmt.Println("Changes to be applied:")
		for _, change := range p.Changes {
			relPath, _ := filepath.Rel(p.SourceDir, change.File)
			fmt.Printf("  • %s: %s\n", relPath, change.Description)
		}
	}

	return nil
}

func (p *Plan) PrintJSON() error {
	output := struct {
		Source  string   `json:"source"`
		Target  string   `json:"target"`
		Files   int      `json:"files"`
		Changes []Change `json:"changes"`
	}{
		Source:  p.SourceDir,
		Target:  p.TargetDir,
		Files:   len(p.Files),
		Changes: p.Changes,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func (p *Plan) Diff(w io.Writer) (bool, error) {
	hasDiffs := false

	for _, file := range p.Files {
		originalPath := file.Path
		originalContent, err := os.ReadFile(originalPath)
		if err != nil {
			originalContent = []byte{}
		}

		if !bytes.Equal(originalContent, []byte(file.Content)) {
			hasDiffs = true
			relPath, _ := filepath.Rel(p.SourceDir, file.Path)

			_, _ = color.New(color.Bold).Fprintf(w, "\n=== %s ===\n", relPath)
			printUnifiedDiff(w, string(originalContent), file.Content)
		}
	}

	return hasDiffs, nil
}

func printUnifiedDiff(w io.Writer, original, modified string) {
	originalLines := strings.Split(original, "\n")
	modifiedLines := strings.Split(modified, "\n")

	diffLines := computeDiff(originalLines, modifiedLines)

	for _, line := range diffLines {
		if strings.HasPrefix(line, "+") {
			_, _ = color.New(color.FgGreen).Fprintln(w, line)
		} else if strings.HasPrefix(line, "-") {
			_, _ = color.New(color.FgRed).Fprintln(w, line)
		} else if strings.HasPrefix(line, "@") {
			_, _ = color.New(color.FgCyan).Fprintln(w, line)
		} else {
			_, _ = fmt.Fprintln(w, line)
		}
	}
}

func computeDiff(original, modified []string) []string {
	var diff []string
	i, j := 0, 0

	for i < len(original) || j < len(modified) {
		if i >= len(original) {
			diff = append(diff, fmt.Sprintf("+%s", modified[j]))
			j++
		} else if j >= len(modified) {
			diff = append(diff, fmt.Sprintf("-%s", original[i]))
			i++
		} else if original[i] == modified[j] {
			if len(diff) > 0 && (strings.HasPrefix(diff[len(diff)-1], "-") || strings.HasPrefix(diff[len(diff)-1], "+")) {
				diff = append(diff, fmt.Sprintf(" %s", original[i]))
			}
			i++
			j++
		} else {
			diff = append(diff, fmt.Sprintf("-%s", original[i]))
			diff = append(diff, fmt.Sprintf("+%s", modified[j]))
			i++
			j++
		}
	}

	return diff
}
