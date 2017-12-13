// Interpreter/runtime error types for littlelang (basically boilerplate)

package interpreter

import (
	"fmt"

	. "littlelang/tokenizer"
)

// Error is the error type returned by Evaluate and Execute. Each error holds
// the position of the error in the source and the error message, which can be
// queried on the type or via Error().
type Error interface {
	error
	Position() Position
}

// TypeError is returned for invalid types and wrong number of arguments.
type TypeError struct {
	Message string
	pos     Position
}

func (e TypeError) Error() string {
	return fmt.Sprintf("type error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

func (e TypeError) Position() Position {
	return e.pos
}

func typeError(pos Position, format string, args ...interface{}) error {
	return TypeError{fmt.Sprintf(format, args...), pos}
}

// ValueError is returned for invalid values (out of bounds index, etc).
type ValueError struct {
	Message string
	pos     Position
}

func (e ValueError) Error() string {
	return fmt.Sprintf("value error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

func (e ValueError) Position() Position {
	return e.pos
}

func valueError(pos Position, format string, args ...interface{}) error {
	return ValueError{fmt.Sprintf(format, args...), pos}
}

// NameError is returned when a variable is not found.
type NameError struct {
	Message string
	pos     Position
}

func (e NameError) Error() string {
	return fmt.Sprintf("name error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

func (e NameError) Position() Position {
	return e.pos
}

func nameError(pos Position, format string, args ...interface{}) error {
	return NameError{fmt.Sprintf(format, args...), pos}
}

// RuntimeError is returned for other or internal runtime errors.
type RuntimeError struct {
	Message string
	pos     Position
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("runtime error at %d:%d: %s", e.pos.Line, e.pos.Column, e.Message)
}

func (e RuntimeError) Position() Position {
	return e.pos
}

func runtimeError(pos Position, format string, args ...interface{}) error {
	return RuntimeError{fmt.Sprintf(format, args...), pos}
}
