// Package interpreter is a tree-walk interpreter for a littlelang AST.
//
// To interprete source code, you must first call parser.ParseExpression()
// or parser.ParseProgram(), and then call Evaluate or Execute, respectively.
//
package interpreter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"littlelang/parser"
	. "littlelang/tokenizer"
)

// Value is a littlelang runtime value (nil, bool, int, str, list, map, func).
type Value interface{}

// Config allows you to configure the interpreter's interaction with the
// outside world.
type Config struct {
	// Vars is a map of pre-defined variables to pass into the interpreter.
	Vars map[string]Value

	// Args is the list of command-line arguments for the interpreter's args()
	// builtin.
	Args []string

	// Stdin is the interpreter's standard input, for the read() builtin.
	// Defaults to os.Stdin if nil.
	Stdin io.Reader

	// Stdout is the interpreter's standard output, for the print() builtin.
	// Defaults to os.Stdout if nil.
	Stdout io.Writer

	// Exit is the function to call when the builtin exit() is called.
	// Defaults to os.Exit if nil.
	Exit func(int)
}

// Statistics about the interpreter from an Evaluate or Execute call.
type Stats struct {
	Ops          int
	UserCalls    int
	BuiltinCalls int
}

type interpreter struct {
	vars   []map[string]Value
	args   []string
	stdin  io.Reader
	stdout io.Writer
	exit   func(int)
	stats  Stats
}

type returnResult struct {
	value Value
	pos   Position
}

type binaryEvalFunc func(pos Position, l, r Value) Value

var binaryEvalFuncs = map[Token]binaryEvalFunc{
	DIVIDE:   evalDivide,
	EQUAL:    evalEqual,
	GT:       func(pos Position, l, r Value) Value { return evalLess(pos, r, l) },
	GTE:      func(pos Position, l, r Value) Value { return !evalLess(pos, l, r).(bool) },
	IN:       evalIn,
	LT:       evalLess,
	LTE:      func(pos Position, l, r Value) Value { return !evalLess(pos, r, l).(bool) },
	MINUS:    evalMinus,
	MODULO:   evalModulo,
	NOTEQUAL: func(pos Position, l, r Value) Value { return !evalEqual(pos, l, r).(bool) },
	PLUS:     evalPlus,
	TIMES:    evalTimes,
}

func evalEqual(pos Position, l, r Value) Value {
	switch l := l.(type) {
	case nil:
		return Value(r == nil)
	case bool:
		if r, rok := r.(bool); rok {
			return Value(l == r)
		}
	case int:
		if r, rok := r.(int); rok {
			return Value(l == r)
		}
	case string:
		if r, rok := r.(string); rok {
			return Value(l == r)
		}
	case *[]Value:
		if r, rok := r.(*[]Value); rok {
			if len(*l) != len(*r) {
				return Value(false)
			}
			for i, elem := range *l {
				if !evalEqual(pos, elem, (*r)[i]).(bool) {
					return Value(false)
				}
			}
			return Value(true)
		}
	case map[string]Value:
		if r, rok := r.(map[string]Value); rok {
			if len(l) != len(r) {
				return Value(false)
			}
			for k, v := range l {
				if !evalEqual(pos, v, r[k]).(bool) {
					return Value(false)
				}
			}
			return Value(true)
		}
	case functionType:
		if r, rok := r.(functionType); rok {
			return Value(l == r)
		}
	}
	return Value(false)
}

func evalIn(pos Position, l, r Value) Value {
	switch r := r.(type) {
	case string:
		if l, ok := l.(string); ok {
			return Value(strings.Index(r, l) >= 0)
		}
		panic(typeError(pos, "in str requires str on left side"))
	case *[]Value:
		for _, v := range *r {
			if evalEqual(pos, l, v).(bool) {
				return Value(true)
			}
		}
		return Value(false)
	case map[string]Value:
		if l, ok := l.(string); ok {
			_, present := r[l]
			return Value(present)
		}
		panic(typeError(pos, "in map requires str on left side"))
	}
	panic(typeError(pos, "in requires str, list, or map on right side"))
}

func evalLess(pos Position, l, r Value) Value {
	switch l := l.(type) {
	case int:
		if r, rok := r.(int); rok {
			return Value(l < r)
		}
	case string:
		if r, rok := r.(string); rok {
			return Value(l < r)
		}
	case *[]Value:
		if r, rok := r.(*[]Value); rok {
			for i := 0; i < len(*l) && i < len(*r); i++ {
				if !evalEqual(pos, (*l)[i], (*r)[i]).(bool) {
					return evalLess(pos, (*l)[i], (*r)[i])
				}
			}
			return Value(len(*l) < len(*r))
		}
	}
	panic(typeError(pos, "comparison requires two ints or two strs (or lists of ints or strs)"))
}

func evalPlus(pos Position, l, r Value) Value {
	switch l := l.(type) {
	case int:
		if r, rok := r.(int); rok {
			return Value(l + r)
		}
	case string:
		if r, rok := r.(string); rok {
			return Value(l + r)
		}
	case *[]Value:
		if r, rok := r.(*[]Value); rok {
			result := make([]Value, 0, len(*l)+len(*r))
			result = append(result, *l...)
			result = append(result, *r...)
			return Value(&result)
		}
	case map[string]Value:
		if r, rok := r.(map[string]Value); rok {
			result := make(map[string]Value)
			for k, v := range l {
				result[k] = v
			}
			for k, v := range r {
				result[k] = v
			}
			return Value(result)
		}
	}
	panic(typeError(pos, "+ requires two ints, strs, lists, or maps"))
}

func ensureInts(pos Position, l, r Value, operation string) (int, int) {
	li, lok := l.(int)
	ri, rok := r.(int)
	if !lok || !rok {
		panic(typeError(pos, "%s requires two ints", operation))
	}
	return li, ri
}

func evalMinus(pos Position, l, r Value) Value {
	li, ri := ensureInts(pos, l, r, "-")
	return Value(li - ri)
}

func evalTimes(pos Position, l, r Value) Value {
	switch l := l.(type) {
	case int:
		if r, rok := r.(int); rok {
			return Value(l * r)
		}
		if r, rok := r.(string); rok {
			if l < 0 {
				panic(valueError(pos, "can't multiply string by a negative number"))
			}
			return Value(strings.Repeat(r, l))
		}
	case string:
		if r, rok := r.(int); rok {
			if r < 0 {
				panic(valueError(pos, "can't multiply string by a negative number"))
			}
			return Value(strings.Repeat(l, r))
		}
	}
	panic(typeError(pos, "* requires two ints or a str and an int"))
}

func evalDivide(pos Position, l, r Value) Value {
	li, ri := ensureInts(pos, l, r, "/")
	if ri == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(li / ri)
}

func evalModulo(pos Position, l, r Value) Value {
	li, ri := ensureInts(pos, l, r, "%")
	if ri == 0 {
		panic(valueError(pos, "can't divide by zero"))
	}
	return Value(li % ri)
}

type unaryEvalFunc func(pos Position, v Value) Value

var unaryEvalFuncs = map[Token]unaryEvalFunc{
	NOT:   evalNot,
	MINUS: evalNegative,
}

func evalNot(pos Position, v Value) Value {
	if v, ok := v.(bool); ok {
		return Value(!v)
	}
	panic(typeError(pos, "not requires a bool"))
}

func evalNegative(pos Position, v Value) Value {
	if v, ok := v.(int); ok {
		return Value(-v)
	}
	panic(typeError(pos, "unary - requires an int"))
}

func evalSubscript(pos Position, container, subscript Value) Value {
	switch c := container.(type) {
	case string:
		if s, ok := subscript.(int); ok {
			if s < 0 || s >= len(c) {
				panic(valueError(pos, "subscript %d out of range", s))
			}
			return Value(string([]byte{c[s]}))
		}
		panic(typeError(pos, "str subscript must be an int"))
	case *[]Value:
		if s, ok := subscript.(int); ok {
			if s < 0 || s >= len(*c) {
				panic(valueError(pos, "subscript %d out of range", s))
			}
			return (*c)[s]
		}
		panic(typeError(pos, "list subscript must be an int"))
	case map[string]Value:
		if s, ok := subscript.(string); ok {
			if value, ok := c[s]; ok {
				return value
			}
			panic(valueError(pos, "key not found: %q", s))
		}
		panic(typeError(pos, "map subscript must be a str"))
	default:
		panic(typeError(pos, "can only subscript str, list, or map"))
	}
}

func (interp *interpreter) evalAnd(pos Position, le, re parser.Expression) Value {
	l := interp.evaluate(le)
	if l, ok := l.(bool); ok {
		if !l {
			// Short circuit: don't evaluate right if left false
			return Value(false)
		}
		r := interp.evaluate(re)
		if r, ok := r.(bool); ok {
			return Value(r)
		} else {
			panic(typeError(pos, "and requires two bools"))
		}
	} else {
		panic(typeError(pos, "and requires two bools"))
	}
}

func (interp *interpreter) evalOr(pos Position, le, re parser.Expression) Value {
	l := interp.evaluate(le)
	if l, ok := l.(bool); ok {
		if l {
			// Short circuit: don't evaluate right if left true
			return Value(true)
		}
		r := interp.evaluate(re)
		if r, ok := r.(bool); ok {
			return Value(r)
		} else {
			panic(typeError(pos, "or requires two bools"))
		}
	} else {
		panic(typeError(pos, "or requires two bools"))
	}
}

func (interp *interpreter) callFunction(pos Position, f functionType, args []Value) (ret Value) {
	defer func() {
		if r := recover(); r != nil {
			if result, ok := r.(returnResult); ok {
				ret = result.value
			} else {
				panic(r)
			}
		}
	}()
	return f.call(interp, pos, args)
}

func (interp *interpreter) evaluate(expr parser.Expression) Value {
	interp.stats.Ops++
	switch e := expr.(type) {
	case *parser.Binary:
		if f, ok := binaryEvalFuncs[e.Operator]; ok {
			return f(e.Position(), interp.evaluate(e.Left), interp.evaluate(e.Right))
		} else if e.Operator == AND {
			return interp.evalAnd(e.Position(), e.Left, e.Right)
		} else if e.Operator == OR {
			return interp.evalOr(e.Position(), e.Left, e.Right)
		}
		// Parser should never give us this
		panic(fmt.Sprintf("unknown binary operator %v", e.Operator))
	case *parser.Unary:
		if f, ok := unaryEvalFuncs[e.Operator]; ok {
			return f(e.Position(), interp.evaluate(e.Operand))
		}
		// Parser should never give us this
		panic(fmt.Sprintf("unknown unary operator %v", e.Operator))
	case *parser.Call:
		function := interp.evaluate(e.Function)
		if f, ok := function.(functionType); ok {
			args := []Value{}
			for _, a := range e.Arguments {
				args = append(args, interp.evaluate(a))
			}
			if e.Ellipsis {
				iterator := getIterator(e.Arguments[len(args)-1].Position(), args[len(args)-1])
				args = args[:len(args)-1]
				for iterator.HasNext() {
					args = append(args, iterator.Value())
				}
			}
			return interp.callFunction(e.Function.Position(), f, args)
		}
		panic(typeError(e.Function.Position(), "can't call non-function type %s", typeName(function)))
	case *parser.Literal:
		return Value(e.Value)
	case *parser.Variable:
		if v, ok := interp.lookup(e.Name); ok {
			return v
		}
		panic(nameError(e.Position(), "name %q not found", e.Name))
	case *parser.List:
		values := make([]Value, len(e.Values))
		for i, v := range e.Values {
			values[i] = interp.evaluate(v)
		}
		return Value(&values)
	case *parser.Map:
		value := make(map[string]Value)
		for _, item := range e.Items {
			key := interp.evaluate(item.Key)
			if k, ok := key.(string); ok {
				value[k] = interp.evaluate(item.Value)
			} else {
				panic(typeError(item.Key.Position(), "map key must be str, not %s", typeName(key)))
			}
		}
		return Value(value)
	case *parser.Subscript:
		container := interp.evaluate(e.Container)
		subscript := interp.evaluate(e.Subscript)
		return evalSubscript(e.Subscript.Position(), container, subscript)
	case *parser.FunctionExpression:
		closure := interp.vars[len(interp.vars)-1]
		return &userFunction{"", e.Parameters, e.Ellipsis, e.Body, closure}
	default:
		// Parser should never give us this
		panic(fmt.Sprintf("unexpected expression type %T", expr))
	}
}

func (interp *interpreter) pushScope(scope map[string]Value) {
	interp.vars = append(interp.vars, scope)
}

func (interp *interpreter) popScope() {
	interp.vars = interp.vars[:len(interp.vars)-1]
}

func (interp *interpreter) assign(name string, value Value) {
	interp.vars[len(interp.vars)-1][name] = value
}

func (interp *interpreter) lookup(name string) (Value, bool) {
	for i := len(interp.vars) - 1; i >= 0; i-- {
		thisVars := interp.vars[i]
		if v, ok := thisVars[name]; ok {
			return v, true
		}
	}
	return nil, false
}

func (interp *interpreter) executeBlock(block parser.Block) {
	for _, s := range block {
		interp.executeStatement(s)
	}
}

type iteratorType interface {
	HasNext() bool
	Value() Value
}

type listIterator struct {
	values []Value
	index  int
}

func (li *listIterator) HasNext() bool {
	return li.index < len(li.values)
}

func (li *listIterator) Value() Value {
	v := li.values[li.index]
	li.index++
	return v
}

func getIterator(pos Position, value Value) iteratorType {
	switch iterable := value.(type) {
	case string:
		strs := []Value{}
		for _, r := range iterable {
			strs = append(strs, string(r))
		}
		return &listIterator{strs, 0}
	case *[]Value:
		return &listIterator{*iterable, 0}
	case map[string]Value:
		keys := make([]Value, len(iterable))
		i := 0
		for key := range iterable {
			keys[i] = key
			i++
		}
		return &listIterator{keys, 0}
	default:
		panic(typeError(pos, "expected iterable (str, list, or map), got %s", typeName(value)))
	}
}

func (interp *interpreter) assignSubscript(pos Position, container, subscript, value Value) {
	switch c := container.(type) {
	case *[]Value:
		if s, ok := subscript.(int); ok {
			if s < 0 || s >= len(*c) {
				panic(valueError(pos, "subscript %d out of range", s))
			}
			(*c)[s] = value
		} else {
			panic(typeError(pos, "list subscript must be an int"))
		}
	case map[string]Value:
		if s, ok := subscript.(string); ok {
			c[s] = value
		} else {
			panic(typeError(pos, "map subscript must be a str"))
		}
	default:
		panic(typeError(pos, "can only assign to subscript of list or map"))
	}
}

func (interp *interpreter) executeStatement(s parser.Statement) {
	interp.stats.Ops++
	switch s := s.(type) {
	case *parser.Assign:
		switch target := s.Target.(type) {
		case *parser.Variable:
			interp.assign(target.Name, interp.evaluate(s.Value))
		case *parser.Subscript:
			container := interp.evaluate(target.Container)
			subscript := interp.evaluate(target.Subscript)
			value := interp.evaluate(s.Value)
			interp.assignSubscript(target.Subscript.Position(), container, subscript, value)
		default:
			// Parser should never get us here
			panic("can only assign to variable or subscript")
		}
	case *parser.If:
		cond := interp.evaluate(s.Condition)
		if c, ok := cond.(bool); ok {
			if c {
				interp.executeBlock(s.Body)
			} else if len(s.Else) > 0 {
				interp.executeBlock(s.Else)
			}
		} else {
			panic(typeError(s.Condition.Position(), "if condition must be bool, got %s", typeName(cond)))
		}
	case *parser.While:
		for {
			cond := interp.evaluate(s.Condition)
			if c, ok := cond.(bool); ok {
				if !c {
					break
				}
				interp.executeBlock(s.Body)
			} else {
				panic(typeError(s.Condition.Position(), "while condition must be bool, got %T", cond))
			}
		}
	case *parser.For:
		iterable := interp.evaluate(s.Iterable)
		iterator := getIterator(s.Iterable.Position(), iterable)
		for iterator.HasNext() {
			interp.assign(s.Name, iterator.Value())
			interp.executeBlock(s.Body)
		}
	case *parser.ExpressionStatement:
		interp.evaluate(s.Expression)
	case *parser.FunctionDefinition:
		closure := interp.vars[len(interp.vars)-1]
		interp.assign(s.Name, &userFunction{s.Name, s.Parameters, s.Ellipsis, s.Body, closure})
	case *parser.Return:
		result := interp.evaluate(s.Result)
		panic(returnResult{result, s.Position()})
	default:
		// Parser should never get us here
		panic(fmt.Sprintf("unexpected statement type %T", s))
	}
}

func (interp *interpreter) execute(prog *parser.Program) {
	for _, statement := range prog.Statements {
		interp.executeStatement(statement)
	}
}

func newInterpreter(config *Config) *interpreter {
	interp := new(interpreter)
	interp.pushScope(make(map[string]Value))
	for k, v := range builtins {
		interp.assign(k, v)
	}
	for k, v := range config.Vars {
		interp.assign(k, v)
	}
	interp.args = config.Args
	interp.stdin = config.Stdin
	if interp.stdin == nil {
		interp.stdin = os.Stdin
	}
	interp.stdout = config.Stdout
	if interp.stdout == nil {
		interp.stdout = os.Stdout
	}
	interp.exit = config.Exit
	if interp.exit == nil {
		interp.exit = os.Exit
	}
	return interp
}

// Evaluate takes a parsed Expression and interpreter config and evaluates the
// expression, returning the Value of the expression, interpreter statistics,
// and an error which is nil on success or an interpreter.Error if there's an
// error.
func Evaluate(expr parser.Expression, config *Config) (v Value, stats *Stats, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(Error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()
	interp := newInterpreter(config)
	v = interp.evaluate(expr)
	stats = &interp.stats
	return
}

// Execute takes a parsed Program and interpreter config and interprets the
// program. Return interpreter statistics, and an error which is nil on
// success or an interpreter.Error if there's an error.
func Execute(prog *parser.Program, config *Config) (stats *Stats, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case Error:
				err = e
			case returnResult:
				err = runtimeError(e.pos, "can't return at top level")
			default:
				panic(r)
			}
		}
	}()
	interp := newInterpreter(config)
	interp.execute(prog)
	stats = &interp.stats
	return
}
