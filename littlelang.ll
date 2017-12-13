// A littlelang interpreter, written in littlelang

// ------------------------------------------------------------------
// Tokenizer
// ------------------------------------------------------------------

// Stop tokens
ILLEGAL = "ILLEGAL"
EOF = "EOF"

// Keywords
AND = "and"
ELSE = "else"
FALSE = "false"
FOR = "for"
FUNC = "func"
IF = "if"
IN = "in"
NIL = "nil"
NOT = "not"
OR = "or"
RETURN = "return"
TRUE = "true"
WHILE = "while"

// Single-character tokens
ASSIGN = "="
COLON = ":"
COMMA = ","
DIVIDE = "/"
DOT = "."
GT = ">"
LBRACE = "{"
LBRACKET = "["
LPAREN = "("
LT = "<"
MINUS = "-"
MODULO = "%"
PLUS = "+"
RBRACE = "}"
RBRACKET = "]"
RPAREN = ")"
TIMES = "*"

single_only_tokens = {}
for tok in [
    COLON, COMMA, DIVIDE, LBRACE, LBRACKET, LPAREN, MINUS, MODULO, PLUS,
    RBRACE, RBRACKET, RPAREN, TIMES,
] {
    single_only_tokens[tok] = true
}

// Two-character tokens
EQUAL = "=="
GTE = ">="
LTE = "<="
NOTEQUAL = "!="

// Three-character tokens
ELLIPSIS = "..."

// Literals and identifiers
INT = "int"
NAME = "name"
STR = "str"

keyword_tokens = {
    "and": true,
    "else": true,
    "false": true,
    "for": true,
    "func": true,
    "if": true,
    "in": true,
    "nil": true,
    "not": true,
    "or": true,
    "return": true,
    "true": true,
    "while": true,
}

func Pos(line, col) {
    self = {}
    self.line = line
    self.col = col
    return self
}

func Token(tok, val, pos) {
    self = {}
    self.tok = tok
    self.val = val
    self.pos = pos
    return self
}

func is_name_start(ch) {
    return ch == "_" or (ch >= "a" and ch <= "z") or (ch >= "A" and ch <= "Z")
}

// Tokenize source string and return list of Token objects
func tokenize(source) {
    t = {}
    t.offset = 0
    t.line = 1
    t.col = 1
    t.nextLine = 1
    t.nextCol = 1
    t.ch = nil

    func next() {
        t.line = t.nextLine
        t.col = t.nextCol
        if t.offset >= len(source) {
            t.ch = nil
            return nil
        }
        t.ch = source[t.offset]
        t.offset = t.offset + 1
        if t.ch == "\n" {
            t.nextLine = t.nextLine + 1
            t.nextCol = 1
        } else {
            t.nextCol = t.nextCol + 1
        }
    }

    func skip_whitespace_and_comments() {
        while true {
            while t.ch == " " or t.ch == "\t" or t.ch == "\r" or t.ch == "\n" {
                next()
            }
            if not (t.ch == "/" and t.offset < len(source) and source[t.offset] == "/") {
                return nil
            }
            // Skip //-prefixed comment (to end of line or end of input)
            next()
            next()
            while t.ch != "\n" and t.ch != nil {
                next()
            }
            next()
        }
    }

    // Kick things off
    tokens = []
    next()

    func end(tok, val, line, col) {
        append(tokens, Token(tok, val, Pos(line, col)))
        return tokens
    }

    while true {
        skip_whitespace_and_comments()
        if t.ch == nil {
            return end(EOF, "", t.line, t.col)
        }

        val = ""
        line = t.line
        col = t.col
        ch = t.ch
        next()

        if is_name_start(ch) {
            chars = [ch]
            while t.ch != nil and (is_name_start(t.ch) or (t.ch >= "0" and t.ch <= "9")) {
                append(chars, t.ch)
                next()
            }
            tok = join(chars, "")
            if not tok in keyword_tokens {
                val = tok
                tok = NAME
            }
        } else if ch in single_only_tokens {
            tok = ch
        } else if ch == "=" {
            if t.ch == "=" {
                next()
                tok = EQUAL
            } else {
                tok = ASSIGN
            }
        } else if ch == "!" {
            if t.ch == "=" {
                next()
                tok = NOTEQUAL
            } else {
                return end(ILLEGAL, "expected != instead of !" + t.ch, line, col)
            }
        } else if ch == "<" {
            if t.ch == "=" {
                next()
                tok = LTE
            } else {
                tok = LT
            }
        } else if ch == ">" {
            if t.ch == "=" {
                next()
                tok = GTE
            } else {
                tok = GT
            }
        } else if ch == "." {
            if t.ch == "." {
                next()
                if t.ch != "." {
                    return end(ILLEGAL, "unexpected ..", line, col)
                }
                next()
                tok = ELLIPSIS
            } else {
                tok = DOT
            }
        } else if ch >= "0" and ch <= "9" {
            chars = [ch]
            while t.ch != nil and t.ch >= "0" and t.ch <= "9" {
                append(chars, t.ch)
                next()
            }
            tok = INT
            val = int(join(chars, ""))
        } else if ch == "\"" {
            chars = []
            while t.ch != "\"" {
                c = t.ch
                if c == nil {
                    return end(ILLEGAL, "didn't find end quote in string", line, col)
                }
                if c == "\r" or c == "\n" {
                    return end(ILLEGAL, "can't have newline in string", line, col)
                }
                if c == "\\" {
                    next()
                    if t.ch == "\"" or t.ch == "\\" {
                        c = t.ch
                    } else if t.ch == "t" {
                        c = "\t"
                    } else if t.ch == "r" {
                        c = "\r"
                    } else if t.ch == "n" {
                        c = "\n"
                    } else {
                        return end(ILLEGAL, "invalid string escape \\" + t.ch, line, col)
                    }
                }
                append(chars, c)
                next()
            }
            next()
            tok = STR
            val = join(chars, "")
        } else {
            return end(ILLEGAL, "unexpected " + ch, line, col)
        }

        append(tokens, Token(tok, val, Pos(line, col)))
    }
}

// ------------------------------------------------------------------
// AST node help functions (node.str() converts node back to source string)
// ------------------------------------------------------------------

func indent(s) {
    lines = split(s, "\n")
    output = []
    for line in lines {
        append(output, "    " + line)
    }
    return join(output, "\n")
}

func Node(type, pos) {
    self = {}
    self.type = type
    self.pos = pos
    return self
}

func Program(pos, body) {
    self = Node("Program", pos)
    self.body = body
    self.str = func() {
        return self.body.str()
    }
    return self
}

func Block(pos, statements) {
    self = Node("Block", pos)
    self.statements = statements
    self.str = func() {
        lines = []
        for statement in self.statements {
            append(lines, statement.str())
        }
        return join(lines, "\n")
    }
    return self
}

func Assign(pos, target, value) {
    self = Node("Assign", pos)
    self.target = target
    self.value = value
    self.str = func() {
        return self.target.str() + " = " + self.value.str()
    }
    return self
}

func If(pos, condition, body, else_body) {
    self = Node("If", pos)
    self.condition = condition
    self.body = body
    self.else_body = else_body
    self.str = func() {
        s = "if " + self.condition.str() + " {\n" + indent(self.body.str()) + "\n}"
        if self.else_body != nil {
            s = s + " else {\n" + indent(self.else_body.str()) + "\n}"
        }
        return s
    }
    return self
}

func While(pos, condition, body) {
    self = Node("While", pos)
    self.condition = condition
    self.body = body
    self.str = func() {
        return "while " + self.condition.str() + " {\n" + indent(self.body.str()) + "\n}"
    }
    return self
}

func For(pos, name, iterable, body) {
    self = Node("For", pos)
    self.name = name
    self.iterable = iterable
    self.body = body
    self.str = func() {
        return "for " + self.name + " in " + self.iterable.str() + " {\n" + indent(self.body.str()) + "\n}"
    }
    return self
}

func Return(pos, result) {
    self = Node("Return", pos)
    self.result = result
    self.str = func() {
        return "return " + self.result.str()
    }
    return self
}

func ExpressionStatement(pos, expr) {
    self = Node("ExpressionStatement", pos)
    self.expr = expr
    self.str = func() {
        return self.expr.str()
    }
    return self
}

func FunctionDefinition(pos, name, params, ellipsis, body) {
    self = Node("FunctionDefinition", pos)
    self.name = name
    self.params = params
    self.ellipsis = ellipsis
    self.body = body
    self.str = func() {
        ellipsis_str = ""
        if self.ellipsis {
            ellipsis_str = "..."
        }
        body_str = ""
        if len(self.body) != 0 {
            body_str = "\n" + indent(self.body.str()) + "\n"
        }
        return "func " + self.name + "(" + join(self.params, ", ") + ellipsis_str + ") {" + body_str + "}"
    }
    return self
}

func Binary(pos, left, op, right) {
    self = Node("Binary", pos)
    self.left = left
    self.op = op
    self.right = right
    self.str = func() {
        return "(" + self.left.str() + " " + self.op + " " + self.right.str() + ")"
    }
    return self
}

func Unary(pos, operator, operand) {
    self = Node("Unary", pos)
    self.operator = operator
    self.operand = operand
    self.str = func() {
        space = ""
        if self.operator == NOT {
            space = " "
        }
        return "(" + self.operator + space + self.operand.str() + ")"
    }
    return self
}

func Call(pos, function, args, ellipsis) {
    self = Node("Call", pos)
    self.function = function
    self.args = args
    self.ellipsis = ellipsis
    self.str = func() {
        args = []
        for arg in self.args {
            append(args, arg.str())
        }
        ellipsis_str = ""
        if self.ellipsis {
            ellipsis_str = "..."
        }
        return self.function.str() + "(" + join(args, ", ") + ellipsis_str + ")"
    }
    return self
}

func Literal(pos, value) {
    self = Node("Literal", pos)
    self.value = value
    self.str = func() {
        if type(self.value) == "str" {
            quoted = str([self.value])
            quoted = slice(quoted, 1, len(quoted)-1)
            return quoted
        }
        return str(self.value)
    }
    return self
}

func List(pos, values) {
    self = Node("List", pos)
    self.values = values
    self.str = func() {
        values = []
        for value in self.values {
            append(values, value.str())
        }
        return "[" + join(values, ", ") + "]"
    }
    return self
}

func Map(pos, items) {
    self = Node("Map", pos)
    self.items = items
    self.str = func() {
        items = []
        for item in self.items {
            append(items, item[0].str() + ": " + item[1].str())
        }
        return "{" + join(items, ", ") + "}"
    }
    return self
}

func FunctionExpression(pos, params, ellipsis, body) {
    self = Node("FunctionExpression", pos)
    self.params = params
    self.ellipsis = ellipsis
    self.body = body
    self.str = func() {
        ellipsis_str = ""
        if self.ellipsis {
            ellipsis_str = "..."
        }
        body_str = ""
        if len(self.body) != 0 {
            body_str = "\n" + indent(self.body.str()) + "\n"
        }
        return "func(" + join(self.params, ", ") + ellipsis_str + ") {" + body_str + "}"
    }
    return self
}

func Subscript(pos, container, subscript) {
    self = Node("Subscript", pos)
    self.container = container
    self.subscript = subscript
    self.str = func() {
        return self.container.str() + "[" + self.subscript.str() + "]"
    }
    return self
}

func Variable(pos, name) {
    self = Node("Variable", pos)
    self.name = name
    self.str = func() {
        return self.name
    }
    return self
}

// ------------------------------------------------------------------
// Parser
// ------------------------------------------------------------------

// Parse source string and return a Program node
func parse(source) {
    p = {}
    p.tokens = tokenize(source)
    p.offset = 0
    p.tok = nil
    p.val = nil
    p.pos = nil

    func error(msg) {
        print("parse error at " + str(p.pos.line) + ":" + str(p.pos.col) + ": " + msg)
        exit(1)
    }

    func next() {
        token = p.tokens[p.offset]
        p.offset = p.offset + 1
        p.tok = token.tok
        p.val = token.val
        p.pos = token.pos
        if p.tok == ILLEGAL {
            error(p.val)
        }
    }

    func expect(tok) {
        if p.tok != tok {
            error("expected " + tok + " and not " + p.tok)
        }
        next()
    }

    func matches(operators...) {
        for operator in operators {
            if p.tok == operator {
                return true
            }
        }
        return false
    }

    func program() {
        pos = p.pos
        return Program(pos, statements(EOF))
    }

    func statements(end) {
        pos = p.pos
        statement_list = []
        while p.tok != end and p.tok != EOF {
            append(statement_list, statement())
        }
        return Block(pos, statement_list)
    }

    func statement() {
        if p.tok == IF {
            return if_()
        } else if p.tok == WHILE {
            return while_()
        } else if p.tok == FOR {
            return for_()
        } else if p.tok == RETURN {
            return return_()
        } else if p.tok == FUNC {
            return func_()
        }
        pos = p.pos
        expr = expression()
        if p.tok == ASSIGN {
            pos = p.pos
            if expr.type == "Variable" or expr.type == "Subscript" {
                next()
                value = expression()
                return Assign(pos, expr, value)
            } else {
                error("expected name, subscript, or dot expression on left side of =")
            }
        }
        return ExpressionStatement(pos, expr)
    }

    func block() {
        expect(LBRACE)
        body = statements(RBRACE)
        expect(RBRACE)
        return body
    }

    func if_() {
        pos = p.pos
        expect(IF)
        condition = expression()
        body = block()
        else_body = nil
        if p.tok == ELSE {
            next()
            if p.tok == LBRACE {
                else_body = block()
            } else if p.tok == IF {
                else_body = Block(p.pos, [if_()])
            } else {
                error("expected { or if after else, not " + p.tok)
            }
        }
        return If(pos, condition, body, else_body)
    }

    func while_() {
        pos = p.pos
        expect(WHILE)
        condition = expression()
        body = block()
        return While(pos, condition, body)
    }

    func for_() {
        pos = p.pos
        expect(FOR)
        name = p.val
        expect(NAME)
        expect(IN)
        iterable = expression()
        body = block()
        return For(pos, name, iterable, body)
    }

    func return_() {
        pos = p.pos
        expect(RETURN)
        result = expression()
        return Return(pos, result)
    }

    func func_() {
        pos = p.pos
        expect(FUNC)
        if p.tok == NAME {
            name = p.val
            next()
            params_ellipsis = params()
            body = block()
            return FunctionDefinition(pos, name, params_ellipsis[0], params_ellipsis[1], body)
        } else {
            params_ellipsis = params()
            body = block()
            expr = FunctionExpression(pos, params_ellipsis[0], params_ellipsis[1], body)
            return ExpressionStatement(pos, expr)
        }
    }

    func params() {
        expect(LPAREN)
        params = []
        got_comma = true
        got_ellipsis = false
        while p.tok != RPAREN and p.tok != EOF and not got_ellipsis {
            if not got_comma {
                error("expected , between parameters")
            }
            param = p.val
            expect(NAME)
            append(params, param)
            if p.tok == ELLIPSIS {
                got_ellipsis = true
                next()
            }
            if p.tok == COMMA {
                got_comma = true
                next()
            } else {
                got_comma = false
            }
        }
        if p.tok != RPAREN and got_ellipsis {
            error("can only have ... after last parameter")
        }
        expect(RPAREN)
        return [params, got_ellipsis]
    }

    func binary(parse_expr, operators...) {
        expr = parse_expr()
        while matches(operators...) {
            op = p.tok
            pos = p.pos
            next()
            right = parse_expr()
            expr = Binary(pos, expr, op, right)
        }
        return expr
    }

    func expression() {
        return binary(and_, OR)
    }

    func and_() {
        return binary(not_, AND)
    }

    func not_() {
        if p.tok == NOT {
            pos = p.pos
            next()
            operand = not_()
            return Unary(pos, NOT, operand)
        }
        return equality()
    }

    func equality() {
        return binary(comparison, EQUAL, NOTEQUAL)
    }

    func comparison() {
        return binary(addition, LT, LTE, GT, GTE, IN)
    }

    func addition() {
        return binary(multiply, PLUS, MINUS)
    }

    func multiply() {
        return binary(negative, TIMES, DIVIDE, MODULO)
    }

    func negative() {
        if p.tok == MINUS {
            pos = p.pos
            next()
            operand = negative()
            return Unary(pos, MINUS, operand)
        }
        return call()
    }

    func call() {
        expr = primary()
        while matches(LPAREN, LBRACKET, DOT) {
            if p.tok == LPAREN {
                pos = p.pos
                next()
                args = []
                got_comma = true
                got_ellipsis = false
                while p.tok != RPAREN and p.tok != EOF and not got_ellipsis {
                    if not got_comma {
                        error("expected , between arguments")
                    }
                    arg = expression()
                    append(args, arg)
                    if p.tok == ELLIPSIS {
                        got_ellipsis = true
                        next()
                    }
                    if p.tok == COMMA {
                        got_comma = true
                        next()
                    } else {
                        got_comma = false
                    }
                }
                if p.tok != RPAREN and got_ellipsis {
                    error("can only have ... after last argument")
                }
                expect(RPAREN)
                expr = Call(pos, expr, args, got_ellipsis)
            } else if p.tok == LBRACKET {
                pos = p.pos
                next()
                subscript = expression()
                expect(RBRACKET)
                expr = Subscript(pos, expr, subscript)
            } else {
                pos = p.pos
                next()
                subscript = Literal(p.pos, p.val)
                expect(NAME)
                expr = Subscript(pos, expr, subscript)
            }
        }
        return expr
    }

    func primary() {
        if p.tok == NAME {
            name = p.val
            pos = p.pos
            next()
            return Variable(pos, name)
        } else if p.tok == INT or p.tok == STR {
            val = p.val
            pos = p.pos
            next()
            return Literal(pos, val)
        } else if p.tok == TRUE {
            pos = p.pos
            next()
            return Literal(pos, true)
        } else if p.tok == FALSE {
            pos = p.pos
            next()
            return Literal(pos, false)
        } else if p.tok == NIL {
            pos = p.pos
            next()
            return Literal(pos, nil)
        } else if p.tok == LBRACKET {
            return list()
        } else if p.tok == LBRACE {
            return map()
        } else if p.tok == FUNC {
            pos = p.pos
            next()
            params_ellipsis = params()
            body = block()
            return FunctionExpression(pos, params_ellipsis[0], params_ellipsis[1], body)
        } else if p.tok == LPAREN {
            next()
            expr = expression()
            expect(RPAREN)
            return expr
        } else {
            error("expected expression, not " + p.tok)
        }
    }

    func list() {
        pos = p.pos
        expect(LBRACKET)
        values = []
        got_comma = true
        while p.tok != RBRACKET and p.tok != EOF {
            if not got_comma {
                error("expected , between list elements")
            }
            value = expression()
            append(values, value)
            if p.tok == COMMA {
                got_comma = true
                next()
            } else {
                got_comma = false
            }
        }
        expect(RBRACKET)
        return List(pos, values)
    }

    func map() {
        pos = p.pos
        expect(LBRACE)
        items = []
        got_comma = true
        while p.tok != RBRACE and p.tok != EOF {
            if not got_comma {
                error("expected , between map items")
            }
            key = expression()
            expect(COLON)
            value = expression()
            append(items, [key, value])
            if p.tok == COMMA {
                got_comma = true
                next()
            } else {
                got_comma = false
            }
        }
        expect(RBRACE)
        return Map(pos, items)
    }

    next()
    return program()
}

// ------------------------------------------------------------------
// Interpreter
// ------------------------------------------------------------------

binary_funcs = {
    DIVIDE:   func(l, r) { return l / r },
    EQUAL:    func(l, r) { return l == r },
    GT:       func(l, r) { return l > r },
    GTE:      func(l, r) { return l >= r },
    IN:       func(l, r) { return l in r },
    LT:       func(l, r) { return l < r },
    LTE:      func(l, r) { return l <= r },
    MINUS:    func(l, r) { return l - r },
    MODULO:   func(l, r) { return l % r },
    NOTEQUAL: func(l, r) { return l != r },
    PLUS:     func(l, r) { return l + r },
    TIMES:    func(l, r) { return l * r },
}

unary_funcs = {
    MINUS:    func(v) { return -v },
    NOT:      func(v) { return not v },
}

// Remove the first arg (source filename) from target's args()
_args = args()
func args() {
    return slice(_args, 1, len(_args))
}

builtins = {
    "append": append,
    "args": args,
    "char": char,
    "exit": exit,
    "find": find,
    "int": int,
    "join": join,
    "len": len,
    "lower": lower,
    "print": print,
    "range": range,
    "read": read,
    "rune": rune,
    "slice": slice,
    "sort": sort,
    "split": split,
    "str": str,
    "type": type,
    "upper": upper,
}

func execute(program) {
    interp = {}
    interp.vars = []

    func error(msg) {
        print("execute error : " + msg)
        exit(1)
    }

    func push_scope(scope) {
        append(interp.vars, scope)
    }

    func pop_scope() {
        interp.vars = slice(interp.vars, 0, len(interp.vars)-1)
    }

    func assign(name, value) {
        interp.vars[len(interp.vars)-1][name] = value
    }

    func lookup(name) {
        i = len(interp.vars) - 1
        while i >= 0 {
            vars = interp.vars[i]
            if name in vars {
                return [vars[name], true]
            }
            i = i - 1
        }
        return [nil, false]
    }

    func user_function(name, params, ellipsis, body, closure) {
        f = func(args...) {
            if ellipsis {
                ellipsis_args = slice(args, len(params)-1, len(args))
                new_args = slice(args, 0, len(params)-1)
                append(new_args, ellipsis_args)
                args = new_args
            }
            push_scope(closure)
            push_scope({})
            for i in range(len(args)) {
                assign(params[i], args[i])
            }
            r = execute_block(body)
            pop_scope()
            pop_scope()
            if r != nil {
                return r[0]
            }
        }
        return f
    }

    func evaluate(e) {
        if e.type == "Binary" {
            if e.op in binary_funcs {
                return binary_funcs[e.op](evaluate(e.left), evaluate(e.right))
            } else if e.op == AND {
                return evaluate(e.left) and evaluate(e.right)
            } else {
                return evaluate(e.left) or evaluate(e.right)
            }
        } else if e.type == "Unary" {
            return unary_funcs[e.operator](evaluate(e.operand))
        } else if e.type == "Call" {
            function = evaluate(e.function)
            args = []
            for a in e.args {
                append(args, evaluate(a))
            }
            if e.ellipsis {
                ellipsis_arg = args[len(args)-1]
                args = slice(args, 0, len(args)-1)
                for a in ellipsis_arg {
                    append(args, a)
                }
            }
            return function(args...)
        } else if e.type == "Literal" {
            return e.value
        } else if e.type == "Variable" {
            value_found = lookup(e.name)
            if not value_found[1] {
                error("name \"" + e.name + "\" not found")
            }
            return value_found[0]
        } else if e.type == "List" {
            values = []
            for v in e.values {
                append(values, evaluate(v))
            }
            return values
        } else if e.type == "Map" {
            value = {}
            for item in e.items {
                value[evaluate(item[0])] = evaluate(item[1])
            }
            return value
        } else if e.type == "Subscript" {
            container = evaluate(e.container)
            subscript = evaluate(e.subscript)
            return container[subscript]
        } else {
            // FunctionExpression
            closure = interp.vars[len(interp.vars)-1]
            return user_function("", e.params, e.ellipsis, e.body, closure)
        }
    }

    func execute_statement(s) {
        if s.type == "Assign" {
            if s.target.type == "Variable" {
                assign(s.target.name, evaluate(s.value))
            } else {
                container = evaluate(s.target.container)
                subscript = evaluate(s.target.subscript)
                container[subscript] = evaluate(s.value)
            }
        } else if s.type == "If" {
            cond = evaluate(s.condition)
            if cond {
                r = execute_block(s.body)
                if r != nil {
                    return r
                }
            } else if s.else_body != nil {
                r = execute_block(s.else_body)
                if r != nil {
                    return r
                }
            }
        } else if s.type == "While" {
            while evaluate(s.condition) {
                r = execute_block(s.body)
                if r != nil {
                    return r
                }
            }
        } else if s.type == "For" {
            for elem in evaluate(s.iterable) {
                assign(s.name, elem)
                r = execute_block(s.body)
                if r != nil {
                    return r
                }
            }
        } else if s.type == "ExpressionStatement" {
            evaluate(s.expr)
        } else if s.type == "FunctionDefinition" {
            closure = interp.vars[len(interp.vars)-1]
            assign(s.name, user_function(s.name, s.params, s.ellipsis, s.body, closure))
        } else {
            // Return
            return [evaluate(s.result)]
        }
    }

    func execute_block(block) {
        for statement in block.statements {
            r = execute_statement(statement)
            if r != nil {
                return r
            }
        }
    }

    push_scope({})
    for name in builtins {
        assign(name, builtins[name])
    }
    r = execute_block(program.body)
    if r != nil {
        error("can't return at top level")
    }
}

if len(_args) == 0 {
    print("usage: littlelang littlelang.ll source_filename")
    exit(1)
}
source = read(_args[0])
prog = parse(source)
execute(prog)
