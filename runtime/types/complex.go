package types

import (
	"fmt"
	"strconv"
	"unsafe"

	"golang.org/x/exp/constraints"
)

type complexArgument[T constraints.Complex] struct {
	value  T
	errors []error
}

func NewComplexArgument[T constraints.Complex](value string) *complexArgument[T] {
	result := complexArgument[T]{}
	bitsize := int(unsafe.Sizeof(result.value) * 8)
	val, err := strconv.ParseComplex(value, bitsize)
	result.value = T(val)
	result.errors = friendlyErr(value, fmt.Sprintf("complex%d", bitsize), err)

	return &result
}

// Omdat we niet kunnen ordenen, valideren we op exacte gelijkheid
func (c *complexArgument[T]) Equals(expected T) *complexArgument[T] {
	if c.value != expected {
		c.errors = append(c.errors, fmt.Errorf("value %v must be exactly %v", c.value, expected))
	}
	return c
}
func (c *complexArgument[T]) Validate(errors *[]error) T {
	*errors = append(*errors, c.errors...)
	return c.value
}
