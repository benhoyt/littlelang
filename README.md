# A little language interpreter

The littlelang programming language is a little language (funny that) designed by Ben Hoyt for fun and (his own) learning. It's kind of a cross between Python, JavaScript, and Go. It's a dynamically but strongly-typed language with the usual data types, first-class functions, closures, and a bit more.

The code includes a tokenizer and parser and a (slowish but simple) tree-walk interpreter written in Go. There's also an [interpreter written in littlelang itself](https://github.com/benhoyt/littlelang/blob/master/littlelang.ll), just to prove the language is powerful enough to write somewhat real programs in.

Below are a couple of [examples](#some-little-examples) of the language, the full language ["spec"](#language-spec), and the littlelang [grammar](#grammar). However, you might be better off if you [**read my introduction first**](http://benhoyt.com/writings/littlelang/).


## Some little examples

```
// Lists, the sort() builtin, and for loops
lst = ["foo", "a", "z", "B"]
sort(lst)
print(lst)
sort(lst, lower)
for x in lst {
    print(x)
}
// Output:
// ["B", "a", "foo", "z"]
// a
// B
// foo
// z

// A closure and first-class functions
func make_adder(n) {
    func adder(x) {
        return x + n
    }
    return adder
}
add5 = make_adder(5)
print("add5(3) =", add5(3))
// Output:
// add5(3) = 8

// A pseudo-class with "methods" using a closure
func Person(name, age) {
    self = {}
    self.name = name
    self.age = age
    self.str = func() {
        return self.name + ", aged " + str(self.age)
    }
    return self
}
p = Person("Bob", 42)
print(p.str())
// Output:
// Bob, aged 42
```


## Language spec

Littlelang's syntax is a cross between Go and Python. Like Go, it uses `func` to define functions (named or anonymous), requires `{` and `}` for blocks, and doesn't need semicolons. But like Python, it uses keywords for `and` and `or` and `in`. Like both those languages, it distinguishes expressions and statements.

It's dynamically typed and garbage collected, with the usual data types: nil, bool, int, str, list, map, and func. There are also several builtin functions.

Calling this a "spec" is probably a bit grandiose, but it's the best you'll get.

### Programs

A littlelang program is simply zero or more statements. Statements don't actually have to be separated by newlines, only by whitespace. The following is a valid program (but you'd probably use newlines in the `if` block in real life):

```
s = "world"
print("Hello, " + s)
if s != "" { t = "The end"  print(t) }
// Hello, world
// The end
```

Between tokens, whitespace and comments (`//` through to the end of a line) are ignored.

### Types

Littlelang has the following data types: nil, bool, int, str, list, map, and func. The int type is a signed 64-bit integer, strings are immutable arrays of bytes, lists are growable arrays (use the `append()` builtin), and maps are unordered hash tables. Trailing commas are allowed after the last element in a list or map:

Type      | Syntax                                    | Comments
--------- | ----------------------------------------- | --------
nil       | `nil`                                     |
bool      | `true false`                              |
int       | `0 42 1234 -5`                            | `-5` is actually `5` with unary `-`
str       | `"" "foo" "\"quotes\" and a\nline break"` | Escapes: `\" \\ \t \r \n`
list      | `[] [1, 2,] [1, 2, 3]`                    |
map       | `{} {"a": 1,} {"a": 1, "b": 2}`           |

### If statements

Littlelang supports `if`, `else if`, and `else`. You must use `{ ... }` braces around the blocks:

```
a = 10
if a > 5 {
    print("large")
} else if a < 0 {
    print("negative")
} else {
    print("small")
}
// large
```

### While loops

While loops are very standard:

```
i = 3
while i > 0 {
    print(i)
    i = i - 1
}
// 3
// 2
// 1
```

Littlelang does not have `break` or `continue`, but you can `return value` as one way of breaking out of a loop early.

### For loops

For loops are similar to Python's `for` loops and Go's `for range` loops. You can iterate through the (Unicode) characters in a string, elements in a list (the `range()` builtin returns a list), and keys in a map.

Note that iteration order of a map is undefined -- create a list of keys and `sort()` if you need that.

```
for c in "foo" {
    print(c)
}
// f
// o
// o

for x in [nil, 3, "z"] {
    print(x)
}
// nil
// 3
// z

for i in range(5) {
    print(i, i*i)
}
// 0 0
// 1 1
// 2 4
// 3 9
// 4 16

map = {"a": 1, "b": 2}
for k in map {
    print(k, map[k])
}
// a 1
// b 2
```

### Functions and return

You can define named or anonymous functions, including functions inside functions that reference outer variables (closures). Vararg functions are supported with `...` syntax like in Go.

```
func add(a, b) {
    return a + b
}
print(add(3, 4))
// 7

func make_adder(n) {
    func adder(x) {
        return x + n
    }
    return adder
}
add5 = make_adder(5)
print(add5(7))
// 12

// Anonymous function, equivalent to "func plus(nums...)"
plus = func(nums...) {
    sum = 0
    for n in nums {
        sum = sum + n
    }
    return sum
}
print(plus(1, 2, 3))
lst = [4, 5, 6]
print(plus(lst...))
// 6
// 15
```

A grammar note: you can't have a "bare return" -- it requires a return value. So if you don't want to return anything (functions always return at least nil anyway), just say `return nil`.

### Assignment

Assignment can assign to a name, a list element by index, or a map value by key. When assigning to a name (variable), it always assigns to the local function scope (like Python). You can't assign to an outer scope without using a mutable list or map (there's no `global` or `nonlocal` keyword).

To help with object-oriented programming, `obj.foo = bar` is syntactic sugar for `obj["foo"] = bar`. They're exactly equivalent.

```
i = 1
func nochange() {
    i = 2
    print(i)
}
print(i)
nochange()
print(i)
// 1
// 2
// 1

map = {"a": 1}
func change() {
    map.a = 2
    print(map.a)
}
print(map.a)
change()
print(map.a)
// 1
// 2
// 2

lst = [0, 1, 2]
lst[1] = "one"
print(lst)
// [0, "one", 2]

map = {"a": 1, "b": 2}
map["a"] = 3
map.c = 4
print(map)
// {"a": 3, "b": 2, "c": 4}
```

### Binary and unary operators

Littlelang supports pretty standard binary and unary operators. Here they are with their precedence, from highest to lowest (operators of the same precedence evaluate left to right):

Operators      | Description
-------------- | -----------
`[]`           | Subscript
`-`            | Unary minus
`* / %`        | Multiplication
`+ -`          | Addition
`< <= > >= in` | Comparison
`== !=`        | Equality
`not`          | Logical not
`and`          | Logical and (short-circuit)
`or`           | Logical or (short-circuit)

Several of the operators are overloaded. Here are the types they can operate on:

Operator   | Types           | Action
---------- | --------------- | ------
`[]`       | `str[int]`      | fetch nth byte of str (0-based)
`[]`       | `list[int]`     | fetch nth element of list (0-based)
`[]`       | `map[str]`      | fetch map value by key str
`-`        | `int`           | negate int
`*`        | `int * int`     | multiply ints
`*`        | `str * int`     | repeat str n times
`*`        | `int * str`     | repeat str n times
`*`        | `list * int`    | repeat list n times, give new list
`*`        | `int * list`    | repeat list n times, give new list
`/`        | `int / int`     | divide ints, truncated
`%`        | `int % int`     | divide ints, give remainder
`+`        | `int + int`     | add ints
`+`        | `str + str`     | concatenate strs, give new string
`+`        | `list + list`   | concatenate lists, give new list
`+`        | `map + map`     | merge maps into new map, keys in right map win
`-`        | `int - int`     | subtract ints
`<`        | `int < int`     | true iff left < right
`<`        | `str < str`     | true iff left < right (lexicographical)
`<`        | `list < list`   | true iff left < right (lexicographical, recursive)
`<= > >=`  | same as `<`     | similar to `<`
`in`       | `str in str`    | true iff left is substr of right
`in`       | `any in list`   | true iff one of list elements == left
`in`       | `str in map`    | true iff key in map
`==`       | `any == any`    | deep equality (always false if different type)
`!=`       | `any != any`    | same as `not ==`
`not`      | `not bool`      | inverse of bool
`and`      | `bool and bool` | true iff both true, right not evaluated if left false
`or`       | `bool or bool`  | true iff either true, right not evaluated if left true

### Builtin functions

`append(list, values...)` appends the given elements to list, modifying the list in place. It returns nil, rather than returning the list, to reinforce the fact that it has side effects.

`args()` returns a list of the command-line arguments passed to the interpreter (after the littlelang source filename).

`char(int)` returns a one-character string with the given Unicode codepoint.

`exit([int])` exits the program immediately with given status code (0 if not given).

`find(haystack, needle)` returns the index of needle str in haystack str, or the index of needle element in haystack list. Returns -1 if not found.

`int(str_or_int)` converts decimal str to int (returns nil if invalid). If argument is an int already, return it directly.

`join(list, sep)` concatenates strs in list to form a single str, with the separator str between each element.

`len(iterable)` returns the length of a str (number of bytes), list (number of elements), or map (number of key/value pairs).

`lower(str)` returns a lowercased version of str.

`print(values...)` prints all values separated by a space and followed by a newline. The equivalent of `str(v)` is called on every value to convert it to a str.

`range(int)` returns a list of the numbers from 0 through int-1.

`read([filename])` reads standard input or the given file and returns the contents as a str.

`rune(str)` returns the Unicode codepoint for the given 1-character str.

`slice(str_or_list, start, end)` returns a subslice of the given str or list from index start through end-1. When slicing a list, the input list is not changed.

`sort(list[, func])` sorts the list in place using a stable sort, and returns nil. Elements in the list must be orderable with `<` (int, str, or list of those). If a key function is provided, it must take the element as an argument and return an orderable value to use as the sort key.

`split(str[, sep])` splits the str using given separator, and returns the parts (excluding the separator) as a list. If sep is not given or nil, it splits on whitespace.

`str(value)` returns the string representation of value: `nil` for nil, `true` or `false` for bool, decimal for int (eg: `1234`), the str itself for str (not quoted), the littlelang representation for list and map (eg: `[1, 2]` and `{"a": 1}` with keys sorted), and something like `<func name>` for func.

`type(value)` returns a str denoting the type of value: `nil`, `bool`, `int`, `str`, `list`, `map`, or `func`.

`upper(str)` returns an uppercased version of str.


## Grammar

Below is the full littlelang grammar in pseudo-BNF format. Rules are in lowercase letters like "statement", and single tokens are in allcaps like "COMMA" and "NAME" (see tokenizer/tokenizer.go for the full list of tokens).

```
program    = statement*
statement  = if | while | for | return | func | assign | expression
if         = IF expression block |
             IF expression block ELSE block |
             IF expression block ELSE if
block      = LBRACE statement* RBRACE
while      = WHILE expression block
for        = FOR NAME IN expression block
return     = RETURN expression
func       = FUNC NAME params block |
             FUNC params block
params     = LPAREN RPAREN |
             LPAREN NAME (COMMA NAME)* ELLIPSIS? COMMA? RPAREN |
assign     = NAME ASSIGN expression |
             call subscript ASSIGN expression |
             call dot ASSIGN expression

expression = and (OR and)*
and        = not (AND not)*
not        = NOT not | equality
equality   = comparison ((EQUAL | NOTEQUAL) comparison)*
comparison = addition ((LT | LTE | GT | GTE | IN) addition)*
addition   = multiply ((PLUS | MINUS) multiply)*
multiply   = negative ((TIMES | DIVIDE | MODULO) negative)*
negative   = MINUS negative | call
call       = primary (args | subscript | dot)*
args       = LPAREN RPAREN |
             LPAREN expression (COMMA expression)* ELLIPSIS? COMMA? RPAREN)
subscript  = LBRACKET expression RBRACKET
dot        = DOT NAME
primary    = NAME | INT | STR | TRUE | FALSE | NIL | list | map |
             FUNC params block |
             LPAREN expression RPAREN
list       = LBRACKET RBRACKET |
             LBRACKET expression (COMMA expression)* COMMA? RBRACKET
map        = LBRACE RBRACE |
             LBRACE expression COLON expression
                    (COMMA expression COLON expression)* COMMA? RBRACE
```


## Building and running

To build, [install Go](https://golang.org/), then fetch and build like so:

```
cd ~/go/src  # or wherever your Go code lives
go get github.com/benhoyt/littlelang
cd github.com/benhoyt/littlelang/
go build
```

You can then run one of the examples using the Go interpreter binary:

```
./littlelang examples/readme.ll
```

If you want to get really meta, run the README example using the littlelang interpreter running under the Go interpreter:

```
./littlelang littlelang.ll examples/readme.ll
./littlelang littlelang.ll littlelang.ll examples/readme.ll
```

How deep does the rabbit hole go?


## Credits

Many thanks to Bob Nystrom for his free book [Crafting Interpreters](http://www.craftinginterpreters.com/), which is a great read and helped me understand how to implement closures.
