// Package ast defines the abstract syntax tree nodes for TotalScript.
package ast

import (
	"bytes"
	"strings"

	"github.com/mishankov/totalscript-lang/internal/token"
)

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node in the AST.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node in the AST.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// VarStatement represents a variable declaration.
// var x: integer = 5
type VarStatement struct {
	Token token.Token // the 'var' token
	Name  *Identifier
	Type  *TypeExpression
	Value Expression
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Literal }
func (vs *VarStatement) String() string {
	var out bytes.Buffer
	out.WriteString(vs.TokenLiteral() + " ")
	out.WriteString(vs.Name.String())
	if vs.Type != nil {
		out.WriteString(": ")
		out.WriteString(vs.Type.String())
	}
	if vs.Value != nil {
		out.WriteString(" = ")
		out.WriteString(vs.Value.String())
	}
	return out.String()
}

// ConstStatement represents a constant declaration.
// const PI: float = 3.14
type ConstStatement struct {
	Token token.Token // the 'const' token
	Name  *Identifier
	Type  *TypeExpression
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) String() string {
	var out bytes.Buffer
	out.WriteString(cs.TokenLiteral() + " ")
	out.WriteString(cs.Name.String())
	if cs.Type != nil {
		out.WriteString(": ")
		out.WriteString(cs.Type.String())
	}
	if cs.Value != nil {
		out.WriteString(" = ")
		out.WriteString(cs.Value.String())
	}
	return out.String()
}

// ImportStatement represents an import declaration.
// import "math"
// import "./utils"
// import "./geometry" as geo
type ImportStatement struct {
	Token      token.Token // the 'import' token
	Path       string      // Module path: "math", "./utils", "./lib/helpers"
	Alias      string      // Optional alias from 'as' clause, "" if none
	ModuleName string      // Computed module name for qualified access
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out bytes.Buffer
	out.WriteString(is.TokenLiteral() + " ")
	out.WriteString(`"` + is.Path + `"`)
	if is.Alias != "" {
		out.WriteString(" as ")
		out.WriteString(is.Alias)
	}
	return out.String()
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral())
	if rs.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(rs.ReturnValue.String())
	}
	return out.String()
}

// ExpressionStatement represents a statement consisting of a single expression.
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement represents a block of statements.
type BlockStatement struct {
	Token      token.Token // the '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
		out.WriteString(" ")
	}
	out.WriteString("}")
	return out.String()
}

// BreakStatement represents a break statement.
type BreakStatement struct {
	Token token.Token // the 'break' token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return bs.TokenLiteral() }

// ContinueStatement represents a continue statement.
type ContinueStatement struct {
	Token token.Token // the 'continue' token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return cs.TokenLiteral() }

// WhileStatement represents a while loop.
type WhileStatement struct {
	Token     token.Token // the 'while' token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(" ")
	out.WriteString(ws.Body.String())
	return out.String()
}

// ForStatement represents a for loop (both for-in and C-style).
type ForStatement struct {
	Token token.Token // the 'for' token

	// For-in style: for i in 0..10 { ... } or for index, value in array { ... }
	Iterator     *Identifier // optional second identifier for index/key
	Value        *Identifier
	Iterable     Expression
	IsRangeStyle bool

	// C-style: for var i = 0; i < 10; i += 1 { ... }
	Init      Statement
	Condition Expression
	Post      Expression

	Body *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	if fs.IsRangeStyle {
		if fs.Iterator != nil {
			out.WriteString(fs.Iterator.String())
			out.WriteString(", ")
		}
		out.WriteString(fs.Value.String())
		out.WriteString(" in ")
		out.WriteString(fs.Iterable.String())
	} else {
		out.WriteString(fs.Init.String())
		out.WriteString("; ")
		out.WriteString(fs.Condition.String())
		out.WriteString("; ")
		out.WriteString(fs.Post.String())
	}
	out.WriteString(" ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// SwitchStatement represents a switch statement.
type SwitchStatement struct {
	Token   token.Token // the 'switch' token
	Value   Expression
	Cases   []*CaseClause
	Default *BlockStatement
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SwitchStatement) String() string {
	var out bytes.Buffer
	out.WriteString("switch ")
	out.WriteString(ss.Value.String())
	out.WriteString(" { ")
	for _, c := range ss.Cases {
		out.WriteString(c.String())
		out.WriteString(" ")
	}
	if ss.Default != nil {
		out.WriteString("default ")
		out.WriteString(ss.Default.String())
	}
	out.WriteString(" }")
	return out.String()
}

// CaseClause represents a case clause in a switch statement.
type CaseClause struct {
	Token  token.Token // the 'case' token
	Values []Expression
	Body   *BlockStatement
}

func (cc *CaseClause) String() string {
	var out bytes.Buffer
	out.WriteString("case ")
	vals := []string{}
	for _, v := range cc.Values {
		vals = append(vals, v.String())
	}
	out.WriteString(strings.Join(vals, ", "))
	out.WriteString(" ")
	out.WriteString(cc.Body.String())
	return out.String()
}

// Identifier represents an identifier expression.
type Identifier struct {
	Token token.Token // the IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a float literal.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string literal.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return `"` + sl.Value + `"` }

// BooleanLiteral represents a boolean literal.
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// NullLiteral represents a null literal.
type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "null" }

// ArrayLiteral represents an array literal.
type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// MapLiteral represents a map literal.
type MapLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range ml.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// PrefixExpression represents a prefix expression.
type PrefixExpression struct {
	Token    token.Token // the prefix token, e.g. !, -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents an infix expression.
type InfixExpression struct {
	Token    token.Token // the operator token, e.g. +, -, *, /
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// IfExpression represents an if expression.
type IfExpression struct {
	Token       token.Token // the 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// FunctionLiteral represents a function literal.
type FunctionLiteral struct {
	Token      token.Token // the 'function' token
	Parameters []*Parameter
	ReturnType *TypeExpression
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fl.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(fl.ReturnType.String())
	}
	out.WriteString(" ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// Parameter represents a function parameter.
type Parameter struct {
	Name *Identifier
	Type *TypeExpression
}

func (p *Parameter) String() string {
	var out bytes.Buffer
	out.WriteString(p.Name.String())
	if p.Type != nil {
		out.WriteString(": ")
		out.WriteString(p.Type.String())
	}
	return out.String()
}

// CallExpression represents a function call.
type CallExpression struct {
	Token     token.Token // the '(' token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// IndexExpression represents an index operation (array[0], map["key"]).
type IndexExpression struct {
	Token token.Token // the '[' token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// MemberExpression represents a member access (obj.property).
type MemberExpression struct {
	Token  token.Token // the '.' token
	Object Expression
	Member *Identifier
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberExpression) String() string {
	var out bytes.Buffer
	out.WriteString(me.Object.String())
	out.WriteString(".")
	out.WriteString(me.Member.String())
	return out.String()
}

// RangeExpression represents a range (0..10, 0..=10).
type RangeExpression struct {
	Token     token.Token // the '..' or '..=' token
	Start     Expression
	End       Expression
	Inclusive bool
}

func (re *RangeExpression) expressionNode()      {}
func (re *RangeExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RangeExpression) String() string {
	var out bytes.Buffer
	out.WriteString(re.Start.String())
	if re.Inclusive {
		out.WriteString("..=")
	} else {
		out.WriteString("..")
	}
	out.WriteString(re.End.String())
	return out.String()
}

// TypeExpression represents a type annotation.
type TypeExpression struct {
	Token    token.Token
	Name     string
	Optional bool     // true if ends with ?
	Union    []string // for union types (integer | string)
	Generic  []string // for generic types (array<integer>)
}

func (te *TypeExpression) String() string {
	var out bytes.Buffer
	if len(te.Union) > 0 {
		out.WriteString(strings.Join(te.Union, " | "))
	} else {
		out.WriteString(te.Name)
		if len(te.Generic) > 0 {
			out.WriteString("<")
			out.WriteString(strings.Join(te.Generic, ", "))
			out.WriteString(">")
		}
	}
	if te.Optional {
		out.WriteString("?")
	}
	return out.String()
}

// ModelField represents a field in a model definition.
type ModelField struct {
	Name *Identifier
	Type *TypeExpression
}

func (mf *ModelField) String() string {
	var out bytes.Buffer
	out.WriteString(mf.Name.String())
	out.WriteString(": ")
	out.WriteString(mf.Type.String())
	return out.String()
}

// ModelMethod represents a method in a model definition.
type ModelMethod struct {
	Name     *Identifier
	Function *FunctionLiteral
}

func (mm *ModelMethod) String() string {
	var out bytes.Buffer
	out.WriteString(mm.Name.String())
	out.WriteString(" = ")
	out.WriteString(mm.Function.String())
	return out.String()
}

// ModelLiteral represents a model definition.
type ModelLiteral struct {
	Token        token.Token // the 'model' token
	Fields       []*ModelField
	Methods      []*ModelMethod
	Constructors []*FunctionLiteral // Custom constructors
}

func (ml *ModelLiteral) expressionNode()      {}
func (ml *ModelLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *ModelLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("model {\n")
	for _, field := range ml.Fields {
		out.WriteString("  ")
		out.WriteString(field.String())
		out.WriteString("\n")
	}
	for _, constructor := range ml.Constructors {
		out.WriteString("  constructor = ")
		out.WriteString(constructor.String())
		out.WriteString("\n")
	}
	for _, method := range ml.Methods {
		out.WriteString("  ")
		out.WriteString(method.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

// EnumValue represents a value in an enum definition.
type EnumValue struct {
	Name  *Identifier
	Value Expression
}

func (ev *EnumValue) String() string {
	var out bytes.Buffer
	out.WriteString(ev.Name.String())
	out.WriteString(" = ")
	out.WriteString(ev.Value.String())
	return out.String()
}

// EnumLiteral represents an enum definition.
type EnumLiteral struct {
	Token  token.Token // the 'enum' token
	Values []*EnumValue
}

func (el *EnumLiteral) expressionNode()      {}
func (el *EnumLiteral) TokenLiteral() string { return el.Token.Literal }
func (el *EnumLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("enum {\n")
	for _, value := range el.Values {
		out.WriteString("  ")
		out.WriteString(value.String())
		out.WriteString("\n")
	}
	out.WriteString("}")
	return out.String()
}

// ThisExpression represents the 'this' keyword.
type ThisExpression struct {
	Token token.Token // the 'this' token
}

func (te *ThisExpression) expressionNode()      {}
func (te *ThisExpression) TokenLiteral() string { return te.Token.Literal }
func (te *ThisExpression) String() string       { return "this" }
