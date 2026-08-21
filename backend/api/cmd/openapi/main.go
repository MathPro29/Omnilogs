package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

const (
	sourceSpec = "api/openapi.yaml"
	docsSpec   = "docs/swagger/openapi.yaml"
)

func main() {
	if len(os.Args) < 2 {
		exitf("usage: go run ./cmd/openapi [validate|sync]")
	}

	switch os.Args[1] {
	case "validate":
		if err := validateSpec(sourceSpec); err != nil {
			exitf("openapi validation failed: %v", err)
		}
		fmt.Println("openapi validation passed")
	case "sync":
		if err := validateSpec(sourceSpec); err != nil {
			exitf("openapi validation failed: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(docsSpec), 0o755); err != nil {
			exitf("create docs directory failed: %v", err)
		}
		content, err := os.ReadFile(sourceSpec)
		if err != nil {
			exitf("read source spec failed: %v", err)
		}
		if err := os.WriteFile(docsSpec, content, 0o644); err != nil {
			exitf("write docs spec failed: %v", err)
		}
		fmt.Println("openapi spec synced")
	default:
		exitf("unknown command: %s", os.Args[1])
	}
}

func validateSpec(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var spec map[string]any
	if err := yaml.Unmarshal(content, &spec); err != nil {
		return err
	}
	if spec["openapi"] == nil {
		return errors.New("missing openapi field")
	}
	info, ok := spec["info"].(map[string]any)
	if !ok || info["title"] == nil || info["version"] == nil {
		return errors.New("missing info.title or info.version")
	}
	paths, ok := spec["paths"].(map[string]any)
	if !ok || len(paths) == 0 {
		return errors.New("paths must not be empty")
	}
	return nil
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
