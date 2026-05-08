package arguments

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"unsafe"

	"golang.org/x/exp/constraints"
)

type numericArgument[T constraints.Integer | constraints.Float] struct {
	value  T
	errors []error
}

func friendlyErr(value string, kind string, err error) []error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, strconv.ErrRange):
		return []error{fmt.Errorf("value %s is out of range for type %s", value, kind)}
	case errors.Is(err, strconv.ErrSyntax):
		return []error{fmt.Errorf("value %s is not a valid %s", value, kind)}
	default:
		return []error{err}
	}
}

func NewIntArgument[T constraints.Signed](value string) *numericArgument[T] {
	result := numericArgument[T]{}
	bitsize := int(unsafe.Sizeof(result.value) * 8)
	val, err := strconv.ParseInt(value, 10, bitsize)
	result.value = T(val)
	result.errors = friendlyErr(value, fmt.Sprintf("int%d", bitsize), err)
	return &result
}
func NewUintArgument[T constraints.Unsigned](value string) *numericArgument[T] {
	result := numericArgument[T]{}
	bitsize := int(unsafe.Sizeof(result.value) * 8)
	val, err := strconv.ParseUint(value, 10, bitsize)
	result.value = T(val)
	result.errors = friendlyErr(value, fmt.Sprintf("uint%d", bitsize), err)
	return &result
}
func NewFloatArgument[T constraints.Float](value string) *numericArgument[T] {
	result := numericArgument[T]{}
	bitsize := int(unsafe.Sizeof(result.value) * 8)
	val, err := strconv.ParseFloat(value, bitsize)
	result.value = T(val)
	result.errors = friendlyErr(value, fmt.Sprintf("float%d", bitsize), err)

	return &result
}

func (c *numericArgument[T]) Gte(min T) *numericArgument[T] {
	if c.value < min {
		c.errors = append(c.errors, fmt.Errorf("value %v must be >= %v", c.value, min))
	}
	return c
}

func (c *numericArgument[T]) Lte(max T) *numericArgument[T] {
	if c.value > max {
		c.errors = append(c.errors, fmt.Errorf("value %v must be <= %v", c.value, max))
	}
	return c
}

func (c *numericArgument[T]) Gt(min T) *numericArgument[T] {
	if c.value <= min {
		c.errors = append(c.errors, fmt.Errorf("value %v must be > %v", c.value, min))
	}
	return c
}

func (c *numericArgument[T]) Lt(max T) *numericArgument[T] {
	if c.value >= max {
		c.errors = append(c.errors, fmt.Errorf("value %v must be < %v", c.value, max))
	}
	return c
}

func (c *numericArgument[T]) Between(min T, max T) *numericArgument[T] {
	if !(c.value >= min && c.value <= max) {
		c.errors = append(c.errors, fmt.Errorf("value %v must be betwween %v and %v ()inclusive", c.value, min, max))
	}
	return c
}

func (c *numericArgument[T]) MultipleOf(factor float64) *numericArgument[T] {
	const epsilon = 1e-9
	if factor == 0 {
		c.errors = append(c.errors, fmt.Errorf("factor for MultipleOf cannot be zero"))
	} else if remainder := math.Mod(float64(c.value), factor); math.Abs(remainder) >= epsilon && math.Abs(remainder-factor) >= epsilon {
		c.errors = append(c.errors, fmt.Errorf("value %v must be a multiple of %v", c.value, factor))
	}
	return c
}

// Validate voegt de error toe aan de centrale lijst en geeft de waarde terug!
func (c *numericArgument[T]) Validate(errors *[]error) T {
	*errors = append(*errors, c.errors...)
	return c.value
}
