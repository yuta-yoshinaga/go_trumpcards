//go:build test
// +build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCuiOutput(t *testing.T) {
	t.Run("wraps content with header and footer", func(t *testing.T) {
		result := buildCuiOutput("Test Game", func(b *strings.Builder) {
			b.WriteString("some content\n")
		})
		assert.Equal(t, "==========\nTest Game\n==========\nsome content\n==========\n", result)
	})

	t.Run("empty content", func(t *testing.T) {
		result := buildCuiOutput("Empty", func(b *strings.Builder) {})
		assert.Equal(t, "==========\nEmpty\n==========\n==========\n", result)
	})
}
