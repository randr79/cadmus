package types

import (
	"encoding"
	"fmt"
	"reflect"
)

type textArgument[T any] struct {
	value  T
	errors []error
}

func NewTextArgument[T any](value string) *textArgument[T] {
	result := textArgument[T]{}

	if _, ok := any(result.value).(encoding.TextUnmarshaler); ok {
		//pointer type, result is a nil pointer, need to do new through reflection
		v := reflect.New(reflect.TypeFor[T]().Elem()).Interface()
		ut := v.(encoding.TextUnmarshaler)
		if err := ut.UnmarshalText([]byte(value)); err != nil {
			result.errors = []error{err}
		} else {
			result.value = v.(T)
		}
	} else if ut, ok := any(&result.value).(encoding.TextUnmarshaler); ok {
		//value type, ut is an interface to a pointer to result
		if err := ut.UnmarshalText([]byte(value)); err != nil {
			result.errors = []error{err}
		}
	} else {
		result.errors = []error{fmt.Errorf("type `%T` does not implement TextUnmarshaler", result)}
	}
	return &result
}

func (c *textArgument[T]) Validate(errors *[]error) T {
	*errors = append(*errors, c.errors...)
	// Alleen valideren als er nog GEEN errors zijn
	if len(c.errors) > 0 {
		//skip validation
	} else if sv, ok := any(c.value).(interface{ Validate() error }); ok {
		if err := sv.Validate(); err != nil {
			c.errors = append(c.errors, err)
		}
	} else if sv, ok := any(&c.value).(interface{ Validate() error }); ok {
		if err := sv.Validate(); err != nil {
			c.errors = append(c.errors, err)
		}
	}
	return c.value
}
