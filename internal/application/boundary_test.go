package application_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestApplicationLayerDoesNotImportDeliveryOrPlatformImplementations(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "./internal/application/...")
	cmd.Dir = "../.."
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list application deps: %v", err)
	}
	for _, disallowed := range []string{
		"github.com/fastygo/ui8kit",
		"github.com/fastygo/platform/pkg/modulehost",
		"github.com/fastygo/platform/pkg/render",
		"net/http",
	} {
		if strings.Contains(string(output), disallowed) {
			t.Fatalf("application layer imports disallowed package %s", disallowed)
		}
	}
}
