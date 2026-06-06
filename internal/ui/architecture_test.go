package ui_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Renderer packages must not reach into application, storage, or delivery layers.
// Screens are built from BFF models only (see Platform roadmap-bff-model.md).
func TestTemplRendererDoesNotImportKernelLayers(t *testing.T) {
	root := filepath.Join("..", "..", "pkg", "templ")
	forbidden := []string{
		"github.com/fastygo/app-gocms/internal/application",
		"github.com/fastygo/app-gocms/internal/storage",
		"github.com/fastygo/app-gocms/internal/delivery",
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".templ" {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			for _, banned := range forbidden {
				if importPath == banned || strings.HasPrefix(importPath, banned+"/") {
					t.Errorf("%s must not import %s (renderer is BFF-only)", path, importPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk pkg/templ: %v", err)
	}
}
