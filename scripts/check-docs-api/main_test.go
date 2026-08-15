package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckMarkdown_ValidatesEveryEvidenceLayer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "example.md")
	source := `<!-- evidence: main_test.go -->
` + "```go" + `
value, err := client.Missing(ctx, &management.Request{Old: true})
` + "```" + "\n"
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	apis := map[string]map[string]apiType{
		"management": {
			"Request": {
				fields: map[string]apiField{
					"Old": {deprecated: true},
				},
			},
		},
	}
	failures, err := checkMarkdown(path, apis, map[string]bool{"Valid": true})
	if err != nil {
		t.Fatalf("checkMarkdown(%q) error = %v", path, err)
	}
	joined := strings.Join(failures, "\n")
	for _, want := range []string{
		"uses compatibility field management.Request.Old",
		"no public SDK Client has method Missing",
		"assigns err without demonstrating error handling",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("checkMarkdown(%q) failures = %q, want substring %q", path, joined, want)
		}
	}
}

func TestQuickstartDocumentationMatchesExample(t *testing.T) {
	document, err := os.ReadFile("../../docs/zh-CN/01-quickstart.md")
	if err != nil {
		t.Fatalf("ReadFile(quickstart doc) error = %v", err)
	}
	source, err := os.ReadFile("../../examples/quickstart/main.go")
	if err != nil {
		t.Fatalf("ReadFile(quickstart example) error = %v", err)
	}

	const start = "```go\npackage main\n"
	startIndex := strings.Index(string(document), start)
	if startIndex < 0 {
		t.Fatal("quickstart document is missing the package main Go example")
	}
	after, ok := strings.CutPrefix(string(document[startIndex:]), "```go\n")
	if !ok {
		t.Fatal("quickstart document is missing the package main Go example")
	}
	got, _, ok := strings.Cut(after, "```\n")
	if !ok {
		t.Fatal("quickstart package main Go example is missing its closing fence")
	}
	want := string(source)
	if got != want {
		t.Error("quickstart package main block differs from examples/quickstart/main.go")
	}
}

func TestLoadTypes_MarksCompatibilityFields(t *testing.T) {
	types, err := loadTypes("../../management")
	if err != nil {
		t.Fatalf("loadTypes(management) error = %v", err)
	}
	profile, ok := types["UpdateUserProfileRequest"]
	if !ok {
		t.Fatal("loadTypes(management) is missing UpdateUserProfileRequest")
	}
	if got := profile.fields["Name"].deprecated; !got {
		t.Errorf("UpdateUserProfileRequest.Name deprecated = %t, want true", got)
	}
	if got := profile.fields["Username"].deprecated; got {
		t.Errorf("UpdateUserProfileRequest.Username deprecated = %t, want false", got)
	}
}

func TestParseGoExample(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:   "statement fragment",
			source: `value, err := client.ListModels(ctx); _ = value; _ = err`,
		},
		{
			name:   "declaration",
			source: `func handler() error { return nil }`,
		},
		{
			name:    "invalid syntax",
			source:  `if err != nil {`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := parseGoExample(test.source)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Errorf("parseGoExample(%q) error = %v, want error presence = %t", test.source, err, test.wantErr)
			}
		})
	}
}
