// Package runtime embeds and writes the JavaScript runtime library.
package runtime

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed js/*
var runtimeFS embed.FS

// WriteRuntime writes the runtime JS files to the output directory.
func WriteRuntime(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	entries, err := runtimeFS.ReadDir("js")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := runtimeFS.ReadFile("js/" + entry.Name())
		if err != nil {
			return err
		}
		outPath := filepath.Join(outputDir, entry.Name())
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
