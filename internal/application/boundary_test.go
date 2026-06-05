package application_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestApplicationLayerDoesNotImportDeliveryOrPlatformImplementations(t *testing.T) {
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}} {{join .Imports \" \"}}", "./internal/application/...")
	cmd.Dir = "../.."
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list application imports: %v", err)
	}
	disallowed := []string{
		"github.com/fastygo/ui8kit",
		"github.com/fastygo/platform/pkg/modulehost",
		"github.com/fastygo/platform/pkg/render",
		"net/http",
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 0 {
			continue
		}
		pkg := parts[0]
		imports := ""
		if len(parts) == 2 {
			imports = parts[1]
		}
		for _, imp := range strings.Fields(imports) {
			for _, banned := range disallowed {
				if imp == banned {
					t.Fatalf("%s directly imports disallowed package %s", pkg, banned)
				}
			}
		}
	}
}
