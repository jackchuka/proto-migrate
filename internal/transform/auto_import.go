package transform

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/jackchuka/proto-migrate/internal/config"
	"github.com/jackchuka/proto-migrate/internal/loader"
)

type AutoImportRule struct {
	SourceDir    string
	TargetDir    string
	PackageRules []PackageRule
}

func NewAutoImportRule(sourceDir, targetDir string, packageRules []PackageRule) *AutoImportRule {
	return &AutoImportRule{
		SourceDir:    sourceDir,
		TargetDir:    targetDir,
		PackageRules: packageRules,
	}
}

func (r *AutoImportRule) ID() string {
	return fmt.Sprintf("auto-import:%s->%s", r.SourceDir, r.TargetDir)
}

func (r *AutoImportRule) Apply(file *loader.ProtoFile) (bool, error) {
	changed := false
	newContent := file.Content

	// Extract directory mappings from source/target and package rules
	dirMappings := r.buildDirectoryMappings()

	proto.Walk(file.Proto,
		proto.WithImport(func(i *proto.Import) {
			newPath := r.transformImportPath(i.Filename, dirMappings)
			if newPath != i.Filename {
				oldImport := fmt.Sprintf(`import "%s";`, i.Filename)
				newImport := fmt.Sprintf(`import "%s";`, newPath)

				// Check if transformation has already been applied
				if !strings.Contains(newContent, newImport) {
					newContent = strings.Replace(newContent, oldImport, newImport, 1)
					changed = true
				}
			}
		}),
	)

	if changed {
		file.Content = newContent
		return true, nil
	}
	return false, nil
}

func (r *AutoImportRule) buildDirectoryMappings() map[string]string {
	mappings := make(map[string]string)

	// Add source/target directory mapping
	if r.SourceDir != "" && r.TargetDir != "" {
		mappings[r.SourceDir] = r.TargetDir
	}

	// Add package-based directory mappings
	for _, pkgRule := range r.PackageRules {
		// Convert package names to directory paths
		fromDir := strings.ReplaceAll(pkgRule.From, ".", "/")
		toDir := strings.ReplaceAll(pkgRule.To, ".", "/")
		mappings[fromDir] = toDir

		// Also handle variations with common prefixes
		if r.SourceDir != "" && r.TargetDir != "" {
			sourcePath := filepath.Join(r.SourceDir, fromDir)
			targetPath := filepath.Join(r.TargetDir, toDir)
			mappings[sourcePath] = targetPath
		}
	}

	return mappings
}

func (r *AutoImportRule) transformImportPath(importPath string, mappings map[string]string) string {
	// Try exact matches first
	for from, to := range mappings {
		if importPath == from {
			return to
		}
	}

	// Try prefix matches (longest first)
	longestMatch := ""
	var replacement string

	for from, to := range mappings {
		if strings.HasPrefix(importPath, from+"/") && len(from) > len(longestMatch) {
			longestMatch = from
			replacement = to
		}
	}

	if longestMatch != "" {
		return strings.Replace(importPath, longestMatch, replacement, 1)
	}

	// Try directory name matches for common patterns
	for from, to := range mappings {
		if strings.Contains(importPath, from) {
			return strings.ReplaceAll(importPath, from, to)
		}
	}

	return importPath
}

// GenerateAutoImportRules creates automatic import rules based on config
func GenerateAutoImportRules(cfg *config.Config, rules []Rule) []Rule {
	var packageRules []PackageRule

	// Extract package rules from the rule set
	for _, rule := range rules {
		if pkgRule, ok := rule.(*PackageRule); ok {
			packageRules = append(packageRules, *pkgRule)
		}
	}

	// Create auto-import rule
	autoRule := NewAutoImportRule(cfg.Source, cfg.Target, packageRules)

	return []Rule{autoRule}
}
