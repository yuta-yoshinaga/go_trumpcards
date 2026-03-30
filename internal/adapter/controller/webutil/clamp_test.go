//go:build test

package webutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoundedIntPtr(t *testing.T) {
	t.Run("nil pointer returns defaultVal", func(t *testing.T) {
		assert.Equal(t, 5, BoundedIntPtr(nil, 1, 10, 5))
	})

	t.Run("value within range returns value", func(t *testing.T) {
		v := 3
		assert.Equal(t, 3, BoundedIntPtr(&v, 1, 10, 5))
	})

	t.Run("value below min returns defaultVal", func(t *testing.T) {
		v := 0
		assert.Equal(t, 5, BoundedIntPtr(&v, 1, 10, 5))
	})

	t.Run("value above max returns defaultVal", func(t *testing.T) {
		v := 11
		assert.Equal(t, 5, BoundedIntPtr(&v, 1, 10, 5))
	})

	t.Run("value at exact min returns value", func(t *testing.T) {
		v := 1
		assert.Equal(t, 1, BoundedIntPtr(&v, 1, 10, 5))
	})

	t.Run("value at exact max returns value", func(t *testing.T) {
		v := 10
		assert.Equal(t, 10, BoundedIntPtr(&v, 1, 10, 5))
	})
}

func TestBoolPtrOr(t *testing.T) {
	t.Run("nil pointer returns defaultVal true", func(t *testing.T) {
		assert.Equal(t, true, BoolPtrOr(nil, true))
	})

	t.Run("nil pointer returns defaultVal false", func(t *testing.T) {
		assert.Equal(t, false, BoolPtrOr(nil, false))
	})

	t.Run("non-nil true returns true", func(t *testing.T) {
		v := true
		assert.Equal(t, true, BoolPtrOr(&v, false))
	})

	t.Run("non-nil false returns false", func(t *testing.T) {
		v := false
		assert.Equal(t, false, BoolPtrOr(&v, true))
	})
}
