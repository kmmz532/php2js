package parser

import (
	"fmt"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/dumper"
	"os"
)

// Parser wraps the VKCOM/php-parser library.
type Parser struct{}

// New creates a new PHP parser.
func New() *Parser {
	return &Parser{}
}

// Parse parses PHP source code and returns the AST root node.
func (p *Parser) Parse(src []byte, filename string) (*ast.Root, []string) {
	var parseErrors []string

	errorHandler := func(e *errors.Error) {
		parseErrors = append(parseErrors, fmt.Sprintf("line %d: %s", e.Pos.StartLine, e.Msg))
	}

	rootNode, err := parser.Parse(src, conf.Config{
		Version:          &version.Version{Major: 7, Minor: 4},
		ErrorHandlerFunc: errorHandler,
	})

	if err != nil {
		parseErrors = append(parseErrors, fmt.Sprintf("fatal: %v", err))
		return nil, parseErrors
	}

	root, ok := rootNode.(*ast.Root)
	if !ok {
		parseErrors = append(parseErrors, "fatal: root node is not *ast.Root")
		return nil, parseErrors
	}

	return root, parseErrors
}

// DumpAST prints the AST for debugging purposes.
func (p *Parser) DumpAST(root *ast.Root, filename string) {
	fmt.Printf("\n=== AST Dump: %s ===\n", filename)
	d := dumper.NewDumper(os.Stdout)
	root.Accept(d)
	fmt.Println()
}
