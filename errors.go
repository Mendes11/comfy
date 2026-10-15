package comfy

import (
	"errors"
	"fmt"
	"reflect"
)

// TypeMismatchError is used whenever a type conversion fails
type TypeMismatchError struct {
	FieldName  string
	FieldValue string
	TargetType reflect.Type
	Err        error
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("cannot convert '%s' to type %s for field %s: %s", e.FieldValue, e.TargetType, e.FieldName, e.Err)
}

func (e TypeMismatchError) Unwrap() error {
	return e.Err
}

type MissingParserError struct {
	TargetType reflect.Type
}

func (e MissingParserError) Error() string {
	return fmt.Sprintf("don't know how to parse type '%s'. Missing implementing ComfyParser interface.", e.TargetType)
}

// Sentinel Errors

var (
	ErrExpectedPtr        error = errors.New("expected a pointer, not the value")
	ErrExpectedStructType error = errors.New("expected a struct")
)
