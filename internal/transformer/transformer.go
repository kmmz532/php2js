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
	"_GET":     "__runtime.superglobals._GET",
	"_POST":    "__runtime.superglobals._POST",
	"_SERVER":  "__runtime.superglobals._SERVER",
	"_COOKIE":  "__runtime.superglobals._COOKIE",
	"_SESSION": "__runtime.superglobals._SESSION",
	"_REQUEST": "__runtime.superglobals._REQUEST",
	"_FILES":   "__runtime.superglobals._FILES",
	"_ENV":     "__runtime.superglobals._ENV",
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

// IsReserved checks if a given string is a JavaScript reserved word.
func IsReserved(name string) bool {
	return jsReservedWords[name]
}

// asyncFunctions are PHP functions that need to become async calls in JS.
var asyncFunctions = map[string]bool{
	// User functions
	"Factory_DList": true, "Factory_Table": true, "Factory_YTable": true, "Factory_Div": true,
	"do_diff": true, "do_update_diff": true, "get_source": true, "page_write": true,
	"file_head": true, "file_write": true, "add_recent": true, "lastmodified_add": true,
	"put_lastmodified": true, "delete_recent_changes_cache": true, "autolink_pattern_write": true,
	"update_autoalias_cache_file": true, "get_readings": true, "pkwk_chown": true,
	"pkwk_touch_file": true, "pkwk_file_get_contents": true, "prepare_display_materials": true,
	"prepare_links_related": true, "pkwk_log_updates": true, "is_page": true,
	"page_exists_in_history": true, "is_freeze": true, "auto_template": true,
	"do_search": true, "catrule": true, "die_message": true, "die_invalid_pagename": true,
	"get_autolink_pattern": true, "get_autoaliases": true, "md5_file": true,
	"catbody": true, "get_html_scripting_data": true, "edit_form": true,
	"get_template_page_list": true, "links_get_related_db": true, "links_update": true,
	"links_init": true, "links_add": true, "links_delete": true, "links_get_objects": true,
	"links_do_search_page": true, "pop_before_smtp": true, "make_link": true,
	"get_interwiki_url": true, "get_ticketlink_jira_projects": true, "exist_plugin": true,
	"exist_plugin_action": true, "exist_plugin_convert": true, "exist_plugin_inline": true,
	"do_plugin_action": true, "do_plugin_convert": true, "do_plugin_inline": true,
	"pkwk_http_request": true, "check_readable": true, "convert_html": true,
	"manage_page_redirect": true, "ensure_valid_auth_user": true, "check_editable": true,
	"pkwk_script_uri_base": true, "get_base_uri": true, "get_page_uri": true, "get_script_uri": true,

	// Runtime file functions
	"file_get_contents": true, "file_put_contents": true, "file_exists": true, "make_search": true,
	"unlink": true, "rename": true, "scandir": true, "glob": true, "mkdir": true, "rmdir": true,
	"is_dir": true, "is_file": true, "file": true, "filesize": true, "filemtime": true,
	"filectime": true, "fileatime": true, "fopen": true, "fclose": true, "copy": true,
	"touch": true, "opendir": true,

	// Original runtime functions
	"mysql_query": true, "mysqli_query": true, "mysql_connect": true,
	"mysqli_connect": true, "mysql_fetch_array": true, "mysqli_fetch_assoc": true,
	"mysql_fetch_assoc": true, "mysql_num_rows": true, "mysqli_num_rows": true,
	"sleep": true, "usleep": true,
	"session_start": true, "session_destroy": true,
}

// Transformer converts PHP AST to JS AST.
type Transformer struct {
	currentFile string
	inClass     string
	inFunction  bool
	needsAsync  bool
	scopes      []map[string]bool
	globalVars  []map[string]bool
	staticVars  []map[string]string
	currentFunc string
}

// New creates a new Transformer.
func New() *Transformer {
	return &Transformer{
		scopes:     []map[string]bool{{}},
		globalVars: []map[string]bool{{}},
		staticVars: []map[string]string{{}},
	}
}

func (t *Transformer) currentScope() map[string]bool {
	return t.scopes[len(t.scopes)-1]
}

func (t *Transformer) pushScope() {
	t.scopes = append(t.scopes, make(map[string]bool))
	t.globalVars = append(t.globalVars, make(map[string]bool))
	t.staticVars = append(t.staticVars, make(map[string]string))
}

func (t *Transformer) popScope() map[string]bool {
	scope := t.scopes[len(t.scopes)-1]
	t.scopes = t.scopes[:len(t.scopes)-1]
	t.globalVars = t.globalVars[:len(t.globalVars)-1]
	t.staticVars = t.staticVars[:len(t.staticVars)-1]
	return scope
}

// Transform converts a PHP AST root to a JS Program.
func (t *Transformer) Transform(root *ast.Root, filename string) *jsast.Program {
	t.currentFile = filename
	t.inClass = ""
	t.inFunction = false
	t.needsAsync = false
	t.pushScope()

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

	scope := t.popScope()
	var decls []jsast.Statement
	for name := range scope {
		if name != "this" && !jsReservedWords[name] {
			decls = append(decls, &jsast.VarDecl{Kind: "let", Name: name})
		}
	}
	if len(decls) > 0 {
		prog.Body = append(decls, prog.Body...)
	}

	return prog
}

// runtimeImportPath calculates relative path to runtime from a file.
func runtimeImportPath(filename string) string {
	dir := filepath.Dir(filename)
	if dir == "." || dir == "/" {
		return "../runtime/index.js"
	}
	// For "plugin", split returns ["plugin"], length 1. We need 2 levels up: "../../runtime"
	depth := len(strings.Split(dir, string(filepath.Separator)))
	prefix := strings.Repeat("../", depth+1)
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
		var stmts []jsast.Statement
		for _, v := range n.Vars {
			if sv, ok := v.(*ast.StmtStaticVar); ok {
				name := t.extractVarName(sv.Var)
				
				prefix := t.currentFunc
				if t.inClass != "" {
					prefix = t.inClass + "_" + prefix
				}
				if prefix == "" {
					prefix = "global"
				}
				staticName := fmt.Sprintf("%s_%s", prefix, name)
				t.staticVars[len(t.staticVars)-1][name] = staticName

				left := &jsast.MemberExpr{
					Object: &jsast.MemberExpr{
						Object:   &jsast.Identifier{Name: "__runtime"},
						Property: &jsast.Identifier{Name: "statics"},
					},
					Property: &jsast.Literal{Value: fmt.Sprintf(`"%s"`, staticName), Kind: "string"},
					Computed: true,
				}

				var right jsast.Expression
				if sv.Expr != nil {
					right = t.transformExpr(sv.Expr)
				} else {
					right = &jsast.Identifier{Name: "null"}
				}

				stmts = append(stmts, &jsast.ExprStatement{
					Expr: &jsast.AssignExpr{
						Op:    "??=",
						Left:  left,
						Right: right,
					},
				})
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
		forOf.Key = t.transformExpr(n.Key)
	}
	forOf.Value = t.transformExpr(n.Var)

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
	prevFunc := t.currentFunc
	t.inFunction = true
	t.currentFunc = t.extractName(n.Name)
	t.needsAsync = false
	t.pushScope()
	defer func() {
		t.inFunction = prevInFunc
		t.currentFunc = prevFunc
	}()

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

	scope := t.popScope()
	var decls []jsast.Statement
	for name := range scope {
		if name != "this" && !jsReservedWords[name] {
			isParam := false
			for _, p := range n.Params {
				if param, ok := p.(*ast.Parameter); ok {
					if name == t.extractVarName(param.Var) {
						isParam = true
					}
				}
			}
			if !isParam {
				decls = append(decls, &jsast.VarDecl{Kind: "let", Name: name})
			}
		}
	}
	if len(decls) > 0 {
		fn.Body = append(decls, fn.Body...)
	}

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
			prevFunc := t.currentFunc
			t.inFunction = true
			t.currentFunc = t.extractName(s.Name)
			t.needsAsync = false

			t.pushScope()

			method := &jsast.ClassMethod{
				Name:   t.extractName(s.Name),
				Access: t.getModifiers(s.Modifiers),
			}

			for _, p := range s.Params {
				method.Params = append(method.Params, t.transformParam(p))
			}

			if s.Stmt != nil {
				method.Body = t.transformStmtBlock(s.Stmt)
			}

			method.IsAsync = t.needsAsync

			scope := t.popScope()
			var decls []jsast.Statement
			for name := range scope {
				if name != "this" && !jsReservedWords[name] {
					isParam := false
					for _, p := range s.Params {
						if param, ok := p.(*ast.Parameter); ok {
							if name == t.extractVarName(param.Var) {
								isParam = true
							}
						}
					}
					if !isParam {
						decls = append(decls, &jsast.VarDecl{Kind: "let", Name: name})
					}
				}
			}
			if len(decls) > 0 {
				method.Body = append(decls, method.Body...)
			}

			cls.Methods = append(cls.Methods, method)
			t.inFunction = prevInFunc
			t.currentFunc = prevFunc
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
	// Register as global in current scope
	for _, v := range n.Vars {
		name := t.extractVarName(v)
		t.globalVars[len(t.globalVars)-1][name] = true
	}
	return nil
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
