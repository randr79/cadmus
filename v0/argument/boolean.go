package argument

import (
	"fmt"
	"strconv"
)

type BoolArgument[T ~bool] struct {
	value  T
	errors []error
}

func NewBoolArgument[T ~bool](value string) *BoolArgument[T] {
	result := BoolArgument[T]{}
	val, err := strconv.ParseBool(value)
	result.value = T(val)
	if err != nil {
		result.errors = []error{fmt.Errorf("value %s is not a valid boolean", value)}
	}
	return &result
}

func (c *BoolArgument[T]) MustBe(expected T) *BoolArgument[T] {
	if c.value != expected {
		c.errors = append(c.errors, fmt.Errorf("value must be %v", expected))
	}
	return c
}

func (c *BoolArgument[T]) Validate(errors *[]error) T {
	*errors = append(*errors, c.errors...)
	return c.value
}
