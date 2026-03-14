//go:build test

package webutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClampIntPtr(t *testing.T) {
	t.Run("nil pointer returns defaultVal", func(t *testing.T) {
		assert.Equal(t, 5, ClampIntPtr(nil, 1, 10, 5))
	})

	t.Run("value within range returns value", func(t *testing.T) {
		v := 3
		assert.Equal(t, 3, ClampIntPtr(&v, 1, 10, 5))
	})

	t.Run("value below min returns defaultVal", func(t *testing.T) {
		v := 0
		assert.Equal(t, 5, ClampIntPtr(&v, 1, 10, 5))
	})

	t.Run("value above max returns defaultVal", func(t *testing.T) {
		v := 11
		assert.Equal(t, 5, ClampIntPtr(&v, 1, 10, 5))
	})

	t.Run("value at exact min returns value", func(t *testing.T) {
		v := 1
		assert.Equal(t, 1, ClampIntPtr(&v, 1, 10, 5))
	})

	t.Run("value at exact max returns value", func(t *testing.T) {
		v := 10
		assert.Equal(t, 10, ClampIntPtr(&v, 1, 10, 5))
	})
}
