// Test interpreter package

// If you want to run tests against both the Go version of the interpreter
// (default) and the littlelang version, use a command line like:
//
// go test littlelang/interpreter -exe ~/go/src/littlelang/littlelang -interp ~/go/src/littlelang/littlelang.ll

package interpreter_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/benhoyt/littlelang/interpreter"
	"github.com/benhoyt/littlelang/parser"
)

var (
	exePath    string
	interpPath string
)

func TestMain(m *testing.M) {
	flag.StringVar(&exePath, "exe", "", "path to Go littlelang interpreter binary")
	flag.StringVar(&interpPath, "interp", "", "path to littlelang.ll")
	flag.Parse()
	os.Exit(m.Run())
}

func TestExecute(t *testing.T) {
	tests := []struct {
		source string
		errpos string
		output string
	}{
		// Miscellaneous inputs
		{``, "", ``},

		// == binary operator
		{`print(nil==nil, nil==true, nil==false, nil==0, nil==1, nil=="", nil=="foo", nil==[], nil==[1], nil=={}, nil=={"a": 1})`, "",
			`true false false false false false false false false false false`},
		{`print(true==nil, true==true, true==false, true==0, true==1, true=="", true=="foo", true==[], true==[1], true=={}, true=={"a": 1})`, "",
			`false true false false false false false false false false false`},
		{`print(false==nil, false==true, false==false, false==0, false==1, false=="", false=="foo", false==[], false==[1], false=={}, false=={"a": 1})`, "",
			`false false true false false false false false false false false`},
		{`print(nil==nil, nil==true, nil==false, nil==0, nil==1, nil=="", nil=="foo", nil==[], nil==[1], nil=={}, nil=={"a": 1})`, "",
			`true false false false false false false false false false false`},
		{`print(0==nil, 0==true, 0==false, 0==0, 0==1, 0=="", 0=="foo", 0==[], 0==[1], 0=={}, 0=={"a": 1})`, "",
			`false false false true false false false false false false false`},
		{`print(1==nil, 1==true, 1==false, 1==0, 1==1, 1=="", 1=="foo", 1==[], 1==[1], 1=={}, 1=={"a": 1})`, "",
			`false false false false true false false false false false false`},
		{`print(1234==1234, 1234==4321, 0==-1, 0==0, 1==0, 0==1, 1==1)`, "",
			`true false false true false false true`},
		{`print(""=="", ""=="foo", "foo"=="", "foo"=="foo", "Foo"=="foo", "foo"=="bar")`, "",
			`true false false true false false`},
		{`print([]==[], []==[nil], [1]==[0], [1]==[1], [{"foo": 1}]==[{"foo": 1}], [["bar"], 1]==[["foo"], 1])`, "",
			`true false false true true false`},
		{`x = []  y = []  print(x==y)  append(y, 42)  print(x==y)  append(x, 42)  print(x==y)`, "",
			"true\nfalse\ntrue"},
		{`print({}=={}, {}=={"a": nil}, {"a": 1}=={"b": 2}, {"a": 1}=={"a": 1}, {"a": [1]}=={"a": [1]}, {"a": [1]}=={"a": [2]})`, "",
			`true false false true true false`},
		{`x = {}  y = {}  print(x==y)  y.a=42  print(x==y)  x.a=42  print(x==y)`, "",
			"true\nfalse\ntrue"},
		{`func f() {}  func g() {}  print(f==g, f==f, g==g)`, "", `false true true`},

		// "in" binary operator
		{`print("foo" in "foobar", "foo" in "bar", "" in "", "" in "foo", "foo" in "Foobar")`, "",
			`true false true true false`},
		{`1234 in "foo"`, "type error at 1:6", "in str requires str on left side"},
		{`"foo" in 1234`, "type error at 1:7", "in requires str, list, or map on right side"},
		{`print(nil in [], nil in [nil], 1 in [], 1 in [1], 1 in [1, 1, 1], 1 in [0, 1, 2], [1] in [0, 1, 2], [1] in [0, [1], 2])`, "",
			`false true false true true true false true`},
		{`print(1234 in {})`, "type error at 1:12", "in map requires str on left side"},
		{`print("" in {}, "" in {"": 1}, "a" in {}, "a" in {"a": 1}, "a" in {"b": 2, "a": 1}, "a" in {"A": 1, "B": []})`, "",
			`false true false true true false`},

		// comparison binary operators
		{`print(nil < "")`, "type error at 1:11", "comparison requires two ints or two strs (or lists of ints or strs)"},
		{`print(1 < "foo")`, "type error at 1:9", "comparison requires two ints or two strs (or lists of ints or strs)"},
		{`print(0 < 1, 1 < 1234, 1 < 1, 1 < 2, 0 < 0, -1 < 0, -1 < 1, 1 < -1)`, "",
			`true true false true false true true false`},
		{`print("a" < "b", "foo" < "foo", "foo" < "foobar", "foo" < "Foo", "bar" < "foo", "foo" < "bar", "abc" < "defghi")`, "",
			`true false true false true false true`},
		{`print([] < [], [1] < [1, 2], [1, 2] < [1], [[1], [2]] < [[1], [3]])`, "",
			`false true false true`},
		{`print(1 <= 0, 1 <= 1, 1 <= 2)`, "", "false true true"},
		{`print(1 > 0, 1 > 1, 1 > 2)`, "", "true false false"},
		{`print(1 >= 0, 1 >= 1, 1 >= 2)`, "", "true true false"},

		// + binary operator
		{`print(1 + 2, -3 + 4, 3 + -4, 1 + 2*3, (1+2)*3)`, "", "3 1 -1 7 9"},
		{`print(1 + "foo")`, "type error at 1:9", "+ requires two ints, strs, lists, or maps"},
		{`s="foo"  print(s + "bar", s)`, "", "foobar foo"},
		{`x=[1, 2]  y=[3, 4]  print(x+y, x, y)`, "", "[1, 2, 3, 4] [1, 2] [3, 4]"},
		{`x={"a": 1}  y={"b": 2}  print(x+y, x, y)`, "", `{"a": 1, "b": 2} {"a": 1} {"b": 2}`},
		{`print({"a": 1} + {"a": 2, "b": 3})`, "", `{"a": 2, "b": 3}`},

		// - binary operator
		{`print(1 - 2, -3 - 4, 3 - -4)`, "", "-1 -7 7"},
		{`print(1 - "foo")`, "type error at 1:9", "- requires two ints"},

		// * binary operator
		{`print(2 * 3, 3 * 4, -1 * 7, 3 * -4)`, "", "6 12 -7 -12"},
		{`print(3 * "foo", "ba" * 3)`, "", "foofoofoo bababa"},
		{`lst=[1,2]  print([]*3, lst*3, 3*lst)`, "", "[] [1, 2, 1, 2, 1, 2] [1, 2, 1, 2, 1, 2]"},
		{`print(1 * true)`, "type error at 1:9", "* requires two ints or a str or list and an int"},

		// / binary operator
		{`print(9 / 3, 10 / 3, 10 / 2, 10 / -2, -10 / 2)`, "", "3 3 5 -5 -5"},
		{`print(1 / "foo")`, "type error at 1:9", "/ requires two ints"},
		{`print(3 / 0)`, "value error at 1:9", "can't divide by zero"},

		{`print(9 % 3, 10 % 3, 10 % -3, -10 % 3)`, "", "0 1 1 -1"},
		{`print(1 % "foo")`, "type error at 1:9", "% requires two ints"},
		{`print(3 % 0)`, "value error at 1:9", "can't divide by zero"},

		// Unary operators
		{`print(not true, not false, not not true, not 1==0)`, "", "false true true true"},
		{`print(not nil)`, "type error at 1:7", "not requires a bool"},
		{`print(-3, --4, ---4, -0)`, "", "-3 4 -4 0"},
		{`print(-"foo")`, "type error at 1:7", "unary - requires an int"},

		// Logical and
		{`print(print("a") == nil and print("b") == nil)`, "", "a\nb\ntrue"},
		{`print(print("a") == nil and print("b") != nil)`, "", "a\nb\nfalse"},
		{`print(print("a") != nil and print("b") == nil)`, "", "a\nfalse"},
		{`print(print("a") != nil and print("b") != nil)`, "", "a\nfalse"},

		// Logical or
		{`print(print("a") == nil or print("b") == nil)`, "", "a\ntrue"},
		{`print(print("a") == nil or print("b") != nil)`, "", "a\ntrue"},
		{`print(print("a") != nil or print("b") == nil)`, "", "a\nb\ntrue"},
		{`print(print("a") != nil or print("b") != nil)`, "", "a\nb\nfalse"},

		// Subscript
		{`s = "foo"  print(s[0], s[1], s[2])`, "", "f o o"},
		{`s = "“smart quotes”"  print([s[0], s[1], s[2], s[3]])`, "", `["\xe2", "\x80", "\x9c", "s"]`},
		{`s = "foo"  print(s[-1])`, "value error at 1:20", "subscript -1 out of range"},
		{`s = "foo"  print(s[3])`, "value error at 1:20", "subscript 3 out of range"},
		{`s = "foo"  print(s[nil])`, "type error at 1:20", "str subscript must be an int"},
		{`lst = [1,2,3]  print(lst[0], lst[1], lst[2])`, "", "1 2 3"},
		{`lst = [1,2,3]  print(lst[-1])`, "value error at 1:26", "subscript -1 out of range"},
		{`lst = [1,2,3]  print(lst[3])`, "value error at 1:26", "subscript 3 out of range"},
		{`lst = [1,2,3]  print(lst[nil])`, "type error at 1:26", "list subscript must be an int"},
		{`m = {"a": 1, "b": 2}  print(m["a"], m.a, m["b"], m.b)`, "", `1 1 2 2`},
		{`m = {"a": 1, "b": 2}  print(m["x"])`, "value error at 1:31", `key not found: "x"`},
		{`m = {"a": 1, "b": 2}  print(m[1])`, "type error at 1:31", `map subscript must be a str`},

		// Function calls
		{`print(print(1), print(2))`, "", "1\n2\nnil nil"},
		{`f = print  f()  f(1)  f(1, 2)`, "", "\n1\n1 2"},
		{`func add(a, b) { return a+b }  print(add(2, 7))`, "", "9"},
		{`n = func(){ return 1 + 2 }()  print(n)`, "", "3"},
		{`print(1, 2, [3, 4])`, "", "1 2 [3, 4]"},
		{`print(1, 2, [3, 4]...)`, "", "1 2 3 4"},
		{`print(nil, 0, true, false, "s", [1, 2], {"a": 3})`, "", `nil 0 true false s [1, 2] {"a": 3}`},
		{`print([]...)`, "", ""},
		{`print([1]...)`, "", "1"},
		{`x = [1, 2, 3]  print(x...)`, "", "1 2 3"},
		{`x=0  func f() { x=1 }  f()  print(x)`, "", "0"},
		{`x=[0]  func f() { x[0]=1 }  f()  print(x[0])`, "", "1"},
		{`
func make_adder(n) {
    func adder(x) {
        return x + n
    }
    return adder
}
add5 = make_adder(5)
add3 = make_adder(3)
print(add5(1), add5(2), add3(10), add3(20))
`, "", "6 7 13 23"},
		{`
func make_counter() {
    i = [0]
    func count() {
        i[0] = i[0] + 1
        print(i[0])
    }
    return count
}
counter = make_counter()
counter()
counter()
counter()
`, "", "1\n2\n3"},
		{`f = 1234  f()`, "type error at 1:11", "can't call non-function type int"},
		{`func add(nums...) { sum = 0  for n in nums { sum = sum + n }  return sum }  print(add(), add(42), add(3, 4, 5), add(range(10)...))`, "",
			"0 42 12 45"},

		// Literals
		{`print(1234)`, "", `1234`},
		{`print("foo")`, "", `foo`},
		{`print(true)`, "", `true`},
		{`print(false)`, "", `false`},
		{`print(nil)`, "", `nil`},
		{`print([1,2,3], {"a": 1, "b": 2})`, "", `[1, 2, 3] {"a": 1, "b": 2}`},

		// Variables
		{`a=1  b=2  a=a+b+1  print(a, b)`, "", "4 2"},
		{`asdf`, "name error at 1:1", `name "asdf" not found`},
		{`func f() { return a }  f()`, "name error at 1:19", `name "a" not found`},
		{`func f() { return a }  a=42  print(f())`, "", `42`},

		// Function expression
		{`print(func() {})`, "", "<func>"},
		{`n = ["z", "A", "b", "a"]  sort(n, func(x) { return lower(x) })  print(n)`, "", `["A", "a", "b", "z"]`},
		{`a=40  b=2  func foo() { return func() { return a+b } }  print(foo()())`, "", "42"},

		// Assign
		{`x = 4  print(x)`, "", "4"},
		{`x = 4  func f() { x = 8  print(x) }  print(x)  f()  print(x)`, "", "4\n8\n4"},
		{`func add(a, b) { a = a  b = b  return a + b }  print(add(3, 4))`, "", "7"},
		{`func f() { x = 4}  print(x)`, "name error at 1:26", `name "x" not found`},
		{`x = [1,2,3]  x[0] = 3  x[2] = 1  print(x)`, "", "[3, 2, 1]"},
		{`x = [1,2,3]  x[-1]`, "value error at 1:16", "subscript -1 out of range"},
		{`x = [1,2,3]  x[3]`, "value error at 1:16", "subscript 3 out of range"},
		{`x = [1,2,3]  x["a"]`, "type error at 1:16", "list subscript must be an int"},
		{`m = {"a": 1}  m["a"] = 2  m.b = 3  print(m)`, "", `{"a": 2, "b": 3}`},
		{`m = {"a": 1}  m[0] = 2`, "type error at 1:17", `map subscript must be a str`},
		{`lst = [1,2,3]  func f() { return lst }  func g() { return 1 }  f()[g()] = 2+2+2  print(lst)`, "", `[1, 6, 3]`},
		{`n = 1234  n[0] = 42`, "type error at 1:13", "can only assign to subscript of list or map"},

		// If
		{`if true { print(1) }`, "", "1"},
		{`if false { print(1) }`, "", ""},
		{`if true { print(1) } else { print(0) }`, "", "1"},
		{`if false { print(1) } else { print(0) }`, "", "0"},
		{`if 1==0 { print(1) } else if 0==1 { print(2) } else { print(3) }`, "", "3"},
		{`if 1234 { print(1) }`, "type error at 1:4", "if condition must be bool, got int"},

		// While
		{`i = 0  while i < 5 { print(i)  i=i+1 }  print("DONE", i)`, "", "0\n1\n2\n3\n4\nDONE 5"},
		{`print("S")  while false { print("hi") }  print("F")`, "", "S\nF"},

		// For
		{`i="foo"  for i in range(5) { print(i) }  print(i)`, "", "0\n1\n2\n3\n4\n4"},
		{`i="foo"  for i in range(5) { print(i) }  print(i)`, "", "0\n1\n2\n3\n4\n4"},
		{`s = "“foo”"  for c in s { print(c) }  print(c)`, "", "“\nf\no\no\n”\n”"},
		{`lst = [1,2,3]  for x in lst { print(x) }  print(lst)`, "", "1\n2\n3\n[1, 2, 3]"},
		{`lst = []  for x in lst { print(x) }  print(lst)`, "", "[]"},
		{`m = {"a": 1, "b": 2}  keys = []  for k in m { append(keys, k) }  sort(keys)  print(keys)`, "",
			`["a", "b"]`},
		{`for x in {"a": 1} { print(x) }`, "", "a"},
		{`for x in {} { print(x) }`, "", ""},

		// ExpressionStatement
		{`1234  print("x")  4321  print(print)`, "", "x\n<builtin print>"},

		// append() builtin
		{`x=[0]  append(x, 1)  append(x, 2, 3, 4)  print(x)`, "", `[0, 1, 2, 3, 4]`},
		{`x=[0]  y=[1,2,3]  append(x, y)  print(x, y)`, "", `[0, [1, 2, 3]] [1, 2, 3]`},
		{`x=[0]  y=[1,2,3]  append(x, y...)  print(x, y)`, "", `[0, 1, 2, 3] [1, 2, 3]`},
		{`x=[0]  y=[]  append(x, y...)  print(x, y)`, "", `[0] []`},
		{`x=[0]  append(x)  print(x)`, "", `[0]`},
		{`x=0  append(x, 1234)`, "type error at 1:6", `append() requires first argument to be list`},

		// args() builtin
		{`print(args())`, "", `["one", "2", "THREE"]`},
		{`args(1)`, "type error at 1:1", "args() requires 0 args, got 1"},

		// char() builtin
		{`print(char(123))`, "", `{`},
		{`print(char(8220))`, "", `“`},
		{`char(1, 2)`, "type error at 1:1", "char() requires 1 arg, got 2"},
		{`char("x")`, "type error at 1:1", "char() requires an int, not str"},

		// exit() builtin
		// Skip these for now as they exit the littlelang.ll version:
		// {`exit()`, "", "exit(0)"},
		// {`exit(42)`, "", "exit(42)"},
		{`exit(1, 2)`, "type error at 1:1", "exit() requires 0 or 1 args, got 2"},
		{`exit("x")`, "type error at 1:1", "exit() requires an int, not str"},

		// find() builtin
		{`print(find("", ""), find("", "foo"), find("foo", ""), find("foo", "foo"), find("foo", "o"), find("foz", "z"), find("foo", "bar"))`, "", "0 -1 0 0 1 2 -1"},
		{`find("foo", 1)`, "type error at 1:1", "find() on str requires second argument to be a str"},
		{`print(find([1,2,3], 2), find([1,2,3], 1), find([1,2,3], 3), find([1,2,3], 4), find([], 0))`, "", "1 0 2 -1 -1"},
		{`print(find([[1], [2], [3]], [2]), find([[1], [2], [3]], 2))`, "", "1 -1"},
		{`print(find([1, 2, 3], nil), find([1, nil, 3], nil))`, "", "-1 1"},
		{`print(find())`, "type error at 1:7", "find() requires 2 args, got 0"},
		{`print(find(1234, 1))`, "type error at 1:7", "find() requires first argument to be a str or list"},

		// int() builtin
		{`print(int(1234), type(int(1234)))`, "", "1234 int"},
		{`print(int("1234"), type(int("1234")))`, "", "1234 int"},
		{`print(int("abc"), type(int("abc")))`, "", "nil nil"},
		{`print(int(nil))`, "type error at 1:7", "int() requires an int or a str"},
		{`print(int())`, "type error at 1:7", "int() requires 1 arg, got 0"},

		// join() builtin
		{`print(join(["abc", "de", "f", "", "."], "|"))`, "", "abc|de|f||."},
		{`print(join(["abc", "de", "f", "", "."], ""))`, "", "abcdef."},
		{`print(join([], "|"))`, "", ""},
		{`print(join([], ""))`, "", ""},
		{`print(join(["x", 1], ""))`, "type error at 1:7", "join() requires all list elements to be strs"},
		{`print(join("", ""))`, "type error at 1:7", "join() requires first argument to be a list"},
		{`print(join())`, "type error at 1:7", "join() requires 2 args, got 0"},

		// len() builtin
		{`print(len("foo"), len("“smart quotes”"), len(""))`, "", "3 18 0"},
		{`print(len([]), len([1, 2, 3]))`, "", "0 3"},
		{`print(len({}), len({"a": 1, "b": 2, "c": 3}))`, "", "0 3"},
		{`print(len(42))`, "type error at 1:7", "len() requires a str, list, or map"},
		{`print(len())`, "type error at 1:7", "len() requires 1 arg, got 0"},

		// lower() builtin
		{`print(lower(""), lower("abc"), lower("FoO"), lower("BAR"))`, "", " abc foo bar"},
		{`print(lower(42))`, "type error at 1:7", "lower() requires a str"},
		{`print(lower())`, "type error at 1:7", "lower() requires 1 arg, got 0"},

		// print() builtin
		{`print()  print("foo")  print("x", 42)  print([1, 2, 3]...)`, "", "\nfoo\nx 42\n1 2 3"},
		{`print(nil, true, false, 1, "x", ["y"], {"z": 2}, func() {})`, "", `nil true false 1 x ["y"] {"z": 2} <func>`},

		// range() builtin
		{`print(range(0), range(5))`, "", "[] [0, 1, 2, 3, 4]"},
		{`range(-1)`, "value error at 1:1", "range() argument must not be negative"},
		{`range(nil)`, "type error at 1:1", "range() requires an int"},

		// read() builtin
		{`print(read())`, "", "dummy stdin"},
		{`read(1)`, "type error at 1:1", "read() argument must be a str"},
		{`read("x", "y")`, "type error at 1:1", "read() requires 0 or 1 args, got 2"},

		// rune() builtin
		{`print(rune("A"), rune(" "), rune("“"))`, "", "65 32 8220"},
		{`print(rune(42))`, "type error at 1:7", "rune() requires a str"},
		{`print(rune("ab"))`, "value error at 1:7", "rune() requires a 1-character str"},
		{`print(rune())`, "type error at 1:7", "rune() requires 1 arg, got 0"},

		// slice() builtin
		{`print(slice("abc", 0, 3), slice("abc", 1, 3), slice("abc", 0, 2))`, "", "abc bc ab"},
		{`print(slice("foo", 0, 0), slice("", 0, 0), slice("“", 0, 3))`, "", "  “"},
		{`print(slice([1,2,3], 0, 3), slice([1,2,3], 1, 3), slice([1,2,3], 0, 2))`, "", "[1, 2, 3] [2, 3] [1, 2]"},
		{`x=[1,2,3]  y=slice(x, 0, 1)  print(x, y)  y[0]=4  print(x, y)`, "", "[1, 2, 3] [1]\n[1, 2, 3] [4]"},
		{`slice("foo", -1, 0)`, "value error at 1:1", "slice() start or end out of bounds"},
		{`slice("foo", 3, 1)`, "value error at 1:1", "slice() start or end out of bounds"},
		{`slice("foo", 1, 4)`, "value error at 1:1", "slice() start or end out of bounds"},
		{`slice([1,2,3], -1, 0)`, "value error at 1:1", "slice() start or end out of bounds"},
		{`slice([1,2,3], 3, 1)`, "value error at 1:1", "slice() start or end out of bounds"},
		{`slice([1,2,3], 1, 4)`, "value error at 1:1", "slice() start or end out of bounds"},
		{`print(slice(42, 0, 0))`, "type error at 1:7", "slice() requires first argument to be a str or list"},
		{`print(slice("x", 0, "z"))`, "type error at 1:7", "slice() requires start and end to be ints"},
		{`print(slice("x", "y", 0))`, "type error at 1:7", "slice() requires start and end to be ints"},

		// sort() builtin
		{`lst = [3,1,2]  sort(lst)  print(lst)  sort(lst)  print(lst)`, "", "[1, 2, 3]\n[1, 2, 3]"},
		{`lst = ["y","x","Z"]  sort(lst)  print(lst)`, "", `["Z", "x", "y"]`},
		{`lst = []  sort(lst)  print(lst)`, "", "[]"},
		{`lst = [42]  sort(lst)  print(lst)`, "", "[42]"},
		{`sort([1, "x"])`, "type error at 1:1", "comparison requires two ints or two strs (or lists of ints or strs)"},
		{`func f(x) { print("KEY:", x)  return -x }  lst=[1,3,2]  sort(lst, f)  print(lst)`, "",
			"KEY: 1\nKEY: 3\nKEY: 2\n[3, 2, 1]"},
		{`lst = [["B", 42], ["a", 43], ["a", 42], ["z", 0]]  sort(lst)  print(lst)`, "",
			`[["B", 42], ["a", 42], ["a", 43], ["z", 0]]`},
		{`lst = [["B", 42], ["a", 43], ["a", 42], ["z", 0]]  sort(lst, func(x) { return x[1] })  print(lst)  sort(lst, func(x) { return lower(x[0]) })  print(lst)`, "",
			`[["z", 0], ["B", 42], ["a", 42], ["a", 43]]
[["a", 42], ["a", 43], ["B", 42], ["z", 0]]`},
		{`lst = [["B", 42], ["a", 43], ["a", 42], ["z", 0]]  sort(lst, func(x) { return [lower(x[0]), x[1]] })  print(lst)`, "",
			`[["a", 42], ["a", 43], ["B", 42], ["z", 0]]`},

		// split() builtin
		{`print(split("\tx\ry\nz ", nil), split("xyz", nil), split("", nil))`, "", `["x", "y", "z"] ["xyz"] []`},
		{`print(split("\tx\ry\nz "), split("xyz"), split(""))`, "", `["x", "y", "z"] ["xyz"] []`},
		{`print(split("x|y|z", "|"), split("xyz", "|"), split("", "|"))`, "", `["x", "y", "z"] ["xyz"] [""]`},
		{`split()`, "type error at 1:1", "split() requires 1 or 2 args, got 0"},
		{`split("x", 42)`, "type error at 1:1", "split() requires separator to be a str or nil"},

		// str() builtin
		{`print(str("foo"))  print(str("x"), str(42))  print(str([1, 2, 3]))`, "", "foo\nx 42\n[1, 2, 3]"},
		{`print(str(nil), str(true), str(false), str(1), str("x"), str(["y"]), str({"z": 2}), str(func() {}))`, "",
			`nil true false 1 x ["y"] {"z": 2} <func>`},
		{`str()`, "type error at 1:1", "str() requires 1 arg, got 0"},

		// type() builtin
		{`print(type(nil), type(true), type(false), type(0), type("x"), type([]), type({}), type(func() {}))`, "",
			"nil bool bool int str list map func"},
		{`type()`, "type error at 1:1", "type() requires 1 arg, got 0"},

		// upper() builtin
		{`print(upper(""), upper("abc"), upper("FoO"), upper("BAR"))`, "", " ABC FOO BAR"},
		{`print(upper(42))`, "type error at 1:7", "upper() requires a str"},
		{`print(upper())`, "type error at 1:7", "upper() requires 1 arg, got 0"},
	}

	// Run tests against Go interpreter
	for _, test := range tests {
		testName := "go_" + test.source
		if len(testName) > 70 {
			testName = testName[:70]
		}
		t.Run(testName, func(t *testing.T) {
			prog, err := parser.ParseProgram([]byte(test.source))
			if err != nil {
				t.Fatalf("%s", err)
			}
			stdin := bytes.NewBuffer([]byte("dummy stdin"))
			stdout := &bytes.Buffer{}
			config := &interpreter.Config{
				Args:   []string{"one", "2", "THREE"},
				Stdin:  stdin,
				Stdout: stdout,
				Exit:   func(n int) { fmt.Fprintf(stdout, "exit(%d)", n) },
			}
			_, err = interpreter.Execute(prog, config)
			var output string
			if err != nil {
				errOutput := fmt.Sprintf("%s", err)
				fields := strings.SplitN(errOutput, ": ", 2)
				if len(fields) < 2 {
					t.Fatalf("expected \": \" in error output, got %q", errOutput)
				}
				errpos := fields[0]
				if errpos != test.errpos {
					t.Fatalf("expected errpos %q, got %q", test.errpos, errpos)
				}
				output = fields[1]
			} else {
				output = strings.TrimRight(stdout.String(), "\n")
			}
			if output != test.output {
				t.Fatalf("expected:\n\"%s\"\ngot:\n\"%s\"", test.output, output)
			}
		})
	}

	// Run tests against external littlelang interpreter
	if exePath != "" {
		for _, test := range tests {
			testName := "ll_" + test.source
			if len(testName) > 70 {
				testName = testName[:70]
			}
			t.Run(testName, func(t *testing.T) {
				srcFile, err := ioutil.TempFile("", "lltest_")
				if err != nil {
					t.Fatalf("error creating temp file: %v", err)
				}
				defer os.Remove(srcFile.Name())
				_, err = srcFile.Write([]byte(test.source))
				if err != nil {
					t.Fatalf("error writing temp file: %v", err)
				}

				cmd := exec.Command(exePath, interpPath, srcFile.Name(), "one", "2", "THREE")
				stdin, err := cmd.StdinPipe()
				if err != nil {
					t.Fatalf("error creating stdin pipe: %v", err)
				}
				_, err = stdin.Write([]byte("dummy stdin"))
				if err != nil {
					t.Fatalf("error writing temp file: %v", err)
				}
				stdin.Close()

				outBytes, err := cmd.Output()
				output := string(outBytes)
				if err != nil {
					if test.errpos == "" {
						t.Fatalf("expected no error, got error %v", err)
					}
					lines := strings.Split(output, "\n")
					if len(lines) < 2 {
						t.Fatalf("expected at least two lines, got %d", len(lines))
					}
					lastLine := lines[len(lines)-2]
					fields := strings.SplitN(lastLine, ": ", 2)
					if len(fields) < 2 {
						t.Fatalf("expected \": \" in error output, got %q", lastLine)
					}
					output = fields[1]
				} else {
					output = strings.TrimRight(output, "\n")
					if test.errpos != "" {
						t.Fatalf("expected error %q, got no error (output %q)", test.errpos, output)
					}
				}
				if output != test.output {
					t.Fatalf("expected:\n\"%s\"\ngot:\n\"%s\"", test.output, output)
				}
			})
		}
	}
}
