package themes

import "testing"

func TestDefaultRegistryResolvesBuiltInThemes(t *testing.T) {
	registry := DefaultRegistry()
	for _, id := range []string{"blank", "gocms-default", "missing"} {
		theme, ok := registry.Resolve(id)
		if !ok {
			t.Fatalf("expected theme %q to resolve", id)
		}
		if err := ValidateManifest(theme.Manifest()); err != nil {
			t.Fatalf("manifest %q invalid: %v", id, err)
		}
	}
}

func TestValidateManifestRejectsInvalidTheme(t *testing.T) {
	err := ValidateManifest(Manifest{ID: "Bad Theme", Name: "Bad", Version: "0.1.0", Contract: "go-codex.theme.v0.1", Templates: map[TemplateRole]string{RoleIndex: "index"}})
	if err == nil {
		t.Fatalf("expected invalid id to fail")
	}
	err = ValidateManifest(Manifest{ID: "valid", Name: "Valid", Version: "0.1.0", Contract: "go-codex.theme.v0.1", Templates: map[TemplateRole]string{}})
	if err == nil {
		t.Fatalf("expected missing index role to fail")
	}
}
