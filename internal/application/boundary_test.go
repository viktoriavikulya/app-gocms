package application_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplicationLayerDoesNotImportDeliveryOrPlatformImplementations(t *testing.T) {
	root := filepath.Join("..", "application")
	disallowed := []string{
		"github.com/fastygo/ui8kit",
		"github.com/fastygo/platform/pkg/modulehost",
		"github.com/fastygo/platform/pkg/render",
		"github.com/fastygo/app-gocms/internal/delivery",
		"net/http",
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			for _, banned := range disallowed {
				if importPath == banned || strings.HasPrefix(importPath, banned+"/") {
					t.Errorf("%s must not directly import %s", path, importPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk application layer: %v", err)
	}
}
