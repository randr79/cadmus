package argument

import (
	"fmt"
	"unicode/utf8"
)

type stringArgument[T ~string] struct {
	value  T
	errors []error
}

// NewStringArgument start de keten voor een string
func NewStringArgument[T ~string](value string) *stringArgument[T] {
	return &stringArgument[T]{
		value: T(value),
	}
}

// MinLen controleert de minimale lengte (veilig voor UTF-8 karakters!)
func (c *stringArgument[T]) MinLen(min int) *stringArgument[T] {
	// RuneCountInString telt echte karakters (zoals 'é'), niet bytes!
	if utf8.RuneCountInString(string(c.value)) < min {
		c.errors = append(c.errors, fmt.Errorf("value length must be >= %d", min))
	}
	return c
}

// MaxLen controleert de maximale lengte
func (c *stringArgument[T]) MaxLen(max int) *stringArgument[T] {
	if utf8.RuneCountInString(string(c.value)) > max {
		c.errors = append(c.errors, fmt.Errorf("value length must be <= %d", max))
	}
	return c
}

func (c *stringArgument[T]) Validate(errors *[]error) T {
	*errors = append(*errors, c.errors...)
	return c.value
}
