// Test parser package

package parser_test

import (
	"fmt"
	"reflect"
	"testing"

	"littlelang/parser"
)

func TestParseExpression(t *testing.T) {
	tests := []struct {
		source   string
		typeName string
		output   string
		line     int
		column   int
	}{
		// Names and literals
		{"foo", "Variable", "foo", 1, 1},
		{"42", "Literal", "42", 1, 1},
		{`"bar"`, "Literal", `"bar"`, 1, 1},
		{"true", "Literal", "true", 1, 1},
		{"false", "Literal", "false", 1, 1},
		{"nil", "Literal", "nil", 1, 1},

		// Lists
		{"[]", "List", "[]", 1, 1},
		{"[1]", "List", "[1]", 1, 1},
		{"[1,]", "List", "[1]", 1, 1},
		{"[1, 2]", "List", "[1, 2]", 1, 1},
		{"[1, 2,]", "List", "[1, 2]", 1, 1},
		{"[a+b, f(),]", "List", "[(a + b), f()]", 1, 1},
		{"[", "", "expected ] and not EOF", 1, 2},
		{"[1 2", "", "expected , between list elements", 1, 4},
		{"[,]", "", "expected expression, not ,", 1, 2},

		// Maps
		{`{}`, "Map", `{}`, 1, 1},
		{`{"a": 1}`, "Map", `{"a": 1}`, 1, 1},
		{`{x: 1}`, "Map", `{x: 1}`, 1, 1},
		{`{x: 1,}`, "Map", `{x: 1}`, 1, 1},
		{`{x: 1, b: 2}`, "Map", `{x: 1, b: 2}`, 1, 1},
		{`{x: 1, b: 2,}`, "Map", `{x: 1, b: 2}`, 1, 1},
		{`{x + y: 1, "a" + f(): g() / 4,}`, "Map", `{(x + y): 1, ("a" + f()): (g() / 4)}`, 1, 1},
		{`{x, 1}`, "", `expected : and not ,`, 1, 3},
		{`{x: 1: b: 2}`, "", `expected , between map items`, 1, 6},
		{`{`, "", `expected } and not EOF`, 1, 2},
		{`{x: 1 b`, "", `expected , between map items`, 1, 7},
		{`{,}`, "", `expected expression, not ,`, 1, 2},

		// Function expressions
		{"func() {}", "FunctionExpression", "func() {}", 1, 1},
		{"func(a) {}", "FunctionExpression", "func(a) {}", 1, 1},
		{"func(a,) {}", "FunctionExpression", "func(a) {}", 1, 1},
		{"func(a...) {}", "FunctionExpression", "func(a...) {}", 1, 1},
		{"func(a, b) {}", "FunctionExpression", "func(a, b) {}", 1, 1},
		{"func(a, b...) {}", "FunctionExpression", "func(a, b...) {}", 1, 1},
		{"func(a, b...,) {}", "FunctionExpression", "func(a, b...) {}", 1, 1},
		{"func(a, b,) {}", "FunctionExpression", "func(a, b) {}", 1, 1},
		{"func(a, b,) { return 0 }", "FunctionExpression", "func(a, b) {\n    return 0\n}", 1, 1},
		{"func(a: b) {}", "", "expected , between parameters", 1, 7},
		{"func(a..., b) {}", "", "can only have ... after last parameter", 1, 12},
		{"func(,) {}", "", "expected name and not ,", 1, 6},
		{"func(", "", "expected ) and not EOF", 1, 6},

		// Grouping
		{"(1 + 2)", "Binary", "(1 + 2)", 1, 4},
		{"(1 + 2) * 3", "Binary", "((1 + 2) * 3)", 1, 9},
		{"(((1) + 2))", "Binary", "(1 + 2)", 1, 7},
		{"(1 + 2]", "", "expected ) and not ]", 1, 7},
		{"(1 +", "", "expected expression, not EOF", 1, 5},
		{"(1 + 2", "", "expected ) and not EOF", 1, 7},

		// Subscript and dot expressions
		{`a.b`, "Subscript", `a["b"]`, 1, 2},
		{`a.b.c`, "Subscript", `a["b"]["c"]`, 1, 4},
		{`a.b["c"]`, "Subscript", `a["b"]["c"]`, 1, 4},
		{`a["b"].c`, "Subscript", `a["b"]["c"]`, 1, 7},
		{`a["b"]["c"]`, "Subscript", `a["b"]["c"]`, 1, 7},
		{`a.`, "", `expected name and not EOF`, 1, 3},
		{`a.1`, "", `expected name and not int`, 1, 3},
		{`a[...]`, "", `expected expression, not ...`, 1, 3},

		// Function calls
		{"f()", "Call", "f()", 1, 2},
		{"f(a)", "Call", "f(a)", 1, 2},
		{"f(a,)", "Call", "f(a)", 1, 2},
		{"f(a, b)", "Call", "f(a, b)", 1, 2},
		{"f(a, b,)", "Call", "f(a, b)", 1, 2},
		{"f(a...)", "Call", "f(a...)", 1, 2},
		{"f(a...,)", "Call", "f(a...)", 1, 2},
		{"f(a, b...)", "Call", "f(a, b...)", 1, 2},
		{"f(a, b, c...)", "Call", "f(a, b, c...)", 1, 2},
		{"f(,)", "", "expected expression, not ,", 1, 3},
		{"f(a b)", "", "expected , between arguments", 1, 5},
		{"f(a..., b)", "", "can only have ... after last argument", 1, 9},
		{"f(a,", "", "expected ) and not EOF", 1, 5},

		// Negative (unary minus)
		{"-3", "Unary", "(-3)", 1, 1},
		{"--3", "Unary", "(-(-3))", 1, 1},
		{"-(a + b)", "Unary", "(-(a + b))", 1, 1},
		{"-", "", "expected expression, not EOF", 1, 2},

		// Multiplication
		{"1 * 2", "Binary", "(1 * 2)", 1, 3},
		{"1 * 2 * 3", "Binary", "((1 * 2) * 3)", 1, 7},
		{"1 * (2 * 3)", "Binary", "(1 * (2 * 3))", 1, 3},
		{"1 + 2 * 3", "Binary", "(1 + (2 * 3))", 1, 3},
		{"1 * 2 + 3", "Binary", "((1 * 2) + 3)", 1, 7},
		{"1 * -2", "Binary", "(1 * (-2))", 1, 3},
		{"-1 * 2", "Binary", "((-1) * 2)", 1, 4},
		{"1 *", "", "expected expression, not EOF", 1, 4},

		// Division (same precedence as multiplication)
		{"1 / 2", "Binary", "(1 / 2)", 1, 3},
		{"1 / 2 / 3", "Binary", "((1 / 2) / 3)", 1, 7},
		{"1 / (2 / 3)", "Binary", "(1 / (2 / 3))", 1, 3},
		{"1 + 2 / 3", "Binary", "(1 + (2 / 3))", 1, 3},
		{"1 / 2 + 3", "Binary", "((1 / 2) + 3)", 1, 7},
		{"1 / -2", "Binary", "(1 / (-2))", 1, 3},
		{"-1 / 2", "Binary", "((-1) / 2)", 1, 4},
		{"1 /", "", "expected expression, not EOF", 1, 4},

		// Modulo (same precedence as multiplication)
		{"1 % 2", "Binary", "(1 % 2)", 1, 3},
		{"1 % 2 % 3", "Binary", "((1 % 2) % 3)", 1, 7},
		{"1 % (2 % 3)", "Binary", "(1 % (2 % 3))", 1, 3},
		{"1 + 2 % 3", "Binary", "(1 + (2 % 3))", 1, 3},
		{"1 % 2 + 3", "Binary", "((1 % 2) + 3)", 1, 7},
		{"1 % -2", "Binary", "(1 % (-2))", 1, 3},
		{"-1 % 2", "Binary", "((-1) % 2)", 1, 4},
		{"1 %", "", "expected expression, not EOF", 1, 4},

		// Addition
		{"1 + 2", "Binary", "(1 + 2)", 1, 3},
		{"1 + 2 + 3", "Binary", "((1 + 2) + 3)", 1, 7},
		{"1 + (2 + 3)", "Binary", "(1 + (2 + 3))", 1, 3},
		{"1 < 2 + 3", "Binary", "(1 < (2 + 3))", 1, 3},
		{"1 + 2 < 3", "Binary", "((1 + 2) < 3)", 1, 7},
		{"1 +", "", "expected expression, not EOF", 1, 4},

		// Subtraction (same precedence as addition)
		{"1 - 2", "Binary", "(1 - 2)", 1, 3},
		{"1 - 2 - 3", "Binary", "((1 - 2) - 3)", 1, 7},
		{"1 - (2 - 3)", "Binary", "(1 - (2 - 3))", 1, 3},
		{"1 < 2 - 3", "Binary", "(1 < (2 - 3))", 1, 3},
		{"1 - 2 < 3", "Binary", "((1 - 2) < 3)", 1, 7},
		{"1 -", "", "expected expression, not EOF", 1, 4},

		// Less than
		{"1 < 2", "Binary", "(1 < 2)", 1, 3},
		{"1 < 2 < 3", "Binary", "((1 < 2) < 3)", 1, 7},
		{"1 < (2 < 3)", "Binary", "(1 < (2 < 3))", 1, 3},
		{"1 == 2 < 3", "Binary", "(1 == (2 < 3))", 1, 3},
		{"1 < 2 == 3", "Binary", "((1 < 2) == 3)", 1, 7},
		{"1 <", "", "expected expression, not EOF", 1, 4},

		// Less than or equal (same precedence as less than)
		{"1 <= 2", "Binary", "(1 <= 2)", 1, 3},
		{"1 <= 2 <= 3", "Binary", "((1 <= 2) <= 3)", 1, 8},
		{"1 <= (2 <= 3)", "Binary", "(1 <= (2 <= 3))", 1, 3},
		{"1 == 2 <= 3", "Binary", "(1 == (2 <= 3))", 1, 3},
		{"1 <= 2 == 3", "Binary", "((1 <= 2) == 3)", 1, 8},
		{"1 <=", "", "expected expression, not EOF", 1, 5},

		// Greater than (same precedence as less than)
		{"1 > 2", "Binary", "(1 > 2)", 1, 3},
		{"1 > 2 > 3", "Binary", "((1 > 2) > 3)", 1, 7},
		{"1 > (2 > 3)", "Binary", "(1 > (2 > 3))", 1, 3},
		{"1 == 2 > 3", "Binary", "(1 == (2 > 3))", 1, 3},
		{"1 > 2 == 3", "Binary", "((1 > 2) == 3)", 1, 7},
		{"1 >", "", "expected expression, not EOF", 1, 4},

		// Greater than or equal (same precedence as less than)
		{"1 >= 2", "Binary", "(1 >= 2)", 1, 3},
		{"1 >= 2 >= 3", "Binary", "((1 >= 2) >= 3)", 1, 8},
		{"1 >= (2 >= 3)", "Binary", "(1 >= (2 >= 3))", 1, 3},
		{"1 == 2 >= 3", "Binary", "(1 == (2 >= 3))", 1, 3},
		{"1 >= 2 == 3", "Binary", "((1 >= 2) == 3)", 1, 8},
		{"1 >=", "", "expected expression, not EOF", 1, 5},

		// The "in" operator (same precedence as less than)
		{"1 in 2", "Binary", "(1 in 2)", 1, 3},
		{"1 in 2 in 3", "Binary", "((1 in 2) in 3)", 1, 8},
		{"1 in (2 in 3)", "Binary", "(1 in (2 in 3))", 1, 3},
		{"1 == 2 in 3", "Binary", "(1 == (2 in 3))", 1, 3},
		{"1 in 2 == 3", "Binary", "((1 in 2) == 3)", 1, 8},
		{"1 in", "", "expected expression, not EOF", 1, 5},

		// Equals
		{"1 == 2", "Binary", "(1 == 2)", 1, 3},
		{"1 == 2 == 3", "Binary", "((1 == 2) == 3)", 1, 8},
		{"1 == (2 == 3)", "Binary", "(1 == (2 == 3))", 1, 3},
		{"1 == not 2", "", "expected expression, not not", 1, 6},
		{"not 1 == 2", "Unary", "(not (1 == 2))", 1, 1},
		{"1 ==", "", "expected expression, not EOF", 1, 5},

		// Not equals (same precedence as equals)
		{"1 != 2", "Binary", "(1 != 2)", 1, 3},
		{"1 != 2 != 3", "Binary", "((1 != 2) != 3)", 1, 8},
		{"1 != (2 != 3)", "Binary", "(1 != (2 != 3))", 1, 3},
		{"1 != not 2", "", "expected expression, not not", 1, 6},
		{"not 1 != 2", "Unary", "(not (1 != 2))", 1, 1},
		{"1 !=", "", "expected expression, not EOF", 1, 5},

		// Logical not
		{"not 1", "Unary", "(not 1)", 1, 1},
		{"not not 1", "Unary", "(not (not 1))", 1, 1},
		{"not 1 and not 2", "Binary", "((not 1) and (not 2))", 1, 7},
		{"not", "", "expected expression, not EOF", 1, 4},

		// Logical and
		{"1 and 2", "Binary", "(1 and 2)", 1, 3},
		{"1 and 2 and 3", "Binary", "((1 and 2) and 3)", 1, 9},
		{"1 and (2 and 3)", "Binary", "(1 and (2 and 3))", 1, 3},
		{"1 or 2 and 3", "Binary", "(1 or (2 and 3))", 1, 3},
		{"1 and 2 or 3", "Binary", "((1 and 2) or 3)", 1, 9},
		{"1 and", "", "expected expression, not EOF", 1, 6},

		// Logical or
		{"1 or 2", "Binary", "(1 or 2)", 1, 3},
		{"1 or 2 or 3", "Binary", "((1 or 2) or 3)", 1, 8},
		{"1 or (2 or 3)", "Binary", "(1 or (2 or 3))", 1, 3},
		{"1 or", "", "expected expression, not EOF", 1, 5},

		// Crazy expression with everything in it
		{
			`f(1 or 2, 3 and 4, not 5, 6==7, 8!=9, a<b, c<=d, e>f, g>=h, i in j, q*r, s/t, u%v, -w, x(), y.z, A["B"], true, false, nil, [], {})`,
			"Call",
			`f((1 or 2), (3 and 4), (not 5), (6 == 7), (8 != 9), (a < b), (c <= d), (e > f), (g >= h), (i in j), (q * r), (s / t), (u % v), (-w), x(), y["z"], A["B"], true, false, nil, [], {})`,
			1,
			2,
		},

		// Miscellaneous expressions
		{"a + b  // add things", "Binary", "(a + b)", 1, 3},
		{"a\n+\nb", "Binary", "(a + b)", 2, 1},
		{"", "", "expected expression, not EOF", 1, 1},
		{"if true { print(1) }", "", "expected expression, not if", 1, 1},
	}
	for _, test := range tests {
		testName := test.source
		if len(testName) > 20 {
			testName = testName[:20]
		}
		t.Run(testName, func(t *testing.T) {
			expr, err := parser.ParseExpression([]byte(test.source))
			if err != nil {
				parseError, ok := err.(parser.Error)
				if !ok {
					t.Fatalf("unexpected parse error type %T", err)
				}
				if test.typeName != "" {
					t.Fatalf("expected error, got %q", test.typeName)
				}
				output := parseError.Message
				if output != test.output {
					t.Fatalf("expected %q, got %q", test.output, output)
				}
				if parseError.Position.Line != test.line || parseError.Position.Column != test.column {
					t.Fatalf("expected %d:%d, got %d:%d", test.line, test.column,
						parseError.Position.Line, parseError.Position.Column)
				}
			} else {
				exprType := reflect.TypeOf(expr)
				shortName := exprType.String()[8:]
				if shortName != test.typeName {
					t.Fatalf("expected type %q, got %q", test.typeName, shortName)
				}
				exprString := fmt.Sprintf("%s", expr)
				if exprString != test.output {
					t.Fatalf("expected %s, got %s", test.output, exprString)
				}
				if expr.Position().Line != test.line || expr.Position().Column != test.column {
					t.Fatalf("expected %d:%d, got %d:%d", test.line, test.column,
						expr.Position().Line, expr.Position().Column)
				}
			}
		})
	}
}

func TestParseProgram(t *testing.T) {
	tests := []struct {
		source string
		output string
		line   int
		column int
	}{
		// If statements
		{"if a { f() }", `if a {
    f()
}`, 1, 1},
		{"if a { f() g() }", `if a {
    f()
    g()
}`, 1, 1},
		{"if a == 0 { f() }", `if (a == 0) {
    f()
}`, 1, 1},
		{"if a { f() } else { g() }", `if a {
    f()
} else {
    g()
}`, 1, 1},
		{"if a { f() } else if b { g() } else { h() }", `if a {
    f()
} else {
    if b {
        g()
    } else {
        h()
    }
}`, 1, 1},
		{"if a { f() } if b { g() }", `if a {
    f()
}
if b {
    g()
}`, 1, 1},
		{"if a { f() } else while b { g() }", "expected { or if after else, not while", 1, 19},
		{"if a {", "expected } and not EOF", 1, 7},
		{"if", "expected expression, not EOF", 1, 3},

		// While statements
		{"while a { f() }", `while a {
    f()
}`, 1, 1},
		{"while a == 0 { f() g() }", `while (a == 0) {
    f()
    g()
}`, 1, 1},
		{"while a {", "expected } and not EOF", 1, 10},
		{"while", "expected expression, not EOF", 1, 6},

		// For statements
		{"for a in c + d { e() }", `for a in (c + d) {
    e()
}`, 1, 1},
		{"for a in b { c() d() }", `for a in b {
    c()
    d()
}`, 1, 1},
		{"for 12 in a { }", "expected name and not int", 1, 5},
		{"for a if b { }", "expected in and not if", 1, 7},
		{"for a in b c()", "expected { and not name", 1, 12},
		{"for a in b {", "expected } and not EOF", 1, 13},
		{"for", "expected name and not EOF", 1, 4},

		// Return statements (return outside of function is legal according
		// to the parser, but causes a runtime error)
		{"return a", "return a", 1, 1},
		{"func() { return 1 }", `func() {
    return 1
}`, 1, 1},
		{"func() { return a + b }", `func() {
    return (a + b)
}`, 1, 1},
		{"func() { return }", "expected expression, not }", 1, 17},
		{"func() { return if }", "expected expression, not if", 1, 17},

		// Function definitions (function expression is kinda useless at the
		// statement level -- does nothing but is valid syntax)
		{"func() {}", "func() {}", 1, 1},
		{"func add(a, b) { return a + b }", `func add(a, b) {
    return (a + b)
}`, 1, 1},
		{"func go() { print(1) print(2) }", `func go() {
    print(1)
    print(2)
}`, 1, 1},
		{`func outside() {
    func inner() {}
    return inner
}`, `func outside() {
    func inner() {}
    return inner
}`, 1, 1},
		{"func f(a) {}", "func f(a) {}", 1, 1},
		{"func f(a,) {}", "func f(a) {}", 1, 1},
		{"func f(a...) {}", "func f(a...) {}", 1, 1},
		{"func f(a, b) {}", "func f(a, b) {}", 1, 1},
		{"func f(a, b...) {}", "func f(a, b...) {}", 1, 1},
		{"func f(a, b...,) {}", "func f(a, b...) {}", 1, 1},
		{"func f(a, b,) {}", "func f(a, b) {}", 1, 1},
		{"func f(a, b,) { return 0 }", "func f(a, b) {\n    return 0\n}", 1, 1},
		{"func f(a: b) {}", "expected , between parameters", 1, 9},
		{"func f(a..., b) {}", "can only have ... after last parameter", 1, 14},
		{"func f(,) {}", "expected name and not ,", 1, 8},
		{"func f(", "expected ) and not EOF", 1, 8},

		// Assignments
		{"a = 1", "a = 1", 1, 3},
		{"a = 1 b = 2", "a = 1\nb = 2", 1, 3},
		{"a = 1\nb = 2", "a = 1\nb = 2", 1, 3},
		{"a = 1 + 2", "a = (1 + 2)", 1, 3},
		{`x.y = 3`, `x["y"] = 3`, 1, 5},
		{`x["y"] = 3`, `x["y"] = 3`, 1, 8},
		{"(a + b) = 3", "expected name, subscript, or dot expression on left side of =", 1, 9},

		// Comments, expression statements, multiline programs, etc
		{"", "", 0, 0},
		{"// just a comment", "", 0, 0},
		{"// foo\nif a {}", "if a {\n    \n}", 2, 1},
		{"print(1234)", "print(1234)", 1, 1},
		{`
// start
if true {
    print("t")  // foo
}
if false {
    print("f")  // bar
}
// end
`, `if true {
    print("t")
}
if false {
    print("f")
}`, 3, 1},
	}
	for _, test := range tests {
		testName := test.source
		if len(testName) > 40 {
			testName = testName[:40]
		}
		t.Run(testName, func(t *testing.T) {
			prog, err := parser.ParseProgram([]byte(test.source))
			if err != nil {
				parseError, ok := err.(parser.Error)
				if !ok {
					t.Fatalf("unexpected parse error type %T", err)
				}
				output := parseError.Message
				if output != test.output {
					t.Fatalf("expected %q, got %q", test.output, output)
				}
				if parseError.Position.Line != test.line || parseError.Position.Column != test.column {
					t.Fatalf("expected %d:%d, got %d:%d", test.line, test.column,
						parseError.Position.Line, parseError.Position.Column)
				}
			} else {
				progString := fmt.Sprintf("%s", prog)
				if progString != test.output {
					t.Fatalf("expected:\n\"%s\"\ngot:\n\"%s\"", test.output, progString)
				}
				if len(prog.Statements) > 0 {
					stmt := prog.Statements[0]
					if stmt.Position().Line != test.line || stmt.Position().Column != test.column {
						t.Fatalf("expected %d:%d, got %d:%d", test.line, test.column,
							stmt.Position().Line, stmt.Position().Column)
					}
				}
			}
		})
	}
}

func Example_valid() {
	prog, err := parser.ParseProgram([]byte("if true { print(1234) }"))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(prog)
	}
	// Output:
	// if true {
	//     print(1234)
	// }
}

func Example_error() {
	prog, err := parser.ParseProgram([]byte("for if"))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(prog)
	}
	// Output:
	// parse error at 1:5: expected name and not if
}
