package envfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFindsParentEnvFileWithoutOverridingEnvironment(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "examples", "gateway-models")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("OWLVIGIL_GATEWAY_KEY=from-file\nOWLVIGIL_EXAMPLE_VALUE=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Error(err)
		}
	})
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OWLVIGIL_GATEWAY_KEY", "from-environment")

	if err := Load(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("OWLVIGIL_GATEWAY_KEY"); got != "from-environment" {
		t.Errorf("OWLVIGIL_GATEWAY_KEY = %q, want %q", got, "from-environment")
	}
	if got := os.Getenv("OWLVIGIL_EXAMPLE_VALUE"); got != "loaded" {
		t.Errorf("OWLVIGIL_EXAMPLE_VALUE = %q, want %q", got, "loaded")
	}
}

func TestRequiredReturnsHelpfulErrorForMissingValue(t *testing.T) {
	t.Setenv("OWLVIGIL_API_KEY", "")

	_, err := Required("OWLVIGIL_API_KEY")
	if err == nil {
		t.Fatal("Required() error = nil, want missing variable error")
	}
}

func TestExampleEnvTemplateListsCredentialVariables(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", ".env.example")
	template, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read %s: %v", templatePath, err)
	}

	for _, key := range []string{
		"OWLVIGIL_GATEWAY_KEY=",
		"OWLVIGIL_API_KEY=",
		"OWLVIGIL_STAGING_API_KEY=",
		"OWLVIGIL_LOCAL_API_KEY=",
		"OWLVIGIL_CLIENT_ID=",
		"OWLVIGIL_CLIENT_SECRET=",
		"OWLVIGIL_ACCESS_TOKEN=",
		"OWLVIGIL_REFRESH_TOKEN=",
	} {
		if !strings.Contains(string(template), key) {
			t.Errorf("%s does not list %s", templatePath, key)
		}
	}
}

func TestExamplesDoNotEmbedAPIKeyLiterals(t *testing.T) {
	examplesPath := filepath.Join("..", "..")
	err := filepath.WalkDir(examplesPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(source), `WithAPIKey(")`) || strings.Contains(string(source), `apiKey = "`) {
			t.Errorf("%s embeds an API key; read it from a named environment variable instead", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
