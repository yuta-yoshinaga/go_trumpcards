package usecase

import (
	"fmt"
	"reflect"
)

// mustNotNil validates that all provided values are non-nil.
// It panics with a descriptive message if any value is nil.
func mustNotNil(name string, params map[string]any) {
	for k, v := range params {
		if v == nil || reflect.ValueOf(v).IsNil() {
			panic(fmt.Sprintf("%s: %s must not be nil", name, k))
		}
	}
}
