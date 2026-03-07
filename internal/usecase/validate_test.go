//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockIF interface {
	Do()
}

type mockImpl struct{}

func (m *mockImpl) Do() {}

func TestMustNotNil_NoNil(t *testing.T) {
	var m mockIF = &mockImpl{}
	assert.NotPanics(t, func() {
		mustNotNil("Test", map[string]any{"m": m})
	})
}

func TestMustNotNil_UntypedNil(t *testing.T) {
	assert.PanicsWithValue(t, "Test: m must not be nil", func() {
		mustNotNil("Test", map[string]any{"m": nil})
	})
}

func TestMustNotNil_NilInterface(t *testing.T) {
	var m mockIF
	assert.PanicsWithValue(t, "Test: m must not be nil", func() {
		mustNotNil("Test", map[string]any{"m": m})
	})
}

func TestMustNotNil_TypedNilPointerInInterface(t *testing.T) {
	var impl *mockImpl
	var m mockIF = impl
	assert.PanicsWithValue(t, "Test: m must not be nil", func() {
		mustNotNil("Test", map[string]any{"m": m})
	})
}

func TestMustNotNil_NonNilableTypes(t *testing.T) {
	assert.NotPanics(t, func() {
		mustNotNil("Test", map[string]any{
			"s":  "hello",
			"i":  123,
			"st": struct{}{},
		})
	})
}
