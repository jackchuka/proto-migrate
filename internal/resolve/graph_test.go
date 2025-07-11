package resolve

import (
	"testing"

	"github.com/emicklei/proto"
	"github.com/jackchuka/proto-migrate/internal/loader"
)

func TestGraphAddFile(t *testing.T) {
	g := NewGraph()

	file := &loader.ProtoFile{
		Path: "test.proto",
		Proto: &proto.Proto{
			Elements: []proto.Visitee{
				&proto.Import{Filename: "other.proto"},
			},
		},
	}

	g.AddFile(file)

	if len(g.files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(g.files))
	}

	imports := g.imports[file.Path]
	if len(imports) != 1 || imports[0] != "other.proto" {
		t.Errorf("Expected import 'other.proto', got %v", imports)
	}
}

func TestUpdateImports(t *testing.T) {
	g := NewGraph()

	file := &loader.ProtoFile{
		Path: "test.proto",
		Proto: &proto.Proto{
			Elements: []proto.Visitee{
				&proto.Import{Filename: "old/v1/types.proto"},
			},
		},
	}

	g.AddFile(file)

	relocations := map[string]string{
		"old/v1": "new/v1",
	}

	updates := g.UpdateImports(relocations)

	if len(updates) != 1 {
		t.Fatalf("Expected 1 file with updates, got %d", len(updates))
	}

	fileUpdates := updates[file.Path]
	if len(fileUpdates) != 1 {
		t.Fatalf("Expected 1 update, got %d", len(fileUpdates))
	}

	update := fileUpdates[0]
	if update.OldPath != "old/v1/types.proto" {
		t.Errorf("Expected old path 'old/v1/types.proto', got %s", update.OldPath)
	}
	if update.NewPath != "new/v1/types.proto" {
		t.Errorf("Expected new path 'new/v1/types.proto', got %s", update.NewPath)
	}
}

func TestBuildRelocations(t *testing.T) {
	r := BuildRelocations("proto/old", "proto/new", []string{"old/v1", "new/v1", "old/v2", "new/v2"})

	expected := map[string]string{
		"proto/old": "proto/new",
		"old/v1":    "new/v1",
		"old/v2":    "new/v2",
	}

	if len(r.PathMap) != len(expected) {
		t.Errorf("Expected %d relocations, got %d", len(expected), len(r.PathMap))
	}

	for k, v := range expected {
		if r.PathMap[k] != v {
			t.Errorf("Expected %s -> %s, got %s", k, v, r.PathMap[k])
		}
	}
}
