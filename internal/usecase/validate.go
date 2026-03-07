package usecase

import (
	"fmt"
	"reflect"
)

// mustNotNil validates that all provided values are non-nil.
// It panics with a descriptive message if any value is nil.
func mustNotNil(name string, params map[string]any) {
	for k, v := range params {
		isNil := v == nil
		if !isNil {
			val := reflect.ValueOf(v)
			switch val.Kind() {
			case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
				isNil = val.IsNil()
			}
		}
		if isNil {
			panic(fmt.Sprintf("%s: %s must not be nil", name, k))
		}
	}
}
