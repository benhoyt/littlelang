// AST nodes for littlelang parser

package parser

import (
	"fmt"
	"strings"

	. "github.com/benhoyt/littlelang/tokenizer"
)

type Program struct {
	Statements Block
}

func (p *Program) String() string {
	return p.Statements.String()
}

type Block []Statement

func (b Block) String() string {
	lines := []string{}
	for _, s := range b {
		lines = append(lines, fmt.Sprintf("%s", s))
	}
	return strings.Join(lines, "\n")
}

type Statement interface {
	Position() Position
	statementNode()
}

type Assign struct {
	pos    Position
	Target Expression
	Value  Expression
}

func (s *Assign) statementNode()     {}
func (s *Assign) Position() Position { return s.pos }

func (s *Assign) String() string {
	return fmt.Sprintf("%s = %s", s.Target, s.Value)
}

type OuterAssign struct {
	pos   Position
	Name  string
	Value Expression
}

func (s *OuterAssign) statementNode()     {}
func (s *OuterAssign) Position() Position { return s.pos }

func (s *OuterAssign) String() string {
	return fmt.Sprintf("outer %s = %s", s.Name, s.Value)
}

type If struct {
	pos       Position
	Condition Expression
	Body      Block
	Else      Block
}

func (s *If) statementNode()     {}
func (s *If) Position() Position { return s.pos }

func indent(s string) string {
	input := strings.Split(s, "\n")
	output := []string{}
	for _, line := range input {
		output = append(output, "    "+line)
	}
	return strings.Join(output, "\n")
}

func (s *If) String() string {
	str := fmt.Sprintf("if %s {\n%s\n}", s.Condition, indent(s.Body.String()))
	if len(s.Else) > 0 {
		str += fmt.Sprintf(" else {\n%s\n}", indent(s.Else.String()))
	}
	return str
}

type While struct {
	pos       Position
	Condition Expression
	Body      Block
}

func (s *While) statementNode()     {}
func (s *While) Position() Position { return s.pos }

func (s *While) String() string {
	return fmt.Sprintf("while %s {\n%s\n}", s.Condition, indent(s.Body.String()))
}

type For struct {
	pos      Position
	Name     string
	Iterable Expression
	Body     Block
}

func (s *For) statementNode()     {}
func (s *For) Position() Position { return s.pos }

func (s *For) String() string {
	return fmt.Sprintf("for %s in %s {\n%s\n}", s.Name, s.Iterable, indent(s.Body.String()))
}

type Return struct {
	pos    Position
	Result Expression
}

func (s *Return) statementNode()     {}
func (s *Return) Position() Position { return s.pos }

func (s *Return) String() string {
	return fmt.Sprintf("return %s", s.Result)
}

type ExpressionStatement struct {
	pos        Position
	Expression Expression
}

func (s *ExpressionStatement) statementNode()     {}
func (s *ExpressionStatement) Position() Position { return s.pos }

func (s *ExpressionStatement) String() string {
	return fmt.Sprintf("%s", s.Expression)
}

type FunctionDefinition struct {
	pos        Position
	Name       string
	Parameters []string
	Ellipsis   bool
	Body       Block
}

func (s *FunctionDefinition) statementNode()     {}
func (s *FunctionDefinition) Position() Position { return s.pos }

func (s *FunctionDefinition) String() string {
	ellipsisStr := ""
	if s.Ellipsis {
		ellipsisStr = "..."
	}
	bodyStr := ""
	if len(s.Body) != 0 {
		bodyStr = "\n" + indent(s.Body.String()) + "\n"
	}
	return fmt.Sprintf("func %s(%s%s) {%s}",
		s.Name, strings.Join(s.Parameters, ", "), ellipsisStr, bodyStr)
}

type Expression interface {
	Position() Position
	expressionNode()
}

type Binary struct {
	pos      Position
	Left     Expression
	Operator Token
	Right    Expression
}

func (e *Binary) expressionNode()    {}
func (e *Binary) Position() Position { return e.pos }

func (e *Binary) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Operator, e.Right)
}

type Unary struct {
	pos      Position
	Operator Token
	Operand  Expression
}

func (e *Unary) expressionNode()    {}
func (e *Unary) Position() Position { return e.pos }

func (e *Unary) String() string {
	space := ""
	if e.Operator == NOT {
		space = " "
	}
	return fmt.Sprintf("(%s%s%s)", e.Operator, space, e.Operand)
}

type Call struct {
	pos       Position
	Function  Expression
	Arguments []Expression
	Ellipsis  bool
}

func (e *Call) expressionNode()    {}
func (e *Call) Position() Position { return e.pos }

func (e *Call) String() string {
	args := []string{}
	for _, arg := range e.Arguments {
		args = append(args, fmt.Sprintf("%s", arg))
	}
	ellipsisStr := ""
	if e.Ellipsis {
		ellipsisStr = "..."
	}
	return fmt.Sprintf("%s(%s%s)", e.Function, strings.Join(args, ", "), ellipsisStr)
}

type Literal struct {
	pos   Position
	Value interface{}
}

func (e *Literal) expressionNode()    {}
func (e *Literal) Position() Position { return e.pos }

func (e *Literal) String() string {
	if e.Value == nil {
		return "nil"
	}
	if s, ok := e.Value.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", e.Value)
}

type List struct {
	pos    Position
	Values []Expression
}

func (e *List) expressionNode()    {}
func (e *List) Position() Position { return e.pos }

func (e *List) String() string {
	values := []string{}
	for _, value := range e.Values {
		values = append(values, fmt.Sprintf("%s", value))
	}
	return fmt.Sprintf("[%s]", strings.Join(values, ", "))
}

type MapItem struct {
	Key   Expression
	Value Expression
}

type Map struct {
	pos   Position
	Items []MapItem
}

func (e *Map) expressionNode()    {}
func (e *Map) Position() Position { return e.pos }

func (e *Map) String() string {
	items := []string{}
	for _, item := range e.Items {
		items = append(items, fmt.Sprintf("%s: %s", item.Key, item.Value))
	}
	return fmt.Sprintf("{%s}", strings.Join(items, ", "))
}

type FunctionExpression struct {
	pos        Position
	Parameters []string
	Ellipsis   bool
	Body       Block
}

func (e *FunctionExpression) expressionNode()    {}
func (e *FunctionExpression) Position() Position { return e.pos }

func (e *FunctionExpression) String() string {
	ellipsisStr := ""
	if e.Ellipsis {
		ellipsisStr = "..."
	}
	bodyStr := ""
	if len(e.Body) != 0 {
		bodyStr = "\n" + indent(e.Body.String()) + "\n"
	}
	return fmt.Sprintf("func(%s%s) {%s}", strings.Join(e.Parameters, ", "), ellipsisStr, bodyStr)
}

type Subscript struct {
	pos       Position
	Container Expression
	Subscript Expression
}

func (e *Subscript) expressionNode()    {}
func (e *Subscript) Position() Position { return e.pos }

func (e *Subscript) String() string {
	return fmt.Sprintf("%s[%s]", e.Container, e.Subscript)
}

type Variable struct {
	pos  Position
	Name string
}

func (e *Variable) expressionNode()    {}
func (e *Variable) Position() Position { return e.pos }

func (e *Variable) String() string {
	return e.Name
}
