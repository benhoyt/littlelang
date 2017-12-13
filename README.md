# A little language interpreter

The littlelang programming language is a little language (surprise!) designed
by Ben Hoyt for fun and (his own) learning. It's kind of a cross between
JavaScript, Python, and Go. It's a dynamically- but strongly-typed language
with the usual data types, first-class functions, closures, and a bit more.

It's a tokenizer and parser and a (slow but simple) tree-walk interpreter
written in Go. There's also an [interpreter written in littlelang
itself](https://github.com/benhoyt/littlelang/blob/master/littlelang.ll), just
to prove the language is powerful enough to write somewhat real programs in.

<!-- Below is an example of the language as well as the language grammar, but you
can [read more here](http://benhoyt.com/writings/littlelang/). -->


## An example

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


## Grammar

Below is the full littlelang grammar in pseudo-BNF format. Rules are in
lowercase letters like "statement", and single tokens are in allcaps like
"COMMA" and "NAME" (see tokenizer/tokenizer.go for the full list of tokens).

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

Many thanks to Bob Nystrom for his free book
[Crafting Interpreters](http://www.craftinginterpreters.com/), which is a
great read and helped me understand how to implement closures.


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
