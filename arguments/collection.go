package arguments

import (
	"strings"
)

func Apply[T any](data []string, transform func(string, *[]error) T, err *[]error) []T {
	result := make([]T, len(data))
	for i, v := range data {
		result[i] = transform(v, err)
	}
	return result
}

// De 'Selector' bepaalt hoe we van de []string naar de input van de transform gaan
func Map[V any, T any](
	data map[string][]string,
	prefix string,
	selector func([]string, func(string, *[]error) V, *[]error) T,
	transform func(string, *[]error) V,
	errs *[]error,
) map[string]T {
	result := make(map[string]T)
	for k, v := range data {
		if sk, ok := strings.CutPrefix(k, prefix); ok {
			result[sk] = selector(v, transform, errs)
		}
	}
	return result
}

// Voor ArgumentCount 1: pakt de eerste string en transformeert die naar V
func SelectFirst[V any](data []string, transform func(string, *[]error) V, errs *[]error) V {
	var first string
	if len(data) > 0 {
		first = data[0]
	}
	return transform(first, errs)
}

// Voor ArgumentCount -1: transformeert de hele slice naar []V (gebruikt je Apply)
func SelectAll[V any](data []string, transform func(string, *[]error) V, errs *[]error) []V {
	result := make([]V, len(data))
	for i, v := range data {
		result[i] = transform(v, errs)
	}
	return result
}
