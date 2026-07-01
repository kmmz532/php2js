package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kmmz532/php2js/internal/transpiler"
)

var (
	version = "0.1.0"
)

func main() {
	// CLI flags
	inputDir := flag.String("input", "", "Input directory containing PHP source files")
	outputDir := flag.String("output", "", "Output directory for generated JavaScript files")
	showVersion := flag.Bool("version", false, "Show version information")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	dumpAST := flag.Bool("dump-ast", false, "Dump PHP AST for debugging")
	genWorker := flag.Bool("worker", true, "Generate Cloudflare Workers entry point")
	projectName := flag.String("name", "php2js-app", "Project name for wrangler.toml")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "php2js - PHP to JavaScript (Cloudflare Workers) Transpiler\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  php2js -input <php-dir> -output <js-dir> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("php2js v%s\n", version)
		os.Exit(0)
	}

	if *inputDir == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: -input and -output flags are required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate input directory exists
	info, err := os.Stat(*inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: input directory %q: %v\n", *inputDir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %q is not a directory\n", *inputDir)
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot create output directory %q: %v\n", *outputDir, err)
		os.Exit(1)
	}

	config := &transpiler.Config{
		InputDir:    *inputDir,
		OutputDir:   *outputDir,
		Verbose:     *verbose,
		DumpAST:     *dumpAST,
		GenWorker:   *genWorker,
		ProjectName: *projectName,
	}

	t := transpiler.New(config)
	result, err := t.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Transpilation complete!\n")
	fmt.Printf("  Files processed: %d\n", result.FilesProcessed)
	fmt.Printf("  Files generated: %d\n", result.FilesGenerated)
	fmt.Printf("  Errors: %d\n", result.Errors)
	if result.Errors > 0 {
		for _, e := range result.ErrorMessages {
			fmt.Printf("    - %s\n", e)
		}
	}
}
