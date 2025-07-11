package resolve

import (
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/jackchuka/proto-migrate/internal/loader"
)

type Graph struct {
	files    map[string]*loader.ProtoFile
	imports  map[string][]string
	external map[string]bool
}

func NewGraph() *Graph {
	return &Graph{
		files:    make(map[string]*loader.ProtoFile),
		imports:  make(map[string][]string),
		external: make(map[string]bool),
	}
}

func (g *Graph) AddFile(file *loader.ProtoFile) {
	g.files[file.Path] = file
	g.imports[file.Path] = g.extractImports(file)
}

func (g *Graph) extractImports(file *loader.ProtoFile) []string {
	var imports []string

	proto.Walk(file.Proto,
		proto.WithImport(func(i *proto.Import) {
			imports = append(imports, i.Filename)
		}),
	)

	return imports
}

func (g *Graph) ResolveImports(baseDir string) error {
	for path, imports := range g.imports {
		for _, imp := range imports {
			resolved := g.resolveImportPath(filepath.Dir(path), imp, baseDir)
			if _, exists := g.files[resolved]; !exists {
				g.external[imp] = true
			}
		}
	}
	return nil
}

func (g *Graph) resolveImportPath(currentDir, importPath, baseDir string) string {
	if filepath.IsAbs(importPath) {
		return importPath
	}

	candidates := []string{
		filepath.Join(currentDir, importPath),
		filepath.Join(baseDir, importPath),
		importPath,
	}

	for _, candidate := range candidates {
		if _, exists := g.files[candidate]; exists {
			return candidate
		}
	}

	return importPath
}

func (g *Graph) UpdateImports(relocations map[string]string) map[string][]ImportUpdate {
	updates := make(map[string][]ImportUpdate)

	for filePath, file := range g.files {
		var fileUpdates []ImportUpdate

		proto.Walk(file.Proto,
			proto.WithImport(func(i *proto.Import) {
				newPath := g.applyRelocations(i.Filename, relocations)
				if newPath != i.Filename {
					fileUpdates = append(fileUpdates, ImportUpdate{
						OldPath: i.Filename,
						NewPath: newPath,
					})
				}
			}),
		)

		if len(fileUpdates) > 0 {
			updates[filePath] = fileUpdates
		}
	}

	return updates
}

func (g *Graph) applyRelocations(importPath string, relocations map[string]string) string {
	for oldPrefix, newPrefix := range relocations {
		if strings.HasPrefix(importPath, oldPrefix) {
			return strings.Replace(importPath, oldPrefix, newPrefix, 1)
		}
	}
	return importPath
}

func (g *Graph) GetExternalImports() []string {
	var externals []string
	for imp := range g.external {
		externals = append(externals, imp)
	}
	return externals
}

func (g *Graph) GetFiles() map[string]*loader.ProtoFile {
	return g.files
}

type ImportUpdate struct {
	OldPath string
	NewPath string
}

type Relocations struct {
	PathMap map[string]string
}

func BuildRelocations(source, target string, rules []string) *Relocations {
	r := &Relocations{
		PathMap: make(map[string]string),
	}

	r.PathMap[source] = target

	for i := 0; i < len(rules)-1; i += 2 {
		r.PathMap[rules[i]] = rules[i+1]
	}

	return r
}
