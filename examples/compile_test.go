package examples_test

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
)

func TestExamplesCompile(t *testing.T) {
	t.Parallel()

	directories, err := exampleDirectories(".")
	if err != nil {
		t.Fatalf("discover example directories: %v", err)
	}
	args := append([]string{"test"}, directories...)
	cmd := exec.Command("go", args...)
	cmd.Dir = ".."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("examples do not compile: %v\n%s", err, output)
	}
}

func TestExampleDirectoriesFindsEveryMainPackage(t *testing.T) {
	directories, err := exampleDirectories(".")
	if err != nil {
		t.Fatalf("exampleDirectories() error = %v", err)
	}
	if got, want := len(directories), 16; got != want {
		t.Errorf("exampleDirectories() count = %d, want %d", got, want)
	}
	for _, want := range []string{
		"./examples/gateway-models",
		"./examples/management-usage",
		"./examples/openapi-smoke",
		"./examples/quickstart",
		"./examples/webhook-verify",
	} {
		if !contains(directories, want) {
			t.Errorf("exampleDirectories() = %v, want %q", directories, want)
		}
	}
}

func exampleDirectories(root string) ([]string, error) {
	directories := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "main.go" {
			return nil
		}
		directories = append(directories, "./examples/"+filepath.ToSlash(filepath.Dir(path)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(directories)
	return directories, nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
