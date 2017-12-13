// Foreign function interface (FFI) for calling Go functions

// TODO: not finished

package interpreter

import (
	"fmt"
	"reflect"

	. "github.com/benhoyt/littlelang/tokenizer"
)

type nativeFunction struct {
	Function reflect.Value
	Name     string
}

func (f nativeFunction) call(interp *interpreter, pos Position, args []Value) Value {
	var getValue func(v reflect.Value) Value
	getValue = func(v reflect.Value) Value {
		x := v.Interface()
		switch v.Kind() {
		case reflect.Int:
			return Value(x)
		case reflect.Int8:
			return Value(x.(int))
		case reflect.Int16:
			return Value(x.(int))
		// case reflect.Int32:
		// 	return Value(x.(int))
		case reflect.Int64:
			return Value(x.(int))
		case reflect.Slice, reflect.Array:
			values := make([]Value, v.Len())
			for i := 0; i < v.Len(); i++ {
				values[i] = getValue(v.Index(i))
			}
			return Value(&values)
		case reflect.String:
			return Value(x)
		}
		panic(runtimeError(pos, fmt.Sprintf("native function returned invalid type %s", v.Kind())))
	}
	// Uint
	// Uint8
	// Uint16
	// Uint32
	// Uint64
	// Uintptr
	// Interface
	// Map
	// String
	// Struct

	// TODO: catch panics, convert args
	values := make([]reflect.Value, len(args))
	for i, a := range args {
		values[i] = reflect.ValueOf(a)
	}
	interp.stats.BuiltinCalls++
	results := f.Function.Call(values)
	if len(results) == 0 {
		return Value(nil)
	} else if len(results) == 1 {
		return getValue(results[0])
	} else {
		panic(runtimeError(pos, fmt.Sprintf("native function must return 0 or 1 result, not %d", len(results))))
	}
}

func (f nativeFunction) name() string {
	return fmt.Sprintf("<native %s>", f.Name)
}
