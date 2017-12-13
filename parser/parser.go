// Package parser turns littlelang source code into an abstract syntax tree.
//
// You can parse a single expression with ParseExpression(), or an entire
// program with ParseProgram().
//
package parser

import (
	"fmt"
	"strconv"

	. "github.com/benhoyt/littlelang/tokenizer"
)

// Error is the error type returned by ParseExpression and ParseProgram when
// they encounter a syntax error. You can use this to get the location (line
// and column) of where the error occurred, as well as the error message.
type Error struct {
	Position Position
	Message  string
}

func (e Error) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", e.Position.Line, e.Position.Column, e.Message)
}

type parser struct {
	tokenizer *Tokenizer
	pos       Position
	tok       Token
	val       string
}

func (p *parser) next() {
	p.pos, p.tok, p.val = p.tokenizer.Next()
	if p.tok == ILLEGAL {
		p.error("%s", p.val)
	}
}

func (p *parser) error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	panic(Error{p.pos, message})
}

func (p *parser) expect(tok Token) {
	if p.tok != tok {
		p.error("expected %s and not %s", tok, p.tok)
	}
	p.next()
}

func (p *parser) matches(operators ...Token) bool {
	for _, operator := range operators {
		if p.tok == operator {
			return true
		}
	}
	return false
}

// program = statement*
func (p *parser) program() *Program {
	statements := p.statements(EOF)
	return &Program{statements}
}

func (p *parser) statements(end Token) Block {
	statements := Block{}
	for p.tok != end && p.tok != EOF {
		statements = append(statements, p.statement())
	}
	return statements
}

// statement = if | while | for | return | func | assign | expression
// assign    = NAME ASSIGN expression |
//             call subscript ASSIGN expression |
//             call dot ASSIGN expression
func (p *parser) statement() Statement {
	switch p.tok {
	case IF:
		return p.if_()
	case WHILE:
		return p.while()
	case FOR:
		return p.for_()
	case RETURN:
		return p.return_()
	case FUNC:
		return p.func_()
	}
	pos := p.pos
	expr := p.expression()
	if p.tok == ASSIGN {
		pos = p.pos
		switch expr.(type) {
		case *Variable, *Subscript:
			p.next()
			value := p.expression()
			return &Assign{pos, expr, value}
		default:
			p.error("expected name, subscript, or dot expression on left side of =")
		}
	}
	return &ExpressionStatement{pos, expr}
}

// block = LBRACE statement* RBRACE
func (p *parser) block() Block {
	p.expect(LBRACE)
	body := p.statements(RBRACE)
	p.expect(RBRACE)
	return body
}

// if = IF expression block |
//      IF expression block ELSE block |
//      IF expression block ELSE if
func (p *parser) if_() Statement {
	pos := p.pos
	p.expect(IF)
	condition := p.expression()
	body := p.block()
	var elseBody Block
	if p.tok == ELSE {
		p.next()
		if p.tok == LBRACE {
			elseBody = p.block()
		} else if p.tok == IF {
			elseBody = Block{p.if_()}
		} else {
			p.error("expected { or if after else, not %s", p.tok)
		}
	}
	return &If{pos, condition, body, elseBody}
}

// while = WHILE expression block
func (p *parser) while() Statement {
	pos := p.pos
	p.expect(WHILE)
	condition := p.expression()
	body := p.block()
	return &While{pos, condition, body}
}

// for = FOR NAME IN expression block
func (p *parser) for_() Statement {
	pos := p.pos
	p.expect(FOR)
	name := p.val
	p.expect(NAME)
	p.expect(IN)
	iterable := p.expression()
	body := p.block()
	return &For{pos, name, iterable, body}
}

// return = RETURN expression
func (p *parser) return_() Statement {
	pos := p.pos
	p.expect(RETURN)
	result := p.expression()
	return &Return{pos, result}
}

// func = FUNC NAME params block |
//        FUNC params block
func (p *parser) func_() Statement {
	pos := p.pos
	p.expect(FUNC)
	if p.tok == NAME {
		name := p.val
		p.next()
		params, ellipsis := p.params()
		body := p.block()
		return &FunctionDefinition{pos, name, params, ellipsis, body}
	} else {
		params, ellipsis := p.params()
		body := p.block()
		expr := &FunctionExpression{pos, params, ellipsis, body}
		return &ExpressionStatement{pos, expr}
	}
}

// params = LPAREN RPAREN |
//          LPAREN NAME (COMMA NAME)* ELLIPSIS? COMMA? RPAREN |
func (p *parser) params() ([]string, bool) {
	p.expect(LPAREN)
	params := []string{}
	gotComma := true
	gotEllipsis := false
	for p.tok != RPAREN && p.tok != EOF && !gotEllipsis {
		if !gotComma {
			p.error("expected , between parameters")
		}
		param := p.val
		p.expect(NAME)
		params = append(params, param)
		if p.tok == ELLIPSIS {
			gotEllipsis = true
			p.next()
		}
		if p.tok == COMMA {
			gotComma = true
			p.next()
		} else {
			gotComma = false
		}
	}
	if p.tok != RPAREN && gotEllipsis {
		p.error("can only have ... after last parameter")
	}
	p.expect(RPAREN)
	return params, gotEllipsis
}

func (p *parser) binary(parseFunc func() Expression, operators ...Token) Expression {
	expr := parseFunc()
	for p.matches(operators...) {
		op := p.tok
		pos := p.pos
		p.next()
		right := parseFunc()
		expr = &Binary{pos, expr, op, right}
	}
	return expr
}

// expression = and (OR and)*
func (p *parser) expression() Expression {
	return p.binary(p.and, OR)
}

// and = not (AND not)*
func (p *parser) and() Expression {
	return p.binary(p.not, AND)
}

// not = NOT not | equality
func (p *parser) not() Expression {
	if p.tok == NOT {
		pos := p.pos
		p.next()
		operand := p.not()
		return &Unary{pos, NOT, operand}
	}
	return p.equality()
}

// equality = comparison ((EQUAL | NOTEQUAL) comparison)*
func (p *parser) equality() Expression {
	return p.binary(p.comparison, EQUAL, NOTEQUAL)
}

// comparison = addition ((LT | LTE | GT | GTE | IN) addition)*
func (p *parser) comparison() Expression {
	return p.binary(p.addition, LT, LTE, GT, GTE, IN)
}

// addition = multiply ((PLUS | MINUS) multiply)*
func (p *parser) addition() Expression {
	return p.binary(p.multiply, PLUS, MINUS)
}

// multiply = negative ((TIMES | DIVIDE | MODULO) negative)*
func (p *parser) multiply() Expression {
	return p.binary(p.negative, TIMES, DIVIDE, MODULO)
}

// negative = MINUS negative | call
func (p *parser) negative() Expression {
	if p.tok == MINUS {
		pos := p.pos
		p.next()
		operand := p.negative()
		return &Unary{pos, MINUS, operand}
	}
	return p.call()
}

// call      = primary (args | subscript | dot)*
// args      = LPAREN RPAREN |
//             LPAREN expression (COMMA expression)* ELLIPSIS? COMMA? RPAREN)
// subscript = LBRACKET expression RBRACKET
// dot       = DOT NAME
func (p *parser) call() Expression {
	expr := p.primary()
	for p.matches(LPAREN, LBRACKET, DOT) {
		if p.tok == LPAREN {
			pos := p.pos
			p.next()
			args := []Expression{}
			gotComma := true
			gotEllipsis := false
			for p.tok != RPAREN && p.tok != EOF && !gotEllipsis {
				if !gotComma {
					p.error("expected , between arguments")
				}
				arg := p.expression()
				args = append(args, arg)
				if p.tok == ELLIPSIS {
					gotEllipsis = true
					p.next()
				}
				if p.tok == COMMA {
					gotComma = true
					p.next()
				} else {
					gotComma = false
				}
			}
			if p.tok != RPAREN && gotEllipsis {
				p.error("can only have ... after last argument")
			}
			p.expect(RPAREN)
			expr = &Call{pos, expr, args, gotEllipsis}
		} else if p.tok == LBRACKET {
			pos := p.pos
			p.next()
			subscript := p.expression()
			p.expect(RBRACKET)
			expr = &Subscript{pos, expr, subscript}
		} else {
			pos := p.pos
			p.next()
			subscript := &Literal{p.pos, p.val}
			p.expect(NAME)
			expr = &Subscript{pos, expr, subscript}
		}
	}
	return expr
}

// primary = NAME | INT | STR | TRUE | FALSE | NIL | list | map |
//           FUNC params block |
//           LPAREN expression RPAREN
func (p *parser) primary() Expression {
	switch p.tok {
	case NAME:
		name := p.val
		pos := p.pos
		p.next()
		return &Variable{pos, name}
	case INT:
		val := p.val
		pos := p.pos
		p.next()
		n, err := strconv.Atoi(val)
		if err != nil {
			// Tokenizer should never give us this
			panic(fmt.Sprintf("tokenizer gave INT token that isn't an int: %s", val))
		}
		return &Literal{pos, n}
	case STR:
		val := p.val
		pos := p.pos
		p.next()
		return &Literal{pos, val}
	case TRUE:
		pos := p.pos
		p.next()
		return &Literal{pos, true}
	case FALSE:
		pos := p.pos
		p.next()
		return &Literal{pos, false}
	case NIL:
		pos := p.pos
		p.next()
		return &Literal{pos, nil}
	case LBRACKET:
		return p.list()
	case LBRACE:
		return p.map_()
	case FUNC:
		pos := p.pos
		p.next()
		args, ellipsis := p.params()
		body := p.block()
		return &FunctionExpression{pos, args, ellipsis, body}
	case LPAREN:
		p.next()
		expr := p.expression()
		p.expect(RPAREN)
		return expr
	default:
		p.error("expected expression, not %s", p.tok)
		return nil
	}
}

// list = LBRACKET RBRACKET |
//        LBRACKET expression (COMMA expression)* COMMA? RBRACKET
func (p *parser) list() Expression {
	pos := p.pos
	p.expect(LBRACKET)
	values := []Expression{}
	gotComma := true
	for p.tok != RBRACKET && p.tok != EOF {
		if !gotComma {
			p.error("expected , between list elements")
		}
		value := p.expression()
		values = append(values, value)
		if p.tok == COMMA {
			gotComma = true
			p.next()
		} else {
			gotComma = false
		}
	}
	p.expect(RBRACKET)
	return &List{pos, values}
}

// map = LBRACE RBRACE |
//       LBRACE expression COLON expression
//              (COMMA expression COLON expression)* COMMA? RBRACE
func (p *parser) map_() Expression {
	pos := p.pos
	p.expect(LBRACE)
	items := []MapItem{}
	gotComma := true
	for p.tok != RBRACE && p.tok != EOF {
		if !gotComma {
			p.error("expected , between map items")
		}
		key := p.expression()
		p.expect(COLON)
		value := p.expression()
		items = append(items, MapItem{key, value})
		if p.tok == COMMA {
			gotComma = true
			p.next()
		} else {
			gotComma = false
		}
	}
	p.expect(RBRACE)
	return &Map{pos, items}
}

// ParseExpression parses a single expression into an Expression interface
// (can be one of many expression types). If the expression parses correctly,
// return an Expression and nil. If there's a syntax error, return nil and
// a parser.Error value.
func ParseExpression(input []byte) (e Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			if parseError, ok := r.(Error); ok {
				err = parseError
			} else {
				panic(r)
			}
		}
	}()
	t := NewTokenizer(input)
	p := parser{tokenizer: t}
	p.next()
	return p.expression(), nil
}

// ParseProgram parses an entire program and returns a *Program (which is
// basically a list of statements). If the program parses correctly, return
// a *Program and nil. If there's a syntax error, return nil and a
// parser.Error value.
func ParseProgram(input []byte) (prog *Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			if parseError, ok := r.(Error); ok {
				err = parseError
			} else {
				panic(r)
			}
		}
	}()
	t := NewTokenizer(input)
	p := parser{tokenizer: t}
	p.next()
	return p.program(), nil
}
