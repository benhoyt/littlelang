// Package tokenizer turns a littlelang soure string into a stream of tokent.
//
// To use the tokenizer, create a new tokenizer with NewTokenizer(source) and
// then call Next() until the token type is EOF or ILLEGAL.
//
// This package can be imported using dot syntax, because the exported names
// are more or less unique:
//
//     import (
//         . "littlelang/tokenizer"
//     )
//
// This allows you to access token types using names like RETURN directly
// rather than tokenizer.RETURN.
//
package tokenizer

import (
	"fmt"
	"unicode/utf8"
)

// Token is the type of a single littlelang token
type Token int

const (
	// Stop tokens
	ILLEGAL Token = iota
	EOF

	// Single-character tokens
	ASSIGN
	COLON
	COMMA
	DIVIDE
	DOT
	GT
	LBRACE
	LBRACKET
	LPAREN
	LT
	MINUS
	MODULO
	PLUS
	RBRACE
	RBRACKET
	RPAREN
	TIMES

	// Two-character tokens
	EQUAL
	GTE
	LTE
	NOTEQUAL

	// Three-character tokens
	ELLIPSIS

	// Keywords
	AND
	ELSE
	FALSE
	FOR
	FUNC
	IF
	IN
	NIL
	NOT
	OR
	RETURN
	TRUE
	WHILE

	// Literals and identifiers
	INT
	NAME
	STR
)

var keywordTokens = map[string]Token{
	"and":    AND,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"func":   FUNC,
	"if":     IF,
	"in":     IN,
	"nil":    NIL,
	"not":    NOT,
	"or":     OR,
	"return": RETURN,
	"true":   TRUE,
	"while":  WHILE,
}

var tokenNames = map[Token]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	ASSIGN:   "=",
	COLON:    ":",
	COMMA:    ",",
	DIVIDE:   "/",
	DOT:      ".",
	GT:       ">",
	LBRACE:   "{",
	LBRACKET: "[",
	LPAREN:   "(",
	LT:       "<",
	MINUS:    "-",
	MODULO:   "%",
	PLUS:     "+",
	RBRACE:   "}",
	RBRACKET: "]",
	RPAREN:   ")",
	TIMES:    "*",

	EQUAL:    "==",
	GTE:      ">=",
	LTE:      "<=",
	NOTEQUAL: "!=",

	ELLIPSIS: "...",

	AND:    "and",
	ELSE:   "else",
	FALSE:  "false",
	FOR:    "for",
	FUNC:   "func",
	IF:     "if",
	IN:     "in",
	NIL:    "nil",
	NOT:    "not",
	OR:     "or",
	RETURN: "return",
	TRUE:   "true",
	WHILE:  "while",

	INT:  "int",
	NAME: "name",
	STR:  "str",
}

func (t Token) String() string {
	return tokenNames[t]
}

// Position stores the line and column a token starts at
type Position struct {
	Line   int
	Column int
}

// Tokenizer parses input source code to a stream of tokens. Use
// NewTokenizer() to actually create a tokenizer, and Next() to get the next
// token in the input.
type Tokenizer struct {
	input    []byte
	offset   int
	ch       rune
	errorMsg string
	pos      Position
	nextPos  Position
}

// NewTokenizer returns a new tokenizer that works off the given input.
func NewTokenizer(input []byte) *Tokenizer {
	t := new(Tokenizer)
	t.input = input
	t.nextPos.Line = 1
	t.nextPos.Column = 1
	t.next()
	return t
}

func (t *Tokenizer) next() {
	t.pos = t.nextPos
	ch, size := utf8.DecodeRune(t.input[t.offset:])
	if size == 0 {
		t.ch = -1
		return
	}
	if ch == utf8.RuneError {
		t.ch = -1
		t.errorMsg = fmt.Sprintf("invalid UTF-8 byte 0x%02x", t.input[t.offset])
		return
	}
	if ch == '\n' {
		t.nextPos.Line++
		t.nextPos.Column = 1
	} else {
		t.nextPos.Column++
	}
	t.ch = ch
	t.offset += size
}

func (t *Tokenizer) skipWhitespaceAndComments() {
	for {
		for t.ch == ' ' || t.ch == '\t' || t.ch == '\r' || t.ch == '\n' {
			t.next()
		}
		if !(t.ch == '/' && t.offset < len(t.input) && t.input[t.offset] == '/') {
			break
		}
		// Skip //-prefixed comment (to end of line or end of input)
		t.next()
		t.next()
		for t.ch != '\n' && t.ch >= 0 {
			t.next()
		}
		t.next()
	}
}

func isNameStart(ch rune) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// Next() returns the position, token type, and token value of the next token
// in the source. For ordinary tokens, the token value is empty. For INT,
// NAME, and STR tokens, it's the number or string value. For an ILLEGAL
// token, it's the error message.
func (t *Tokenizer) Next() (Position, Token, string) {
	t.skipWhitespaceAndComments()
	if t.ch < 0 {
		if t.errorMsg != "" {
			return t.pos, ILLEGAL, t.errorMsg
		}
		return t.pos, EOF, ""
	}

	pos := t.pos
	token := ILLEGAL
	value := ""

	ch := t.ch
	t.next()

	// Names (identifiers) and keywords
	if isNameStart(ch) {
		runes := []rune{ch}
		for isNameStart(t.ch) || (t.ch >= '0' && t.ch <= '9') {
			runes = append(runes, t.ch)
			t.next()
		}
		name := string(runes)
		token, isKeyword := keywordTokens[name]
		if !isKeyword {
			token = NAME
			value = name
		}
		return pos, token, value
	}

	switch ch {
	case ':':
		token = COLON
	case ',':
		token = COMMA
	case '/':
		token = DIVIDE
	case '{':
		token = LBRACE
	case '[':
		token = LBRACKET
	case '(':
		token = LPAREN
	case '-':
		token = MINUS
	case '%':
		token = MODULO
	case '+':
		token = PLUS
	case '}':
		token = RBRACE
	case ']':
		token = RBRACKET
	case ')':
		token = RPAREN
	case '*':
		token = TIMES

	case '=':
		if t.ch == '=' {
			t.next()
			token = EQUAL
		} else {
			token = ASSIGN
		}
	case '!':
		if t.ch == '=' {
			t.next()
			token = NOTEQUAL
		} else {
			token = ILLEGAL
			value = fmt.Sprintf("expected != instead of !%c", t.ch)
		}
	case '<':
		if t.ch == '=' {
			t.next()
			token = LTE
		} else {
			token = LT
		}
	case '>':
		if t.ch == '=' {
			t.next()
			token = GTE
		} else {
			token = GT
		}

	case '.':
		if t.ch == '.' {
			t.next()
			if t.ch != '.' {
				return pos, ILLEGAL, "unexpected .."
			}
			t.next()
			token = ELLIPSIS
		} else {
			token = DOT
		}

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		runes := []rune{ch}
		for t.ch >= '0' && t.ch <= '9' {
			runes = append(runes, t.ch)
			t.next()
		}
		token = INT
		value = string(runes)

	case '"':
		runes := []rune{}
		for t.ch != '"' {
			c := t.ch
			if c < 0 {
				return pos, ILLEGAL, "didn't find end quote in string"
			}
			if c == '\r' || c == '\n' {
				return pos, ILLEGAL, "can't have newline in string"
			}
			if c == '\\' {
				t.next()
				switch t.ch {
				case '"', '\\':
					c = t.ch
				case 't':
					c = '\t'
				case 'r':
					c = '\r'
				case 'n':
					c = '\n'
				default:
					return pos, ILLEGAL, fmt.Sprintf("invalid string escape \\%c", t.ch)
				}
			}
			runes = append(runes, c)
			t.next()
		}
		t.next()
		token = STR
		value = string(runes)

	default:
		token = ILLEGAL
		value = fmt.Sprintf("unexpected %c", ch)
	}
	return pos, token, value
}
