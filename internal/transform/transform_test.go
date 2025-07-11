package transform

import (
	"strings"
	"testing"

	"github.com/emicklei/proto"
	"github.com/jackchuka/proto-migrate/internal/loader"
)

func TestPackageRule(t *testing.T) {
	rule := &PackageRule{From: "old.v1", To: "new.v1"}

	content := `syntax = "proto3";

package old.v1;

message Test {
  string id = 1;
}`

	file := &loader.ProtoFile{
		Path:    "test.proto",
		Content: content,
		Proto:   parseProto(t, content),
	}

	changed, err := rule.Apply(file)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !changed {
		t.Error("Expected file to be changed")
	}
	if !strings.Contains(file.Content, "package new.v1;") {
		t.Error("Package was not renamed")
	}
}

func TestServiceRule(t *testing.T) {
	rule := &ServiceRule{From: "OldService", To: "NewService"}

	content := `syntax = "proto3";

package test.v1;

service OldService {
  rpc GetItem(GetRequest) returns (GetResponse);
}`

	file := &loader.ProtoFile{
		Path:    "test.proto",
		Content: content,
		Proto:   parseProto(t, content),
	}

	changed, err := rule.Apply(file)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !changed {
		t.Error("Expected file to be changed")
	}
	if !strings.Contains(file.Content, "service NewService {") {
		t.Error("Service was not renamed")
	}
}

func TestImportRule(t *testing.T) {
	rule := &ImportRule{From: "old/v1", To: "new/v1"}

	content := `syntax = "proto3";

package test.v1;

import "old/v1/types.proto";

message Test {
  string id = 1;
}`

	file := &loader.ProtoFile{
		Path:    "test.proto",
		Content: content,
		Proto:   parseProto(t, content),
	}

	changed, err := rule.Apply(file)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !changed {
		t.Error("Expected file to be changed")
	}
	if !strings.Contains(file.Content, `import "new/v1/types.proto";`) {
		t.Error("Import was not updated")
	}
}

func TestRegexpRule(t *testing.T) {
	rule := &RegexpRule{Pattern: `old\.v1`, Replace: "new.v1"}

	content := `syntax = "proto3";

package test.v1;

// Reference to old.v1.Service
message Test {
  string id = 1; // old.v1 field
}`

	file := &loader.ProtoFile{
		Path:    "test.proto",
		Content: content,
		Proto:   parseProto(t, content),
	}

	changed, err := rule.Apply(file)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !changed {
		t.Error("Expected file to be changed")
	}
	if strings.Contains(file.Content, "old.v1") {
		t.Error("Pattern was not replaced")
	}
	if !strings.Contains(file.Content, "new.v1") {
		t.Error("Replacement not found")
	}
}

func parseProto(t *testing.T, content string) *proto.Proto {
	t.Helper()
	reader := strings.NewReader(content)
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse proto: %v", err)
	}
	return definition
}
