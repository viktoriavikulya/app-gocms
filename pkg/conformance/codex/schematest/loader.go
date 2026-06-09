package schematest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaVersion = "codex.v1"

// SchemaVersion returns the Codex contract version for schema artifacts.
func SchemaVersion() string {
	return schemaVersion
}

// SchemaDir returns the absolute path to schema/codex/v1.
func SchemaDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("schema", "codex", "v1")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "schema", "codex", "v1")
}

// FixtureDir returns the absolute path to schema/codex/v1/fixtures.
func FixtureDir() string {
	return filepath.Join(SchemaDir(), "fixtures")
}

// LoadSchema compiles a named schema file from schema/codex/v1.
func LoadSchema(name string) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	dir := SchemaDir()
	if err := registerDir(compiler, dir); err != nil {
		return nil, err
	}
	return compiler.Compile(schemaURL(name))
}

// ValidateFixture loads and validates a fixture file against the named schema.
func ValidateFixture(schemaName, fixtureName string) error {
	schema, err := LoadSchema(schemaName)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(FixtureDir(), fixtureName))
	if err != nil {
		return err
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("decode fixture %q: %w", fixtureName, err)
	}
	return schema.Validate(value)
}

func registerDir(compiler *jsonschema.Compiler, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		var doc any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return fmt.Errorf("decode %s: %w", entry.Name(), err)
		}
		if err := compiler.AddResource(schemaURL(entry.Name()), doc); err != nil {
			return err
		}
	}
	return nil
}

func schemaURL(name string) string {
	return fmt.Sprintf("https://fastygo.dev/schema/codex/v1/%s", name)
}
