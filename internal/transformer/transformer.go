// Package transformer converts PHP AST nodes to JavaScript AST nodes.
package transformer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/kmmz532/php2js/internal/jsast"
)

// phpReservedToJS maps PHP superglobals to runtime accessors.
var phpSuperglobals = map[string]string{
	"_GET":     "__runtime.GET",
	"_POST":    "__runtime.POST",
	"_SERVER":  "__runtime.SERVER",
	"_COOKIE":  "__runtime.COOKIE",
	"_SESSION": "__runtime.SESSION",
	"_REQUEST": "__runtime.REQUEST",
	"_FILES":   "__runtime.FILES",
	"_ENV":     "__runtime.ENV",
	"GLOBALS":  "__runtime.GLOBALS",
}

// jsReservedWords are JavaScript reserved words that need mangling.
var jsReservedWords = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "debugger": true, "default": true, "delete": true, "do": true,
	"else": true, "export": true, "extends": true, "finally": true, "for": true,
	"function": true, "if": true, "import": true, "in": true, "instanceof": true,
	"let": true, "new": true, "return": true, "super": true, "switch": true,
	"this": true, "throw": true, "try": true, "typeof": true, "var": true,
	"void": true, "while": true, "with": true, "yield": true, "await": true,
	"enum": true, "implements": true, "interface": true, "package": true,
	"private": true, "protected": true, "public": true, "static": true,
}

// asyncFunctions are PHP functions that need to become async calls in JS.
var asyncFunctions = map[string]bool{
	"file_get_contents": true, "file_put_contents": true,
	"file_exists": true, "unlink": true, "rename": true,
	"fopen": true, "fclose": true, "fread": true, "fwrite": true,
	"fgets": true, "feof": true,
	"scandir": true, "glob": true, "mkdir": true, "rmdir": true,
	"mysql_query": true, "mysqli_query": true, "mysql_connect": true,
	"mysqli_connect": true, "mysql_fetch_array": true, "mysqli_fetch_assoc": true,
	"mysql_fetch_assoc": true, "mysql_num_rows": true, "mysqli_num_rows": true,
	"sleep": true, "usleep": true,
	// Session
	"session_start": true, "session_destroy": true,
}

// Transformer converts PHP AST to JS AST.
type Transformer struct {
	currentFile string
	inClass     string
	inFunction  bool
	needsAsync  bool
}

// New creates a new Transformer.
func New() *Transformer {
	return &Transformer{}
}

// Transform converts a PHP AST root to a JS Program.
func (t *Transformer) Transform(root *ast.Root, filename string) *jsast.Program {
	t.currentFile = filename
	t.inClass = ""
	t.inFunction = false

	prog := &jsast.Program{
		SourceFile: filename,
	}

	// Add runtime import
	prog.Imports = append(prog.Imports, &jsast.ImportDecl{
		Star: "__runtime",
		Path: runtimeImportPath(filename),
	})

	// Transform statements
	for _, stmt := range root.Stmts {
		jsStmts := t.transformStmt(stmt)
		prog.Body = append(prog.Body, jsStmts...)
	}

	return prog
}

// runtimeImportPath calculates relative path to runtime from a file.
func runtimeImportPath(filename string) string {
	dir := filepath.Dir(filename)
	depth := len(strings.Split(dir, string(filepath.Separator)))
	prefix := strings.Repeat("../", depth)
	return prefix + "runtime/index.js"
}

// transformStmt converts a PHP statement to JS statements.
func (t *Transformer) transformStmt(node ast.Vertex) []jsast.Statement {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.StmtExpression:
		return t.transformExprStmt(n)
	case *ast.StmtEcho:
		return t.transformEcho(n)
	case *ast.StmtIf:
		return t.transformIf(n)
	case *ast.StmtFor:
		return t.transformFor(n)
	case *ast.StmtForeach:
		return t.transformForeach(n)
	case *ast.StmtWhile:
		return t.transformWhile(n)
	case *ast.StmtDo:
		return t.transformDoWhile(n)
	case *ast.StmtReturn:
		return t.transformReturn(n)
	case *ast.StmtFunction:
		return t.transformFunction(n)
	case *ast.StmtClass:
		return t.transformClass(n)
	case *ast.StmtSwitch:
		return t.transformSwitch(n)
	case *ast.StmtTry:
		return t.transformTry(n)
	case *ast.StmtThrow:
		return t.transformThrow(n)
	case *ast.StmtBreak:
		return []jsast.Statement{&jsast.BreakStatement{}}
	case *ast.StmtContinue:
		return []jsast.Statement{&jsast.ContinueStatement{}}
	case *ast.StmtInlineHtml:
		return t.transformInlineHTML(n)
	case *ast.StmtGlobal:
		return t.transformGlobal(n)
	case *ast.StmtStmtList:
		var stmts []jsast.Statement
		for _, s := range n.Stmts {
			stmts = append(stmts, t.transformStmt(s)...)
		}
		return stmts
	case *ast.StmtNop:
		return nil
	case *ast.StmtUnset:
		return t.transformUnset(n)
	case *ast.StmtNamespace:
		// Flatten namespace - just process inner stmts
		var stmts []jsast.Statement
		for _, s := range n.Stmts {
			stmts = append(stmts, t.transformStmt(s)...)
		}
		return stmts
	case *ast.StmtUse:
		// use statements become imports - simplified
		return nil
	case *ast.StmtStatic:
		// static $var = val -> let var = val (static is per-request in Workers)
		var stmts []jsast.Statement
		for _, v := range n.Vars {
			if sv, ok := v.(*ast.StmtStaticVar); ok {
				decl := &jsast.VarDecl{
					Kind: "let",
					Name: t.extractVarName(sv.Var),
				}
				if sv.Expr != nil {
					decl.Init = t.transformExpr(sv.Expr)
				}
				stmts = append(stmts, decl)
			}
		}
		return stmts
	case *ast.StmtConstList:
		var stmts []jsast.Statement
		for _, c := range n.Consts {
			if cn, ok := c.(*ast.StmtConstant); ok {
				stmts = append(stmts, &jsast.ExprStatement{
					Expr: &jsast.CallExpr{
						Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "define"}},
						Args:   []jsast.Expression{&jsast.Literal{Value: fmt.Sprintf(`"%s"`, t.extractName(cn.Name)), Kind: "string"}, t.transformExpr(cn.Expr)},
					},
				})
			}
		}
		return stmts
	default:
		return []jsast.Statement{&jsast.RawJS{
			Code: fmt.Sprintf("/* TODO: unhandled stmt %T */", node),
		}}
	}
}

func (t *Transformer) transformExprStmt(n *ast.StmtExpression) []jsast.Statement {
	expr := t.transformExpr(n.Expr)
	if expr == nil {
		return nil
	}
	return []jsast.Statement{&jsast.ExprStatement{Expr: expr}}
}

func (t *Transformer) transformEcho(n *ast.StmtEcho) []jsast.Statement {
	var stmts []jsast.Statement
	for _, e := range n.Exprs {
		expr := t.transformExpr(e)
		stmts = append(stmts, &jsast.ExprStatement{
			Expr: &jsast.CallExpr{
				Callee: &jsast.MemberExpr{
					Object:   &jsast.Identifier{Name: "__runtime"},
					Property: &jsast.Identifier{Name: "echo"},
				},
				Args: []jsast.Expression{expr},
			},
		})
	}
	return stmts
}

func (t *Transformer) transformInlineHTML(n *ast.StmtInlineHtml) []jsast.Statement {
	html := string(n.Value)
	return []jsast.Statement{&jsast.ExprStatement{
		Expr: &jsast.CallExpr{
			Callee: &jsast.MemberExpr{
				Object:   &jsast.Identifier{Name: "__runtime"},
				Property: &jsast.Identifier{Name: "echo"},
			},
			Args: []jsast.Expression{
				&jsast.Literal{Value: "`" + escapeJSString(html) + "`", Kind: "string"},
			},
		},
	}}
}

func (t *Transformer) transformIf(n *ast.StmtIf) []jsast.Statement {
	ifStmt := &jsast.IfStatement{
		Condition: t.transformExpr(n.Cond),
	}

	// Body
	ifStmt.Body = t.transformStmtBlock(n.Stmt)

	// Else if
	for _, elseIf := range n.ElseIf {
		if ei, ok := elseIf.(*ast.StmtElseIf); ok {
			ifStmt.ElseIf = append(ifStmt.ElseIf, &jsast.ElseIfClause{
				Condition: t.transformExpr(ei.Cond),
				Body:      t.transformStmtBlock(ei.Stmt),
			})
		}
	}

	// Else
	if n.Else != nil {
		if el, ok := n.Else.(*ast.StmtElse); ok {
			ifStmt.Else = t.transformStmtBlock(el.Stmt)
		}
	}

	return []jsast.Statement{ifStmt}
}

func (t *Transformer) transformFor(n *ast.StmtFor) []jsast.Statement {
	forStmt := &jsast.ForStatement{}

	if len(n.Init) > 0 {
		expr := t.transformExpr(n.Init[0])
		forStmt.Init = &jsast.ExprStatement{Expr: expr}
	}
	if len(n.Cond) > 0 {
		forStmt.Condition = t.transformExpr(n.Cond[0])
	}
	if len(n.Loop) > 0 {
		forStmt.Update = t.transformExpr(n.Loop[0])
	}

	forStmt.Body = t.transformStmtBlock(n.Stmt)
	return []jsast.Statement{forStmt}
}

func (t *Transformer) transformForeach(n *ast.StmtForeach) []jsast.Statement {
	forOf := &jsast.ForOfStatement{
		Iter: t.transformExpr(n.Expr),
	}

	if n.Key != nil {
		forOf.Key = t.extractVarName(n.Key)
	}
	forOf.Value = t.extractVarName(n.Var)

	forOf.Body = t.transformStmtBlock(n.Stmt)
	return []jsast.Statement{forOf}
}

func (t *Transformer) transformWhile(n *ast.StmtWhile) []jsast.Statement {
	return []jsast.Statement{&jsast.WhileStatement{
		Condition: t.transformExpr(n.Cond),
		Body:      t.transformStmtBlock(n.Stmt),
	}}
}

func (t *Transformer) transformDoWhile(n *ast.StmtDo) []jsast.Statement {
	return []jsast.Statement{&jsast.DoWhileStatement{
		Condition: t.transformExpr(n.Cond),
		Body:      t.transformStmtBlock(n.Stmt),
	}}
}

func (t *Transformer) transformReturn(n *ast.StmtReturn) []jsast.Statement {
	ret := &jsast.ReturnStatement{}
	if n.Expr != nil {
		ret.Value = t.transformExpr(n.Expr)
	}
	return []jsast.Statement{ret}
}

func (t *Transformer) transformThrow(n *ast.StmtThrow) []jsast.Statement {
	return []jsast.Statement{&jsast.ThrowStatement{
		Value: t.transformExpr(n.Expr),
	}}
}

func (t *Transformer) transformFunction(n *ast.StmtFunction) []jsast.Statement {
	prevInFunc := t.inFunction
	t.inFunction = true
	t.needsAsync = false
	defer func() { t.inFunction = prevInFunc }()

	fn := &jsast.FunctionDecl{
		Name:       t.extractName(n.Name),
		IsExported: true,
	}

	// Parameters
	for _, p := range n.Params {
		fn.Params = append(fn.Params, t.transformParam(p))
	}

	// Body
	for _, s := range n.Stmts {
		fn.Body = append(fn.Body, t.transformStmt(s)...)
	}

	fn.IsAsync = t.needsAsync
	return []jsast.Statement{fn}
}

func (t *Transformer) transformParam(node ast.Vertex) *jsast.Param {
	p, ok := node.(*ast.Parameter)
	if !ok {
		return &jsast.Param{Name: "unknown"}
	}

	param := &jsast.Param{
		Name: t.extractVarName(p.Var),
	}

	if p.DefaultValue != nil {
		param.Default = t.transformExpr(p.DefaultValue)
	}
	if p.VariadicTkn != nil {
		param.Rest = true
	}

	return param
}

func (t *Transformer) transformClass(n *ast.StmtClass) []jsast.Statement {
	prevClass := t.inClass
	className := t.extractName(n.Name)
	t.inClass = className
	defer func() { t.inClass = prevClass }()

	cls := &jsast.ClassDecl{
		Name:       className,
		IsExported: true,
	}

	// Extends
	if n.Extends != nil {
		if ext, ok := n.Extends.(*ast.Name); ok {
			cls.Extends = t.nameToString(ext)
		}
	}

	// Class body
	for _, stmt := range n.Stmts {
		switch s := stmt.(type) {
		case *ast.StmtPropertyList:
			for _, prop := range s.Props {
				if p, ok := prop.(*ast.StmtProperty); ok {
					cp := &jsast.ClassProperty{
						Name:   t.extractVarName(p.Var),
						Access: t.getModifiers(s.Modifiers),
					}
					if p.Expr != nil {
						cp.Init = t.transformExpr(p.Expr)
					}
					cls.Properties = append(cls.Properties, cp)
				}
			}
		case *ast.StmtClassMethod:
			prevInFunc := t.inFunction
			t.inFunction = true
			t.needsAsync = false

			method := &jsast.ClassMethod{
				Name:   t.extractName(s.Name),
				Access: t.getModifiers(s.Modifiers),
			}

			if method.Name == "__construct" {
				method.Name = "constructor"
			}

			for _, p := range s.Params {
				method.Params = append(method.Params, t.transformParam(p))
			}

			if s.Stmt != nil {
				method.Body = t.transformStmtBlock(s.Stmt)
			}

			method.IsAsync = t.needsAsync
			cls.Methods = append(cls.Methods, method)
			t.inFunction = prevInFunc
		}
	}

	return []jsast.Statement{cls}
}

func (t *Transformer) transformSwitch(n *ast.StmtSwitch) []jsast.Statement {
	sw := &jsast.SwitchStatement{
		Discriminant: t.transformExpr(n.Cond),
	}

	for _, c := range n.Cases {
		switch cs := c.(type) {
		case *ast.StmtCase:
			sc := &jsast.SwitchCase{
				Test: t.transformExpr(cs.Cond),
			}
			for _, s := range cs.Stmts {
				sc.Body = append(sc.Body, t.transformStmt(s)...)
			}
			sw.Cases = append(sw.Cases, sc)
		case *ast.StmtDefault:
			sc := &jsast.SwitchCase{Test: nil}
			for _, s := range cs.Stmts {
				sc.Body = append(sc.Body, t.transformStmt(s)...)
			}
			sw.Cases = append(sw.Cases, sc)
		}
	}

	return []jsast.Statement{sw}
}

func (t *Transformer) transformTry(n *ast.StmtTry) []jsast.Statement {
	tc := &jsast.TryCatchStatement{}

	for _, s := range n.Stmts {
		tc.Try = append(tc.Try, t.transformStmt(s)...)
	}

	for _, c := range n.Catches {
		if catch, ok := c.(*ast.StmtCatch); ok {
			cc := &jsast.CatchClause{
				Param: t.extractVarName(catch.Var),
			}
			for _, s := range catch.Stmts {
				cc.Body = append(cc.Body, t.transformStmt(s)...)
			}
			tc.Catches = append(tc.Catches, cc)
		}
	}

	if n.Finally != nil {
		if fin, ok := n.Finally.(*ast.StmtFinally); ok {
			for _, s := range fin.Stmts {
				tc.Finally = append(tc.Finally, t.transformStmt(s)...)
			}
		}
	}

	return []jsast.Statement{tc}
}

func (t *Transformer) transformGlobal(n *ast.StmtGlobal) []jsast.Statement {
	// Convert global $var to: let var_ = __runtime.GLOBALS["var"]
	var stmts []jsast.Statement
	for _, v := range n.Vars {
		name := t.extractVarName(v)
		stmts = append(stmts, &jsast.VarDecl{
			Kind: "let",
			Name: name,
			Init: &jsast.MemberExpr{
				Object:   &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "GLOBALS"}},
				Property: &jsast.Literal{Value: fmt.Sprintf(`"%s"`, name), Kind: "string"},
				Computed: true,
			},
		})
	}
	return stmts
}

func (t *Transformer) transformUnset(n *ast.StmtUnset) []jsast.Statement {
	var stmts []jsast.Statement
	for _, v := range n.Vars {
		expr := t.transformExpr(v)
		if _, ok := expr.(*jsast.Identifier); ok {
			stmts = append(stmts, &jsast.ExprStatement{
				Expr: &jsast.AssignExpr{Op: "=", Left: expr, Right: &jsast.Identifier{Name: "undefined"}},
			})
		} else {
			stmts = append(stmts, &jsast.ExprStatement{
				Expr: &jsast.UnaryExpr{Op: "delete", Operand: expr, Prefix: true},
			})
		}
	}
	return stmts
}

// transformStmtBlock extracts statements from a block node.
func (t *Transformer) transformStmtBlock(node ast.Vertex) []jsast.Statement {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.StmtStmtList:
		var stmts []jsast.Statement
		for _, s := range n.Stmts {
			stmts = append(stmts, t.transformStmt(s)...)
		}
		return stmts
	default:
		return t.transformStmt(node)
	}
}
