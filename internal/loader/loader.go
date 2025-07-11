package loader

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/emicklei/proto"
)

type Loader struct {
	excludes []string
	mu       sync.Mutex
	cache    map[string]*proto.Proto
}

func New(excludes []string) *Loader {
	return &Loader{
		excludes: excludes,
		cache:    make(map[string]*proto.Proto),
	}
}

func (l *Loader) LoadDirectory(root string) ([]*ProtoFile, error) {
	var files []*ProtoFile
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".proto") {
			return nil
		}

		if l.shouldExclude(path, root) {
			return nil
		}

		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			pf, err := l.LoadFile(p)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("loading %s: %w", p, err):
				default:
				}
				return
			}

			mu.Lock()
			files = append(files, pf)
			mu.Unlock()
		}(path)

		return nil
	})

	if err != nil {
		return nil, err
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return files, nil
}

func (l *Loader) LoadFile(path string) (*ProtoFile, error) {
	l.mu.Lock()
	if cached, ok := l.cache[path]; ok {
		l.mu.Unlock()
		return &ProtoFile{
			Path:  path,
			Proto: cached,
		}, nil
	}
	l.mu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	reader := strings.NewReader(string(content))
	parser := proto.NewParser(reader)
	parser.Filename(path)

	definition, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing proto: %w", err)
	}

	l.mu.Lock()
	l.cache[path] = definition
	l.mu.Unlock()

	return &ProtoFile{
		Path:    path,
		Proto:   definition,
		Content: string(content),
	}, nil
}

func (l *Loader) shouldExclude(path, root string) bool {
	// Calculate relative path from root
	relativePath, err := filepath.Rel(root, path)
	if err != nil {
		relativePath = path
	}

	for _, pattern := range l.excludes {
		// Try relative path matching (primary method)
		if matched, _ := doublestar.Match(pattern, relativePath); matched {
			return true
		}

		// Try full path matching
		if matched, _ := doublestar.Match(pattern, path); matched {
			return true
		}

		// (patterns without directory separators)
		if !strings.Contains(pattern, "/") && !strings.Contains(pattern, "**") {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return true
			}
		}
	}
	return false
}

type ProtoFile struct {
	Path    string
	Proto   *proto.Proto
	Content string
}
