// Test tokenizer package

package tokenizer_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	. "github.com/benhoyt/littlelang/tokenizer"
)

type Info struct {
	Line   int
	Column int
	Token  Token
	Value  string
}

func tokenize(input string) ([]Info, error) {
	k := NewTokenizer([]byte(input))
	infos := []Info{}
	var err error
	for {
		pos, token, value := k.Next()
		if token == EOF {
			eofLine := strings.Count(input, "\n") + 1

			var eofColumn int
			lastNewline := strings.LastIndex(input, "\n")
			if lastNewline < 0 {
				eofColumn = utf8.RuneCountInString(input) + 1
			} else {
				eofColumn = utf8.RuneCountInString(input[lastNewline+1:]) + 1
			}

			if pos.Line != eofLine || pos.Column != eofColumn {
				err = fmt.Errorf("expected EOF at %d:%d, got %d:%d", eofLine, eofColumn, pos.Line, pos.Column)
			}
			break
		}
		infos = append(infos, Info{pos.Line, pos.Column, token, value})
		if token == ILLEGAL {
			break
		}
	}
	return infos, err
}

func infosEqual(output []Info, expected []Info) string {
	for i := 0; i < len(output) && i < len(expected); i++ {
		if output[i] != expected[i] {
			return fmt.Sprintf("token %d: got %v instead of %v", i, output[i], expected[i])
		}
	}
	if len(output) < len(expected) {
		return fmt.Sprintf("got %d too few tokens: %v", len(expected)-len(output), expected[len(output):])
	}
	if len(output) > len(expected) {
		return fmt.Sprintf("got %d too many tokens: %v", len(output)-len(expected), output[len(expected):])
	}
	return ""
}

func tokenStrings(input string) string {
	output := ""
	k := NewTokenizer([]byte(input))
	for {
		_, token, _ := k.Next()
		if token == EOF {
			break
		}
		output += token.String() + " "
		if token == ILLEGAL {
			break
		}
	}
	return output
}

func TestAll(t *testing.T) {
	tests := []struct {
		input  string
		output []Info
	}{
		{"", []Info{}},
		{"  \n  \n", []Info{}},
		{"/", []Info{
			{1, 1, DIVIDE, ""},
		}},
		{"//", []Info{}},
		{"///", []Info{}},
		{"/ //\n/", []Info{
			{1, 1, DIVIDE, ""},
			{2, 1, DIVIDE, ""},
		}},
		{"// foo", []Info{}},
		{"// foo\n1", []Info{
			{2, 1, INT, "1"},
		}},
		{"# foo", []Info{
			{1, 1, ILLEGAL, "unexpected #"},
		}},
		{" \"foo", []Info{
			{1, 2, ILLEGAL, "didn't find end quote in string"},
		}},
		{"\x80", []Info{
			{1, 1, ILLEGAL, "invalid UTF-8 byte 0x80"},
		}},
		{"1234 0 42 -42 1234x 0x321", []Info{
			{1, 1, INT, "1234"},
			{1, 6, INT, "0"},
			{1, 8, INT, "42"},
			{1, 11, MINUS, ""},
			{1, 12, INT, "42"},
			{1, 15, INT, "1234"},
			{1, 19, NAME, "x"},
			{1, 21, INT, "0"},
			{1, 22, NAME, "x321"},
		}},
		{`"foo" "'" "\"" "x\"y" "\t\r\n" "\\" "\z"`, []Info{
			{1, 1, STR, "foo"},
			{1, 7, STR, `'`},
			{1, 11, STR, `"`},
			{1, 16, STR, `x"y`},
			{1, 23, STR, "\t\r\n"},
			{1, 32, STR, `\`},
			{1, 37, ILLEGAL, `invalid string escape \z`},
		}},
		{"\"\n\"", []Info{
			{1, 1, ILLEGAL, "can't have newline in string"},
		}},
		{"1 + 2  // comment", []Info{
			{1, 1, INT, "1"},
			{1, 3, PLUS, ""},
			{1, 5, INT, "2"},
		}},
		{"func() {\n    return a+b\n}", []Info{
			{1, 1, FUNC, ""},
			{1, 5, LPAREN, ""},
			{1, 6, RPAREN, ""},
			{1, 8, LBRACE, ""},
			{2, 5, RETURN, ""},
			{2, 12, NAME, "a"},
			{2, 13, PLUS, ""},
			{2, 14, NAME, "b"},
			{3, 1, RBRACE, ""},
		}},
		{"_ __ _a _A _0 a0 0a Abc a_b", []Info{
			{1, 1, NAME, "_"},
			{1, 3, NAME, "__"},
			{1, 6, NAME, "_a"},
			{1, 9, NAME, "_A"},
			{1, 12, NAME, "_0"},
			{1, 15, NAME, "a0"},
			{1, 18, INT, "0"},
			{1, 19, NAME, "a"},
			{1, 21, NAME, "Abc"},
			{1, 25, NAME, "a_b"},
		}},
		{"and else false for func if in nil not or return true", []Info{
			{1, 1, AND, ""},
			{1, 5, ELSE, ""},
			{1, 10, FALSE, ""},
			{1, 16, FOR, ""},
			{1, 20, FUNC, ""},
			{1, 25, IF, ""},
			{1, 28, IN, ""},
			{1, 31, NIL, ""},
			{1, 35, NOT, ""},
			{1, 39, OR, ""},
			{1, 42, RETURN, ""},
			{1, 49, TRUE, ""},
		}},
		{"= == != < <= > >= !!", []Info{
			{1, 1, ASSIGN, ""},
			{1, 3, EQUAL, ""},
			{1, 6, NOTEQUAL, ""},
			{1, 9, LT, ""},
			{1, 11, LTE, ""},
			{1, 14, GT, ""},
			{1, 16, GTE, ""},
			{1, 19, ILLEGAL, "expected != instead of !!"},
		}},
		{"+-*/% ()[]{}:, . ... .... @", []Info{
			{1, 1, PLUS, ""},
			{1, 2, MINUS, ""},
			{1, 3, TIMES, ""},
			{1, 4, DIVIDE, ""},
			{1, 5, MODULO, ""},
			{1, 7, LPAREN, ""},
			{1, 8, RPAREN, ""},
			{1, 9, LBRACKET, ""},
			{1, 10, RBRACKET, ""},
			{1, 11, LBRACE, ""},
			{1, 12, RBRACE, ""},
			{1, 13, COLON, ""},
			{1, 14, COMMA, ""},
			{1, 16, DOT, ""},
			{1, 18, ELLIPSIS, ""},
			{1, 22, ELLIPSIS, ""},
			{1, 25, DOT, ""},
			{1, 27, ILLEGAL, "unexpected @"},
		}},
	}
	for _, test := range tests {
		output, err := tokenize(test.input)
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		unequalMsg := infosEqual(output, test.output)
		if unequalMsg != "" {
			t.Errorf("%q: %s", test.input, unequalMsg)
		}
	}
}

func TestString(t *testing.T) {
	output := tokenStrings(`
and else false for func if in nil not or return true while
+-*/% ()[]{}:, . ...
1234 "foo" abc
@
`)
	expected := "and else false for func if in nil not or return true while + - * / % ( ) [ ] { } : , . ... int str name ILLEGAL "
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func Example() {
	tokenizer := NewTokenizer([]byte(`print(1234, "foo") @`))
	for {
		pos, tok, val := tokenizer.Next()
		if tok == EOF {
			break
		}
		fmt.Printf("%d:%d %s %q\n", pos.Line, pos.Column, tok, val)
		if tok == ILLEGAL {
			break
		}
	}
	// Output:
	// 1:1 name "print"
	// 1:6 ( ""
	// 1:7 int "1234"
	// 1:11 , ""
	// 1:13 str "foo"
	// 1:18 ) ""
	// 1:20 ILLEGAL "unexpected @"
}
