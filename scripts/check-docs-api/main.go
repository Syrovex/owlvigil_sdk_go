// Command check-docs-api verifies that Go API names used in Markdown examples
// exist in the current SDK source. It is intentionally dependency-free so it
// can run in CI with the repository's declared Go toolchain.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type apiType struct {
	fields map[string]apiField
}

type apiField struct {
	deprecated bool
}

type literal struct {
	pkg      string
	typeName string
	depth    int
}

var (
	typeLiteralRE = regexp.MustCompile(`\b(owlvigil|gateway|management|oauth2|webhook)\.([A-Z][A-Za-z0-9]*)\s*\{`)
	inlineFieldRE = regexp.MustCompile(`\b(owlvigil|gateway|management|oauth2|webhook)\.([A-Z][A-Za-z0-9]*)\s*\{\s*([A-Z][A-Za-z0-9]*)\s*:`)
	fieldRE       = regexp.MustCompile(`^\s*([A-Z][A-Za-z0-9]*)\s*:`)
	clientCallRE  = regexp.MustCompile(`\b(?:client|gatewayClient|managementClient|auth)\.([A-Z][A-Za-z0-9]*)\s*\(`)
	errAssignRE   = regexp.MustCompile(`\berr\s*:=|,\s*err\s*:=`)
	evidenceRE    = regexp.MustCompile(`^<!-- evidence: ([^>]+) -->$`)
	goBlockRE     = regexp.MustCompile("(?s)```go\\n(.*?)```")
	otherBlockRE  = regexp.MustCompile("(?s)```(bash|yaml|text)\\n(.*?)```")
)

func main() {
	write := flag.Bool("write", false, "format fenced Go examples in place")
	flag.Parse()
	apis := map[string]map[string]apiType{}
	methods := map[string]bool{}
	for pkg, dir := range map[string]string{
		"owlvigil":   ".",
		"gateway":    "gateway",
		"management": "management",
		"oauth2":     "oauth2",
		"webhook":    "webhook",
	} {
		loaded, err := loadTypes(dir)
		if err != nil {
			fatalf("load %s API: %v", pkg, err)
		}
		apis[pkg] = loaded
		loadedMethods, err := loadClientMethods(dir)
		if err != nil {
			fatalf("load %s client methods: %v", pkg, err)
		}
		for method := range loadedMethods {
			methods[method] = true
		}
	}

	files, err := markdownFiles()
	if err != nil {
		fatalf("discover Markdown: %v", err)
	}
	if *write {
		for _, path := range files {
			if err := formatMarkdownFile(path); err != nil {
				fatalf("format %s: %v", path, err)
			}
		}
	}
	var failures []string
	exampleCount := 0
	otherExampleCount := 0
	for _, path := range files {
		count, err := countGoExamples(path)
		if err != nil {
			fatalf("count examples in %s: %v", path, err)
		}
		exampleCount += count
		count, err = countOtherExamples(path)
		if err != nil {
			fatalf("count non-Go examples in %s: %v", path, err)
		}
		otherExampleCount += count
		found, err := checkMarkdown(path, apis, methods)
		if err != nil {
			fatalf("check %s: %v", path, err)
		}
		failures = append(failures, found...)
		found, err = checkOtherExamples(path)
		if err != nil {
			fatalf("check shell/YAML examples in %s: %v", path, err)
		}
		failures = append(failures, found...)
	}
	if len(failures) > 0 {
		sort.Strings(failures)
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, failure)
		}
		os.Exit(1)
	}
	fmt.Printf("documentation API checks passed (%d Go and %d non-Go examples in %d Markdown files)\n", exampleCount, otherExampleCount, len(files))
}

func countOtherExamples(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return len(otherBlockRE.FindAll(data, -1)), nil
}

func checkOtherExamples(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	source := string(data)
	var failures []string
	for _, match := range otherBlockRE.FindAllStringSubmatchIndex(source, -1) {
		lineNumber := strings.Count(source[:match[0]], "\n") + 1
		language := source[match[2]:match[3]]
		prefixLines := strings.Split(strings.TrimSpace(source[:match[0]]), "\n")
		previous := ""
		if len(prefixLines) > 0 {
			previous = prefixLines[len(prefixLines)-1]
		}
		evidence := evidenceRE.FindStringSubmatch(previous)
		if evidence == nil {
			failures = append(failures, fmt.Sprintf("%s:%d: %s example is missing an evidence marker", path, lineNumber, language))
		} else {
			for _, evidencePath := range strings.Split(evidence[1], ",") {
				evidencePath = strings.TrimSpace(evidencePath)
				if evidencePath == "" {
					failures = append(failures, fmt.Sprintf("%s:%d: %s example has an empty evidence path", path, lineNumber, language))
					continue
				}
				if _, err := os.Stat(evidencePath); err != nil {
					failures = append(failures, fmt.Sprintf("%s:%d: %s example evidence %s is unavailable: %v", path, lineNumber, language, evidencePath, err))
				}
			}
		}
		if language == "bash" {
			command := exec.Command("sh", "-n")
			command.Stdin = strings.NewReader(source[match[4]:match[5]])
			if output, err := command.CombinedOutput(); err != nil {
				failures = append(failures, fmt.Sprintf("%s:%d: bash example has invalid syntax: %v: %s", path, lineNumber, err, strings.TrimSpace(string(output))))
			}
		}
	}
	return failures, nil
}

func countGoExamples(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return len(goBlockRE.FindAll(data, -1)), nil
}

func formatMarkdownFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	source := string(data)
	matches := goBlockRE.FindAllStringSubmatchIndex(source, -1)
	if len(matches) == 0 {
		return nil
	}
	var output strings.Builder
	last := 0
	for _, match := range matches {
		output.WriteString(source[last:match[2]])
		formatted, err := formatGoExample(source[match[2]:match[3]])
		if err != nil {
			return err
		}
		output.WriteString(formatted)
		last = match[3]
	}
	output.WriteString(source[last:])
	if output.String() == source {
		return nil
	}
	return os.WriteFile(path, []byte(output.String()), 0o644)
}

func loadClientMethods(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := map[string]bool{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil || !function.Name.IsExported() {
				continue
			}
			for _, receiver := range function.Recv.List {
				if isNamedType(receiver.Type, "Client") {
					result[function.Name.Name] = true
				}
			}
		}
	}
	return result, nil
}

func isNamedType(expression ast.Expr, want string) bool {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name == want
	case *ast.StarExpr:
		return isNamedType(value.X, want)
	default:
		return false
	}
}

func loadTypes(dir string) (map[string]apiType, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := map[string]apiType{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				if !typeSpec.Name.IsExported() {
					continue
				}
				fields := map[string]apiField{}
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					compatibilityBlock := false
					previousEndLine := 0
					for _, field := range structType.Fields.List {
						fieldLine := fset.Position(field.Pos()).Line
						if previousEndLine > 0 && fieldLine-previousEndLine > 1 {
							compatibilityBlock = false
						}
						comment := fieldComment(field)
						if isCompatibilityComment(comment) {
							compatibilityBlock = true
						}
						for _, name := range field.Names {
							if name.IsExported() {
								fields[name.Name] = apiField{deprecated: compatibilityBlock}
							}
						}
						previousEndLine = fset.Position(field.End()).Line
					}
				}
				result[typeSpec.Name.Name] = apiType{fields: fields}
			}
		}
	}
	return result, nil
}

func fieldComment(field *ast.Field) string {
	var parts []string
	if field.Doc != nil {
		parts = append(parts, field.Doc.Text())
	}
	if field.Comment != nil {
		parts = append(parts, field.Comment.Text())
	}
	return strings.ToLower(strings.Join(parts, " "))
}

func isCompatibilityComment(comment string) bool {
	for _, marker := range []string{"deprecated", "legacy", "compatibility", "pre-refactor"} {
		if strings.Contains(comment, marker) {
			return true
		}
	}
	return false
}

func markdownFiles() ([]string, error) {
	files, err := filepath.Glob("*.md")
	if err != nil {
		return nil, err
	}
	filtered := files[:0]
	for _, path := range files {
		if path != "AGENTS.md" {
			filtered = append(filtered, path)
		}
	}
	files = filtered
	err = filepath.WalkDir("docs", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path == filepath.Join("docs", "superpowers") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func checkMarkdown(path string, apis map[string]map[string]apiType, methods map[string]bool) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var failures []string
	inGo := false
	depth := 0
	var stack []literal
	var block strings.Builder
	blockStart := 0
	previousNonEmpty := ""
	lineNumber := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if !inGo {
			if line == "```go" {
				match := evidenceRE.FindStringSubmatch(previousNonEmpty)
				if match == nil {
					failures = append(failures, fmt.Sprintf("%s:%d: Go example is missing an evidence marker", path, lineNumber))
				} else {
					for _, evidencePath := range strings.Split(match[1], ",") {
						evidencePath = strings.TrimSpace(evidencePath)
						if evidencePath == "" {
							failures = append(failures, fmt.Sprintf("%s:%d: Go example has an empty evidence path", path, lineNumber))
							continue
						}
						if _, err := os.Stat(evidencePath); err != nil {
							failures = append(failures, fmt.Sprintf("%s:%d: Go example evidence %s is unavailable: %v", path, lineNumber, evidencePath, err))
						}
					}
				}
				inGo = true
				depth = 0
				stack = nil
				block.Reset()
				blockStart = lineNumber + 1
			}
			if strings.TrimSpace(line) != "" {
				previousNonEmpty = line
			}
			continue
		}
		if line == "```" {
			if err := parseGoExample(block.String()); err != nil {
				failures = append(failures, fmt.Sprintf("%s:%d: Go example is not syntactically valid: %v", path, blockStart, err))
			} else if formatted, err := formatGoExample(block.String()); err != nil {
				failures = append(failures, fmt.Sprintf("%s:%d: Go example cannot be formatted: %v", path, blockStart, err))
			} else if formatted != block.String() {
				failures = append(failures, fmt.Sprintf("%s:%d: Go example is not gofmt-formatted", path, blockStart))
			}
			if errAssignRE.MatchString(block.String()) &&
				!strings.Contains(block.String(), "err != nil") &&
				!strings.Contains(block.String(), "errors.As(err") &&
				!strings.Contains(block.String(), "errors.Is(err") {
				failures = append(failures, fmt.Sprintf("%s:%d: Go example assigns err without demonstrating error handling", path, blockStart))
			}
			inGo = false
			continue
		}
		block.WriteString(line)
		block.WriteByte('\n')

		if len(stack) > 0 {
			if match := fieldRE.FindStringSubmatch(line); match != nil {
				current := stack[len(stack)-1]
				api := apis[current.pkg][current.typeName]
				field, ok := api.fields[match[1]]
				if !ok {
					failures = append(failures, fmt.Sprintf("%s:%d: %s.%s has no exported field %s", path, lineNumber, current.pkg, current.typeName, match[1]))
				} else if field.deprecated {
					failures = append(failures, fmt.Sprintf("%s:%d: Go example uses compatibility field %s.%s.%s", path, lineNumber, current.pkg, current.typeName, match[1]))
				}
			}
		}
		for _, match := range inlineFieldRE.FindAllStringSubmatch(line, -1) {
			api, ok := apis[match[1]][match[2]]
			field, fieldExists := api.fields[match[3]]
			if ok && !fieldExists {
				failures = append(failures, fmt.Sprintf("%s:%d: %s.%s has no exported field %s", path, lineNumber, match[1], match[2], match[3]))
			} else if ok && field.deprecated {
				failures = append(failures, fmt.Sprintf("%s:%d: Go example uses compatibility field %s.%s.%s", path, lineNumber, match[1], match[2], match[3]))
			}
		}
		for _, match := range clientCallRE.FindAllStringSubmatch(line, -1) {
			if !methods[match[1]] {
				failures = append(failures, fmt.Sprintf("%s:%d: no public SDK Client has method %s", path, lineNumber, match[1]))
			}
		}

		for _, match := range typeLiteralRE.FindAllStringSubmatch(line, -1) {
			pkg, name := match[1], match[2]
			if _, ok := apis[pkg][name]; !ok {
				failures = append(failures, fmt.Sprintf("%s:%d: unknown documented type %s.%s", path, lineNumber, pkg, name))
				continue
			}
			stack = append(stack, literal{pkg: pkg, typeName: name, depth: depth + 1})
		}

		depth += strings.Count(line, "{") - strings.Count(line, "}")
		for len(stack) > 0 && depth < stack[len(stack)-1].depth {
			stack = stack[:len(stack)-1]
		}
	}
	return failures, scanner.Err()
}

func parseGoExample(source string) error {
	fset := token.NewFileSet()
	if strings.Contains(source, "package ") {
		_, err := parser.ParseFile(fset, "example.go", source, parser.AllErrors)
		return err
	}
	if _, err := parser.ParseFile(fset, "example.go", "package doctest\nfunc example() {\n"+source+"\n}\n", parser.AllErrors); err == nil {
		return nil
	}
	_, err := parser.ParseFile(fset, "example.go", "package doctest\n"+source, parser.AllErrors)
	return err
}

func formatGoExample(source string) (string, error) {
	if strings.Contains(source, "package ") {
		formatted, err := format.Source([]byte(source))
		return string(formatted), err
	}

	const prefix = "package doctest\n\nfunc example() {\n"
	wrapped := "package doctest\nfunc example() {\n" + source + "\n}\n"
	if _, err := parser.ParseFile(token.NewFileSet(), "example.go", wrapped, parser.AllErrors); err == nil {
		formatted, err := format.Source([]byte(wrapped))
		if err != nil {
			return "", err
		}
		body := strings.TrimPrefix(string(formatted), prefix)
		body = strings.TrimSuffix(body, "}\n")
		lines := strings.Split(body, "\n")
		for index, line := range lines {
			lines[index] = strings.TrimPrefix(line, "\t")
		}
		return strings.Join(lines, "\n"), nil
	}

	formatted, err := format.Source([]byte("package doctest\n" + source))
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(string(formatted), "package doctest\n\n"), nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
