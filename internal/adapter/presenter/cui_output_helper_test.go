//go:build test
// +build test

package presenter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestCuiTrickBlock(t *testing.T) {
	type fakeTrick struct {
		pidx    int
		cardStr string
	}
	names := map[int]string{0: "You", 1: "CPU1"}
	getIdx := func(tc fakeTrick) int { return tc.pidx }
	getCardStr := func(tc fakeTrick) string { return tc.cardStr }
	getName := func(i int) string { return names[i] }

	t.Run("empty trick", func(t *testing.T) {
		var b strings.Builder
		cuiTrickBlock(&b, []fakeTrick{}, getIdx, getCardStr, getName)
		assert.Equal(t, "", b.String())
	})
	t.Run("with cards", func(t *testing.T) {
		var b strings.Builder
		tricks := []fakeTrick{
			{pidx: 0, cardStr: "S-A"},
			{pidx: 1, cardStr: "H-5"},
		}
		cuiTrickBlock(&b, tricks, getIdx, getCardStr, getName)
		assert.Equal(t, "トリック: You=S-A, CPU1=H-5\n", b.String())
	})
	// Issue #1699 Phase 1: trick line label follows the active locale.
	t.Run("with cards (en)", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		var b strings.Builder
		cuiTrickBlock(&b, []fakeTrick{{pidx: 0, cardStr: "S-A"}}, getIdx, getCardStr, getName)
		assert.Equal(t, "Trick: You=S-A\n", b.String())
	})
}

func TestCuiErrorBlock(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		var b strings.Builder
		cuiErrorBlock(&b, nil)
		assert.Equal(t, "", b.String())
	})
	t.Run("with error", func(t *testing.T) {
		var b strings.Builder
		cuiErrorBlock(&b, fmt.Errorf("test error"))
		assert.Contains(t, b.String(), "test error")
	})
}

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
