//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockIF interface {
	Do()
}

func TestMustNotNil_NoNil(t *testing.T) {
	var m mockIF = &mockImpl{}
	assert.NotPanics(t, func() {
		mustNotNil("Test", map[string]any{"m": m})
	})
}

func TestMustNotNil_NilInterface(t *testing.T) {
	assert.PanicsWithValue(t, "Test: m must not be nil", func() {
		mustNotNil("Test", map[string]any{"m": nil})
	})
}

func TestMustNotNil_NilConcreteInterface(t *testing.T) {
	var m mockIF
	assert.PanicsWithValue(t, "Test: m must not be nil", func() {
		mustNotNil("Test", map[string]any{"m": m})
	})
}

type mockImpl struct{}

func (m *mockImpl) Do() {}
