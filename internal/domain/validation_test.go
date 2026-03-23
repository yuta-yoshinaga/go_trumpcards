//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRange(t *testing.T) {
	t.Run("in range", func(t *testing.T) {
		assert.NoError(t, ValidateRange("field", 5, 0, 10))
		assert.NoError(t, ValidateRange("field", 0, 0, 10))
		assert.NoError(t, ValidateRange("field", 10, 0, 10))
	})
	t.Run("below min", func(t *testing.T) {
		err := ValidateRange("field", -1, 0, 10)
		assert.EqualError(t, err, "field must be 0-10, got -1")
	})
	t.Run("above max", func(t *testing.T) {
		err := ValidateRange("field", 11, 0, 10)
		assert.EqualError(t, err, "field must be 0-10, got 11")
	})
}

func TestValidateMin(t *testing.T) {
	t.Run("at min", func(t *testing.T) {
		assert.NoError(t, ValidateMin("field", 1, 1))
	})
	t.Run("above min", func(t *testing.T) {
		assert.NoError(t, ValidateMin("field", 5, 1))
	})
	t.Run("below min", func(t *testing.T) {
		err := ValidateMin("field", 0, 1)
		assert.EqualError(t, err, "field must be >= 1, got 0")
	})
}
