package transpiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kmmz532/php2js/internal/generator"
	"github.com/kmmz532/php2js/internal/parser"
	"github.com/kmmz532/php2js/internal/runtime"
	"github.com/kmmz532/php2js/internal/transformer"
	"github.com/kmmz532/php2js/internal/worker"
)

// Config holds transpiler configuration.
type Config struct {
	InputDir    string
	OutputDir   string
	Verbose     bool
	DumpAST     bool
	GenWorker   bool
	ProjectName string
}

// Result holds the result of a transpilation run.
type Result struct {
	FilesProcessed int
	FilesGenerated int
	Errors         int
	ErrorMessages  []string
}

// Transpiler orchestrates the PHP-to-JS transpilation pipeline.
type Transpiler struct {
	config *Config
}

// New creates a new Transpiler with the given config.
func New(config *Config) *Transpiler {
	return &Transpiler{config: config}
}

// Run executes the transpilation pipeline.
func (t *Transpiler) Run() (*Result, error) {
	result := &Result{}

	// 1. Discover PHP files
	phpFiles, err := t.discoverFiles()
	if err != nil {
		return nil, fmt.Errorf("discovering files: %w", err)
	}

	if t.config.Verbose {
		fmt.Printf("Found %d PHP files\n", len(phpFiles))
	}

	// 2. Create output directory structure
	transpiledDir := filepath.Join(t.config.OutputDir, "src", "transpiled")
	if err := os.MkdirAll(transpiledDir, 0755); err != nil {
		return nil, fmt.Errorf("creating transpiled dir: %w", err)
	}

	// 3. Process each file
	p := parser.New()
	trans := transformer.New()
	gen := generator.New()

	for _, phpFile := range phpFiles {
		result.FilesProcessed++

		relPath, err := filepath.Rel(t.config.InputDir, phpFile)
		if err != nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s: %v", phpFile, err))
			continue
		}

		if t.config.Verbose {
			fmt.Printf("Processing: %s\n", relPath)
		}

		// Read source
		src, err := os.ReadFile(phpFile)
		if err != nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s: read error: %v", relPath, err))
			continue
		}

		// Parse PHP
		ast, parseErrors := p.Parse(src, relPath)
		if len(parseErrors) > 0 && t.config.Verbose {
			for _, e := range parseErrors {
				fmt.Printf("  Parse warning: %s: %s\n", relPath, e)
			}
		}

		if ast == nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s: parse failed completely", relPath))
			continue
		}

		// Dump AST if requested
		if t.config.DumpAST {
			p.DumpAST(ast, relPath)
		}

		// Transform to JS AST
		jsAST := trans.Transform(ast, relPath)

		// Generate JS code
		jsCode := gen.Generate(jsAST)

		// Write output
		outPath := t.outputPath(relPath, transpiledDir)
		if err := t.writeOutput(outPath, jsCode); err != nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s: write error: %v", relPath, err))
			continue
		}

		result.FilesGenerated++

		if t.config.Verbose {
			fmt.Printf("  -> %s\n", outPath)
		}
	}

	// 3.5 Generate registry.js
	registryCode := "export default async function loadModule(path) {\n  switch(path) {\n"
	for _, phpFile := range phpFiles {
		relPath, _ := filepath.Rel(t.config.InputDir, phpFile)
		jsPath := t.outputPath(relPath, "")
		jsPath = strings.ReplaceAll(jsPath, "\\", "/") // Ensure forward slashes
		// Remove leading slash if any
		jsPath = strings.TrimPrefix(jsPath, "/")
		
		// The transpiled code might request the original php file name, or the js file name.
		// We'll support both the PHP path and the JS path.
		phpPath := strings.ReplaceAll(relPath, "\\", "/")
		
		registryCode += fmt.Sprintf("    case '%s': return await import('./%s');\n", phpPath, jsPath)
		registryCode += fmt.Sprintf("    case '%s': return await import('./%s');\n", jsPath, jsPath)
	}
	registryCode += "    default: return null;\n  }\n}\n"
	if err := t.writeOutput(filepath.Join(transpiledDir, "registry.js"), registryCode); err != nil {
		return nil, fmt.Errorf("writing registry.js: %w", err)
	}

	// 4. Write runtime library
	runtimeDir := filepath.Join(t.config.OutputDir, "src", "runtime")
	if err := runtime.WriteRuntime(runtimeDir); err != nil {
		return nil, fmt.Errorf("writing runtime: %w", err)
	}
	if t.config.Verbose {
		fmt.Println("Runtime library written")
	}

	// 5. Generate Workers entry point and config
	if t.config.GenWorker {
		if err := worker.Generate(t.config.OutputDir, t.config.ProjectName); err != nil {
			return nil, fmt.Errorf("generating worker: %w", err)
		}
		if t.config.Verbose {
			fmt.Println("Workers entry point generated")
		}
	}

	return result, nil
}

// discoverFiles finds all PHP files in the input directory.
func (t *Transpiler) discoverFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(t.config.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".php" || ext == ".inc" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// outputPath converts a PHP file path to a JS output path.
func (t *Transpiler) outputPath(relPath string, outputDir string) string {
	// Create subdirectories as needed
	dir := filepath.Dir(relPath)
	base := filepath.Base(relPath)

	// Change extension: .php -> .js, .inc.php -> .js
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if filepath.Ext(name) == ".inc" {
		name = strings.TrimSuffix(name, ".inc")
	}

	return filepath.Join(outputDir, dir, name+".js")
}

// writeOutput writes generated JS to disk.
func (t *Transpiler) writeOutput(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
