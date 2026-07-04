package transformer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/kmmz532/php2js/internal/jsast"
)

// transformExpr converts a PHP expression to a JS expression.
func (t *Transformer) transformExpr(node ast.Vertex) jsast.Expression {
	if node == nil {
		return &jsast.Literal{Value: "undefined", Kind: "undefined"}
	}

	switch n := node.(type) {
	case *ast.ExprVariable:
		return t.transformVariable(n)
	case *ast.ScalarString:
		s := string(n.Value)
		isSingle := false
		if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') {
			isSingle = s[0] == '\''
			s = s[1 : len(s)-1]
		}
		s = parsePHPString(s, isSingle)
		return &jsast.Literal{Value: "`" + escapeJSString(s) + "`", Kind: "string"}
	case *ast.ScalarLnumber:
		val := string(n.Value)
		if len(val) > 1 && val[0] == '0' && val[1] >= '0' && val[1] <= '7' {
			val = "0o" + val[1:]
		}
		return &jsast.Literal{Value: val, Kind: "number"}
	case *ast.ScalarDnumber:
		return &jsast.Literal{Value: string(n.Value), Kind: "number"}
	case *ast.ScalarEncapsed:
		return t.transformEncapsed(n)
	case *ast.ScalarEncapsedStringPart:
		return &jsast.Literal{Value: fmt.Sprintf(`"%s"`, escapeJSString(string(n.Value))), Kind: "string"}
	case *ast.ScalarEncapsedStringBrackets:
		return t.transformExpr(n.Var)
	case *ast.ScalarEncapsedStringVar:
		return &jsast.Identifier{Name: t.extractName(n.Name)}
	case *ast.ExprAssign:
		return t.transformAssign(n)
	case *ast.ExprAssignPlus:
		return &jsast.AssignExpr{Op: "+=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignMinus:
		return &jsast.AssignExpr{Op: "-=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignMul:
		return &jsast.AssignExpr{Op: "*=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignDiv:
		return &jsast.AssignExpr{Op: "/=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignMod:
		return &jsast.AssignExpr{Op: "%=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignConcat:
		return &jsast.AssignExpr{Op: "+=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprBinaryPlus:
		return &jsast.BinaryExpr{Op: "+", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryMinus:
		return &jsast.BinaryExpr{Op: "-", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryMul:
		return &jsast.BinaryExpr{Op: "*", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryDiv:
		return &jsast.BinaryExpr{Op: "/", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryMod:
		return &jsast.BinaryExpr{Op: "%", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryConcat:
		return &jsast.BinaryExpr{Op: "+", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryEqual:
		return &jsast.BinaryExpr{Op: "==", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryIdentical:
		return &jsast.BinaryExpr{Op: "===", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryNotEqual:
		return &jsast.BinaryExpr{Op: "!=", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryNotIdentical:
		return &jsast.BinaryExpr{Op: "!==", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinarySmaller:
		return &jsast.BinaryExpr{Op: "<", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinarySmallerOrEqual:
		return &jsast.BinaryExpr{Op: "<=", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryGreater:
		return &jsast.BinaryExpr{Op: ">", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryGreaterOrEqual:
		return &jsast.BinaryExpr{Op: ">=", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryBooleanAnd:
		return &jsast.LogicalExpr{Op: "&&", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryBooleanOr:
		return &jsast.LogicalExpr{Op: "||", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryLogicalAnd:
		return &jsast.LogicalExpr{Op: "&&", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryLogicalOr:
		return &jsast.LogicalExpr{Op: "||", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryBitwiseAnd:
		return &jsast.BinaryExpr{Op: "&", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryBitwiseOr:
		return &jsast.BinaryExpr{Op: "|", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryBitwiseXor:
		return &jsast.BinaryExpr{Op: "^", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryShiftLeft:
		return &jsast.BinaryExpr{Op: "<<", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryShiftRight:
		return &jsast.BinaryExpr{Op: ">>", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBinaryCoalesce:
		return &jsast.NullCoalesceExpr{Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprBooleanNot:
		return &jsast.UnaryExpr{Op: "!", Operand: t.transformExpr(n.Expr), Prefix: true}
	case *ast.ExprUnaryMinus:
		return &jsast.UnaryExpr{Op: "-", Operand: t.transformExpr(n.Expr), Prefix: true}
	case *ast.ExprUnaryPlus:
		return &jsast.UnaryExpr{Op: "+", Operand: t.transformExpr(n.Expr), Prefix: true}
	case *ast.ExprBitwiseNot:
		return &jsast.UnaryExpr{Op: "~", Operand: t.transformExpr(n.Expr), Prefix: true}
	case *ast.ExprPreInc:
		return &jsast.UpdateExpr{Op: "++", Operand: t.transformExpr(n.Var), Prefix: true}
	case *ast.ExprPreDec:
		return &jsast.UpdateExpr{Op: "--", Operand: t.transformExpr(n.Var), Prefix: true}
	case *ast.ExprPostInc:
		return &jsast.UpdateExpr{Op: "++", Operand: t.transformExpr(n.Var), Prefix: false}
	case *ast.ExprPostDec:
		return &jsast.UpdateExpr{Op: "--", Operand: t.transformExpr(n.Var), Prefix: false}
	case *ast.ExprFunctionCall:
		return t.transformFunctionCall(n)
	case *ast.ExprMethodCall:
		return t.transformMethodCall(n)
	case *ast.ExprStaticCall:
		return t.transformStaticCall(n)
	case *ast.ExprPropertyFetch:
		return t.transformPropertyFetch(n)
	case *ast.ExprStaticPropertyFetch:
		return t.transformStaticPropertyFetch(n)
	case *ast.ExprArrayDimFetch:
		return t.transformArrayDimFetch(n)
	case *ast.ExprArray:
		return t.transformArray(n)
	case *ast.ExprArrayItem:
		return t.transformExpr(n.Val)
	case *ast.ExprNew:
		return t.transformNew(n)
	case *ast.ExprTernary:
		return t.transformTernary(n)
	case *ast.ExprIsset:
		return t.transformIsset(n)
	case *ast.ExprEmpty:
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "empty"}},
			Args:   []jsast.Expression{t.transformExpr(n.Expr)},
		}
	case *ast.ExprCastInt:
		return &jsast.CallExpr{Callee: &jsast.Identifier{Name: "parseInt"}, Args: []jsast.Expression{t.transformExpr(n.Expr)}}
	case *ast.ExprCastDouble:
		return &jsast.CallExpr{Callee: &jsast.Identifier{Name: "parseFloat"}, Args: []jsast.Expression{t.transformExpr(n.Expr)}}
	case *ast.ExprCastString:
		return &jsast.CallExpr{Callee: &jsast.Identifier{Name: "String"}, Args: []jsast.Expression{t.transformExpr(n.Expr)}}
	case *ast.ExprCastBool:
		return &jsast.CallExpr{Callee: &jsast.Identifier{Name: "Boolean"}, Args: []jsast.Expression{t.transformExpr(n.Expr)}}
	case *ast.ExprCastArray:
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "toArray"}},
			Args:   []jsast.Expression{t.transformExpr(n.Expr)},
		}
	case *ast.ExprInstanceOf:
		return &jsast.BinaryExpr{Op: "instanceof", Left: t.transformExpr(n.Expr), Right: t.transformExpr(n.Class)}
	case *ast.ExprClosure:
		return t.transformClosure(n)
	case *ast.ExprConstFetch:
		return t.transformConstFetch(n)
	case *ast.ExprClassConstFetch:
		return t.transformClassConstFetch(n)
	case *ast.ExprInclude:
		t.needsAsync = true
		return t.transformInclude(n)
	case *ast.ExprRequire:
		t.needsAsync = true
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "require"}},
			Args:   []jsast.Expression{t.transformExpr(n.Expr)},
			Await:  true,
		}
	case *ast.ExprRequireOnce:
		t.needsAsync = true
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "require_once"}},
			Args:   []jsast.Expression{t.transformExpr(n.Expr)},
			Await:  true,
		}
	case *ast.ExprIncludeOnce:
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "include_once"}},
			Args:   []jsast.Expression{t.transformExpr(n.Expr)},
			Await:  true,
		}
	case *ast.ExprExit:
		if n.Expr != nil {
			return &jsast.CallExpr{
				Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "exit"}},
				Args:   []jsast.Expression{t.transformExpr(n.Expr)},
			}
		}
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "exit"}},
			Args:   []jsast.Expression{},
		}
	case *ast.ExprAssignReference:
		// JS doesn't have references; treat as normal assignment
		return &jsast.AssignExpr{Op: "=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ScalarHeredoc:
		return t.transformHeredoc(n)
	case *ast.ExprBrackets:
		return t.transformExpr(n.Expr)
	case *ast.ExprErrorSuppress:
		// @ operator - just evaluate the expression, wrap in try-catch at runtime
		return t.transformExpr(n.Expr)
	case *ast.ExprPrint:
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "print"}},
			Args:   []jsast.Expression{t.transformExpr(n.Expr)},
		}
	case *ast.ExprBinaryLogicalXor:
		// XOR: (a && !b) || (!a && b)
		return &jsast.BinaryExpr{Op: "!==", Left: &jsast.UnaryExpr{Op: "!", Operand: &jsast.UnaryExpr{Op: "!", Operand: t.transformExpr(n.Left), Prefix: true}, Prefix: true}, Right: &jsast.UnaryExpr{Op: "!", Operand: &jsast.UnaryExpr{Op: "!", Operand: t.transformExpr(n.Right), Prefix: true}, Prefix: true}}
	case *ast.ExprCastObject:
		return &jsast.CallExpr{Callee: &jsast.Identifier{Name: "Object"}, Args: []jsast.Expression{t.transformExpr(n.Expr)}}
	case *ast.ExprAssignBitwiseAnd:
		return &jsast.AssignExpr{Op: "&=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignBitwiseOr:
		return &jsast.AssignExpr{Op: "|=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignBitwiseXor:
		return &jsast.AssignExpr{Op: "^=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignShiftLeft:
		return &jsast.AssignExpr{Op: "<<=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignShiftRight:
		return &jsast.AssignExpr{Op: ">>=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignPow:
		return &jsast.AssignExpr{Op: "**=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprAssignCoalesce:
		return &jsast.AssignExpr{Op: "??=", Left: t.transformAssignLHS(n.Var), Right: t.transformExpr(n.Expr)}
	case *ast.ExprClone:
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "Object"}, Property: &jsast.Identifier{Name: "assign"}},
			Args:   []jsast.Expression{&jsast.ObjectExpr{}, t.transformExpr(n.Expr)},
		}
	case *ast.ExprBinaryPow:
		return &jsast.BinaryExpr{Op: "**", Left: t.transformExpr(n.Left), Right: t.transformExpr(n.Right)}
	case *ast.ExprList:
		arr := &jsast.ArrayExpr{}
		for _, item := range n.Items {
			if a, ok := item.(*ast.ExprArrayItem); ok && a != nil && a.Val != nil {
				arr.Elements = append(arr.Elements, t.transformExpr(a.Val))
			} else {
				arr.Elements = append(arr.Elements, nil)
			}
		}
		return arr
	case *ast.Name:
		return &jsast.Identifier{Name: t.nameToString(n)}
	case *ast.NameFullyQualified:
		return &jsast.Identifier{Name: t.namePartsToString(n.Parts)}
	case *ast.NameRelative:
		return &jsast.Identifier{Name: t.namePartsToString(n.Parts)}
	default:
		return &jsast.Identifier{Name: fmt.Sprintf("/* TODO: expr %T */", node)}
	}
}

func (t *Transformer) transformVariable(n *ast.ExprVariable) jsast.Expression {
	name := t.extractVarName(n)
	if name == "this_" || name == "this" {
		return &jsast.Identifier{Name: "this"}
	}

	// Check for superglobals
	if mapped, ok := phpSuperglobals[name]; ok {
		parts := strings.SplitN(mapped, ".", 2)
		return &jsast.MemberExpr{
			Object:   &jsast.Identifier{Name: parts[0]},
			Property: &jsast.Identifier{Name: parts[1]},
		}
	}

	// Check for static
	if staticName, ok := t.staticVars[len(t.staticVars)-1][name]; ok {
		return &jsast.MemberExpr{
			Object: &jsast.MemberExpr{
				Object:   &jsast.Identifier{Name: "__runtime"},
				Property: &jsast.Identifier{Name: "statics"},
			},
			Property: &jsast.Literal{Value: fmt.Sprintf(`"%s"`, staticName), Kind: "string"},
			Computed: true,
		}
	}

	isGlobal := !t.inFunction || t.globalVars[len(t.globalVars)-1][name]
	if isGlobal {
		return &jsast.MemberExpr{
			Object:   &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "GLOBALS"}},
			Property: &jsast.Literal{Value: fmt.Sprintf(`"%s"`, name), Kind: "string"},
			Computed: true,
		}
	}

	name = sanitizeVarName(name)
	t.currentScope()[name] = true
	return &jsast.Identifier{Name: name}
}

func (t *Transformer) transformAssign(n *ast.ExprAssign) jsast.Expression {
	return &jsast.AssignExpr{
		Op:    "=",
		Left:  t.transformAssignLHS(n.Var),
		Right: t.transformExpr(n.Expr),
	}
}

func (t *Transformer) transformAssignLHS(node ast.Vertex) jsast.Expression {
	if dimFetch, ok := node.(*ast.ExprArrayDimFetch); ok {
		obj := t.transformAssignLHS(dimFetch.Var)
		
		var defaultVal jsast.Expression = &jsast.ObjectExpr{}
		if dimFetch.Dim == nil {
			defaultVal = &jsast.ArrayExpr{}
		}

		leftObj := &jsast.AssignExpr{Op: "??=", Left: obj, Right: defaultVal}
		
		if dimFetch.Dim == nil {
			return &jsast.MemberExpr{
				Object: leftObj,
				Property: &jsast.CallExpr{
					Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "arrayNextIndex"}},
					Args: []jsast.Expression{leftObj},
				},
				Computed: true,
			}
		}
		
		return &jsast.MemberExpr{
			Object:   leftObj,
			Property: t.transformExpr(dimFetch.Dim),
			Computed: true,
		}
	}
	return t.transformExpr(node)
}

func (t *Transformer) transformEncapsed(n *ast.ScalarEncapsed) jsast.Expression {
	tl := &jsast.TemplateLiteral{}
	for _, part := range n.Parts {
		tl.Parts = append(tl.Parts, t.transformExpr(part))
	}
	return tl
}

func (t *Transformer) transformFunctionCall(n *ast.ExprFunctionCall) jsast.Expression {
	name := t.extractCallName(n.Function)

	// Check if it's an async function
	isAsync := asyncFunctions[name]
	if isAsync {
		t.needsAsync = true
	}

	// Map PHP built-in functions to runtime
	callee := t.mapFunctionName(name)

	var args []jsast.Expression
	for _, a := range n.Args {
		if arg, ok := a.(*ast.Argument); ok {
			args = append(args, t.transformExpr(arg.Expr))
		}
	}

	return &jsast.CallExpr{Callee: callee, Args: args, Await: isAsync}
}

func (t *Transformer) transformMethodCall(n *ast.ExprMethodCall) jsast.Expression {
	obj := t.transformExpr(n.Var)
	method := t.extractName(n.Method)

	var args []jsast.Expression
	for _, a := range n.Args {
		if arg, ok := a.(*ast.Argument); ok {
			args = append(args, t.transformExpr(arg.Expr))
		}
	}

	return &jsast.CallExpr{
		Callee: &jsast.MemberExpr{Object: obj, Property: &jsast.Identifier{Name: method}},
		Args:   args,
	}
}

func (t *Transformer) transformStaticCall(n *ast.ExprStaticCall) jsast.Expression {
	className := t.extractCallName(n.Class)
	method := t.extractName(n.Call)
	if className == "parent" && t.inClass != "" {
		className = "super"
	}

	var args []jsast.Expression
	for _, a := range n.Args {
		if arg, ok := a.(*ast.Argument); ok {
			args = append(args, t.transformExpr(arg.Expr))
		}
	}

	return &jsast.CallExpr{
		Callee: &jsast.MemberExpr{
			Object:   &jsast.Identifier{Name: className},
			Property: &jsast.Identifier{Name: method},
		},
		Args: args,
	}
}

func (t *Transformer) transformPropertyFetch(n *ast.ExprPropertyFetch) jsast.Expression {
	return &jsast.MemberExpr{
		Object:   t.transformExpr(n.Var),
		Property: &jsast.Identifier{Name: t.extractName(n.Prop)},
	}
}

func (t *Transformer) transformStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) jsast.Expression {
	className := t.extractCallName(n.Class)
	if className == "parent" && t.inClass != "" {
		className = "super"
	}
	return &jsast.MemberExpr{
		Object:   &jsast.Identifier{Name: className},
		Property: &jsast.Identifier{Name: t.extractVarName(n.Prop)},
	}
}

func (t *Transformer) transformArrayDimFetch(n *ast.ExprArrayDimFetch) jsast.Expression {
	obj := t.transformExpr(n.Var)
	if n.Dim == nil {
		// $arr[] = val -> array push, handled at assignment level
		return obj
	}
	return &jsast.MemberExpr{
		Object:   obj,
		Property: t.transformExpr(n.Dim),
		Computed: true,
	}
}

func (t *Transformer) transformArray(n *ast.ExprArray) jsast.Expression {
	hasStringKeys := false
	for _, item := range n.Items {
		if ai, ok := item.(*ast.ExprArrayItem); ok && ai.Key != nil {
			hasStringKeys = true
			break
		}
	}

	if hasStringKeys {
		// Associative array -> PhpArray or object
		obj := &jsast.ObjectExpr{}
		nextIdx := 0
		for _, item := range n.Items {
			if ai, ok := item.(*ast.ExprArrayItem); ok {
				prop := &jsast.ObjectProperty{Value: t.transformExpr(ai.Val)}
				if ai.Key != nil {
					prop.Key = t.transformExpr(ai.Key)
					prop.Computed = true
				} else {
					prop.Key = &jsast.Literal{Value: strconv.Itoa(nextIdx), Kind: "number"}
					prop.Computed = true
					nextIdx++
				}
				obj.Properties = append(obj.Properties, prop)
			}
		}
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "array"}},
			Args:   []jsast.Expression{obj},
		}
	}

	// Indexed array -> array literal
	arr := &jsast.ArrayExpr{}
	for _, item := range n.Items {
		if ai, ok := item.(*ast.ExprArrayItem); ok {
			arr.Elements = append(arr.Elements, t.transformExpr(ai.Val))
		}
	}
	return arr
}

func (t *Transformer) transformNew(n *ast.ExprNew) jsast.Expression {
	t.needsAsync = true
	className := t.extractCallName(n.Class)
	var args []jsast.Expression
	for _, a := range n.Args {
		if arg, ok := a.(*ast.Argument); ok {
			args = append(args, t.transformExpr(arg.Expr))
		}
	}
	var classExpr jsast.Expression
	if className != "unknown" {
		classExpr = &jsast.Literal{Value: fmt.Sprintf("`%s`", className), Kind: "string"}
	} else {
		classExpr = t.transformExpr(n.Class)
	}

	return &jsast.CallExpr{
		Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "createObject"}},
		Args:   append([]jsast.Expression{classExpr}, args...),
		Await:  true,
	}
}

func (t *Transformer) transformTernary(n *ast.ExprTernary) jsast.Expression {
	cond := t.transformExpr(n.Cond)
	if n.IfTrue == nil {
		// Short ternary: $a ?: $b -> a || b
		return &jsast.LogicalExpr{Op: "||", Left: cond, Right: t.transformExpr(n.IfFalse)}
	}
	return &jsast.ConditionalExpr{
		Test: cond, Consequent: t.transformExpr(n.IfTrue), Alternate: t.transformExpr(n.IfFalse),
	}
}

func setOptionalMember(expr jsast.Expression) jsast.Expression {
	if mem, ok := expr.(*jsast.MemberExpr); ok {
		return &jsast.MemberExpr{
			Object:   setOptionalMember(mem.Object),
			Property: mem.Property,
			Computed: mem.Computed,
			Optional: true,
		}
	}
	return expr
}

func (t *Transformer) transformIsset(n *ast.ExprIsset) jsast.Expression {
	if len(n.Vars) == 1 {
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "isset"}},
			Args:   []jsast.Expression{setOptionalMember(t.transformExpr(n.Vars[0]))},
		}
	}
	// Multiple vars: isset($a, $b) -> __runtime.isset(a) && __runtime.isset(b)
	var exprs jsast.Expression
	for i, v := range n.Vars {
		call := &jsast.CallExpr{
			Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "isset"}},
			Args:   []jsast.Expression{setOptionalMember(t.transformExpr(v))},
		}
		if i == 0 {
			exprs = call
		} else {
			exprs = &jsast.LogicalExpr{Op: "&&", Left: exprs, Right: call}
		}
	}
	return exprs
}

func (t *Transformer) transformClosure(n *ast.ExprClosure) jsast.Expression {
	arrow := &jsast.ArrowFunc{IsAsync: false}
	for _, p := range n.Params {
		arrow.Params = append(arrow.Params, t.transformParam(p))
	}
	for _, s := range n.Stmts {
		arrow.Body = append(arrow.Body, t.transformStmt(s)...)
	}
	return arrow
}

func (t *Transformer) transformConstFetch(n *ast.ExprConstFetch) jsast.Expression {
	name := t.extractCallName(n.Const)
	switch strings.ToLower(name) {
	case "true":
		return &jsast.Literal{Value: "true", Kind: "bool"}
	case "false":
		return &jsast.Literal{Value: "false", Kind: "bool"}
	case "null":
		return &jsast.Literal{Value: "null", Kind: "null"}
	case "php_eol":
		return &jsast.Literal{Value: `"\\n"`, Kind: "string"}
	case "directory_separator", "dir_separator":
		return &jsast.Literal{Value: `"/"`, Kind: "string"}
	default:
		return &jsast.CallExpr{
			Callee: &jsast.MemberExpr{
				Object:   &jsast.Identifier{Name: "__runtime"},
				Property: &jsast.Identifier{Name: "constant"},
			},
			Args: []jsast.Expression{
				&jsast.Literal{Value: fmt.Sprintf(`"%s"`, name), Kind: "string"},
			},
		}
	}
}

func (t *Transformer) transformClassConstFetch(n *ast.ExprClassConstFetch) jsast.Expression {
	className := t.extractCallName(n.Class)
	constName := t.extractName(n.Const)
	if className == "self" && t.inClass != "" {
		className = t.inClass
	}
	if className == "parent" && t.inClass != "" {
		className = "super"
	}
	return &jsast.MemberExpr{
		Object:   &jsast.Identifier{Name: className},
		Property: &jsast.Identifier{Name: constName},
	}
}

func (t *Transformer) transformInclude(n *ast.ExprInclude) jsast.Expression {
	// include/require -> dynamic import (simplified)
	return &jsast.CallExpr{
		Callee: &jsast.MemberExpr{Object: &jsast.Identifier{Name: "__runtime"}, Property: &jsast.Identifier{Name: "include"}},
		Args:   []jsast.Expression{t.transformExpr(n.Expr)},
		Await:  true,
	}
}

func (t *Transformer) transformHeredoc(n *ast.ScalarHeredoc) jsast.Expression {
	// Heredoc/Nowdoc -> template literal
	tl := &jsast.TemplateLiteral{}
	for _, part := range n.Parts {
		tl.Parts = append(tl.Parts, t.transformExpr(part))
	}
	return tl
}
