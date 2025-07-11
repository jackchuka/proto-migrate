package vendor

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackchuka/proto-migrate/internal/resolve"
)

type Vendorer struct {
	targetDir string
	vendorDir string
}

func New(targetDir string) *Vendorer {
	return &Vendorer{
		targetDir: targetDir,
		vendorDir: filepath.Join(targetDir, "vendor"),
	}
}

func (v *Vendorer) VendorExternalDeps(graph *resolve.Graph) error {
	externals := graph.GetExternalImports()
	if len(externals) == 0 {
		return nil
	}

	if err := os.MkdirAll(v.vendorDir, 0755); err != nil {
		return fmt.Errorf("creating vendor directory: %w", err)
	}

	for _, imp := range externals {
		if err := v.vendorFile(imp); err != nil {
			return fmt.Errorf("vendoring %s: %w", imp, err)
		}
	}

	return nil
}

func (v *Vendorer) vendorFile(importPath string) error {
	destPath := filepath.Join(v.vendorDir, importPath)

	if _, err := os.Stat(destPath); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	content, err := v.fetchProto(importPath)
	if err != nil {
		return fmt.Errorf("fetching proto: %w", err)
	}

	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func (v *Vendorer) fetchProto(importPath string) ([]byte, error) {
	wellKnown := map[string]string{
		"google/protobuf/timestamp.proto":  "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/timestamp.proto",
		"google/protobuf/duration.proto":   "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/duration.proto",
		"google/protobuf/empty.proto":      "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/empty.proto",
		"google/protobuf/any.proto":        "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/any.proto",
		"google/protobuf/struct.proto":     "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/struct.proto",
		"google/protobuf/wrappers.proto":   "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/wrappers.proto",
		"google/protobuf/field_mask.proto": "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/field_mask.proto",
		"google/api/annotations.proto":     "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto",
		"google/api/http.proto":            "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto",
	}

	url, ok := wellKnown[importPath]
	if !ok {
		if strings.HasPrefix(importPath, "google/") {
			url = fmt.Sprintf("https://raw.githubusercontent.com/googleapis/googleapis/master/%s", importPath)
		} else {
			return nil, fmt.Errorf("unknown import path: %s", importPath)
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching from %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return content, nil
}
