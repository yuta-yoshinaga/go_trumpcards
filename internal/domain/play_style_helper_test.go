package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayStyleName(t *testing.T) {
	names := []string{"Alpha", "Beta", "Gamma"}

	t.Run("valid index 0", func(t *testing.T) {
		assert.Equal(t, "Alpha", playStyleName(0, names))
	})

	t.Run("valid index last", func(t *testing.T) {
		assert.Equal(t, "Gamma", playStyleName(2, names))
	})

	t.Run("out of range positive", func(t *testing.T) {
		assert.Equal(t, "Unknown", playStyleName(3, names))
	})

	t.Run("negative index", func(t *testing.T) {
		assert.Equal(t, "Unknown", playStyleName(-1, names))
	})

	t.Run("empty names slice", func(t *testing.T) {
		assert.Equal(t, "Unknown", playStyleName(0, []string{}))
	})
}
