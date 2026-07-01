// Package jsast defines the JavaScript AST node types used by the transpiler.
package jsast

// Node is the base interface for all JS AST nodes.
type Node interface {
	nodeType() string
}

// Program represents a complete JS module.
type Program struct {
	Imports    []*ImportDecl
	Body       []Statement
	Exports    []string
	SourceFile string
}

func (p *Program) nodeType() string { return "Program" }

// ImportDecl represents an import statement.
type ImportDecl struct {
	Names   []string
	Default string
	Path    string
	Star    string
}

func (i *ImportDecl) nodeType() string { return "ImportDecl" }

// Statement is a JS statement.
type Statement interface {
	Node
	stmtNode()
}

// Expression is a JS expression.
type Expression interface {
	Node
	exprNode()
}

// ExprStatement wraps an expression as a statement.
type ExprStatement struct{ Expr Expression }

func (s *ExprStatement) nodeType() string { return "ExprStatement" }
func (s *ExprStatement) stmtNode()        {}

// VarDecl represents let/const/var.
type VarDecl struct {
	Kind string
	Name string
	Init Expression
}

func (s *VarDecl) nodeType() string { return "VarDecl" }
func (s *VarDecl) stmtNode()        {}

// Param represents a function parameter.
type Param struct {
	Name    string
	Default Expression
	Rest    bool
}

// FunctionDecl represents a function declaration.
type FunctionDecl struct {
	Name       string
	Params     []*Param
	Body       []Statement
	IsAsync    bool
	IsExported bool
}

func (s *FunctionDecl) nodeType() string { return "FunctionDecl" }
func (s *FunctionDecl) stmtNode()        {}

// ClassDecl represents a class declaration.
type ClassDecl struct {
	Name       string
	Extends    string
	Properties []*ClassProperty
	Methods    []*ClassMethod
	IsExported bool
}

func (s *ClassDecl) nodeType() string { return "ClassDecl" }
func (s *ClassDecl) stmtNode()        {}

// ClassProperty represents a class property.
type ClassProperty struct {
	Name, Access string
	Init         Expression
	IsStatic     bool
}

// ClassMethod represents a class method.
type ClassMethod struct {
	Name, Access string
	Params       []*Param
	Body         []Statement
	IsStatic     bool
	IsAsync      bool
}

// IfStatement represents if/else if/else.
type IfStatement struct {
	Condition Expression
	Body      []Statement
	ElseIf    []*ElseIfClause
	Else      []Statement
}

func (s *IfStatement) nodeType() string { return "IfStatement" }
func (s *IfStatement) stmtNode()        {}

type ElseIfClause struct {
	Condition Expression
	Body      []Statement
}

// ForStatement represents a for loop.
type ForStatement struct {
	Init      Statement
	Condition Expression
	Update    Expression
	Body      []Statement
}

func (s *ForStatement) nodeType() string { return "ForStatement" }
func (s *ForStatement) stmtNode()        {}

// ForOfStatement represents for...of (from foreach).
type ForOfStatement struct {
	Key, Value Expression
	Iter       Expression
	Body       []Statement
}

func (s *ForOfStatement) nodeType() string { return "ForOfStatement" }
func (s *ForOfStatement) stmtNode()        {}

// WhileStatement represents while.
type WhileStatement struct {
	Condition Expression
	Body      []Statement
}

func (s *WhileStatement) nodeType() string { return "WhileStatement" }
func (s *WhileStatement) stmtNode()        {}

// DoWhileStatement represents do-while.
type DoWhileStatement struct {
	Body      []Statement
	Condition Expression
}

func (s *DoWhileStatement) nodeType() string { return "DoWhileStatement" }
func (s *DoWhileStatement) stmtNode()        {}

// ReturnStatement represents return.
type ReturnStatement struct{ Value Expression }

func (s *ReturnStatement) nodeType() string { return "ReturnStatement" }
func (s *ReturnStatement) stmtNode()        {}

// ThrowStatement represents throw.
type ThrowStatement struct{ Value Expression }

func (s *ThrowStatement) nodeType() string { return "ThrowStatement" }
func (s *ThrowStatement) stmtNode()        {}

// TryCatchStatement represents try/catch/finally.
type TryCatchStatement struct {
	Try     []Statement
	Catches []*CatchClause
	Finally []Statement
}

func (s *TryCatchStatement) nodeType() string { return "TryCatchStatement" }
func (s *TryCatchStatement) stmtNode()        {}

type CatchClause struct {
	Param string
	Types []string
	Body  []Statement
}

// SwitchStatement represents switch.
type SwitchStatement struct {
	Discriminant Expression
	Cases        []*SwitchCase
}

func (s *SwitchStatement) nodeType() string { return "SwitchStatement" }
func (s *SwitchStatement) stmtNode()        {}

type SwitchCase struct {
	Test Expression
	Body []Statement
}

type BreakStatement struct{}

func (s *BreakStatement) nodeType() string { return "BreakStatement" }
func (s *BreakStatement) stmtNode()        {}

type ContinueStatement struct{}

func (s *ContinueStatement) nodeType() string { return "ContinueStatement" }
func (s *ContinueStatement) stmtNode()        {}

// RawJS allows injecting raw JavaScript code.
type RawJS struct{ Code string }

func (s *RawJS) nodeType() string { return "RawJS" }
func (s *RawJS) stmtNode()        {}

// --- Expressions ---

type Identifier struct{ Name string }

func (e *Identifier) nodeType() string { return "Identifier" }
func (e *Identifier) exprNode()        {}

// Literal represents string, number, bool, null.
type Literal struct {
	Value string
	Kind  string // "string","number","bool","null","undefined","regex"
}

func (e *Literal) nodeType() string { return "Literal" }
func (e *Literal) exprNode()        {}

type TemplateLiteral struct{ Parts []Expression }

func (e *TemplateLiteral) nodeType() string { return "TemplateLiteral" }
func (e *TemplateLiteral) exprNode()        {}

type BinaryExpr struct {
	Op          string
	Left, Right Expression
}

func (e *BinaryExpr) nodeType() string { return "BinaryExpr" }
func (e *BinaryExpr) exprNode()        {}

type UnaryExpr struct {
	Op      string
	Operand Expression
	Prefix  bool
}

func (e *UnaryExpr) nodeType() string { return "UnaryExpr" }
func (e *UnaryExpr) exprNode()        {}

type AssignExpr struct {
	Op          string
	Left, Right Expression
}

func (e *AssignExpr) nodeType() string { return "AssignExpr" }
func (e *AssignExpr) exprNode()        {}

type CallExpr struct {
	Callee Expression
	Args   []Expression
	Await  bool
}

func (e *CallExpr) nodeType() string { return "CallExpr" }
func (e *CallExpr) exprNode()        {}

type NewExpr struct {
	Callee Expression
	Args   []Expression
}

func (e *NewExpr) nodeType() string { return "NewExpr" }
func (e *NewExpr) exprNode()        {}

type MemberExpr struct {
	Object, Property Expression
	Computed         bool
}

func (e *MemberExpr) nodeType() string { return "MemberExpr" }
func (e *MemberExpr) exprNode()        {}

type ArrayExpr struct{ Elements []Expression }

func (e *ArrayExpr) nodeType() string { return "ArrayExpr" }
func (e *ArrayExpr) exprNode()        {}

type ObjectExpr struct{ Properties []*ObjectProperty }

func (e *ObjectExpr) nodeType() string { return "ObjectExpr" }
func (e *ObjectExpr) exprNode()        {}

type ObjectProperty struct {
	Key, Value Expression
	Computed   bool
}

type ConditionalExpr struct {
	Test, Consequent, Alternate Expression
}

func (e *ConditionalExpr) nodeType() string { return "ConditionalExpr" }
func (e *ConditionalExpr) exprNode()        {}

type ArrowFunc struct {
	Params  []*Param
	Body    []Statement
	IsAsync bool
}

func (e *ArrowFunc) nodeType() string { return "ArrowFunc" }
func (e *ArrowFunc) exprNode()        {}

type LogicalExpr struct {
	Op          string
	Left, Right Expression
}

func (e *LogicalExpr) nodeType() string { return "LogicalExpr" }
func (e *LogicalExpr) exprNode()        {}

type SpreadExpr struct{ Argument Expression }

func (e *SpreadExpr) nodeType() string { return "SpreadExpr" }
func (e *SpreadExpr) exprNode()        {}

type AwaitExpr struct{ Argument Expression }

func (e *AwaitExpr) nodeType() string { return "AwaitExpr" }
func (e *AwaitExpr) exprNode()        {}

type UpdateExpr struct {
	Op      string
	Operand Expression
	Prefix  bool
}

func (e *UpdateExpr) nodeType() string { return "UpdateExpr" }
func (e *UpdateExpr) exprNode()        {}

type NullCoalesceExpr struct{ Left, Right Expression }

func (e *NullCoalesceExpr) nodeType() string { return "NullCoalesceExpr" }
func (e *NullCoalesceExpr) exprNode()        {}
