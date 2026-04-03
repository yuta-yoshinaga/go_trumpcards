//go:build test
// +build test

package presenter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestLookupHintReason(t *testing.T) {
	gameReasons := map[string]string{
		"pass_high_risk_cards": "リスクの高いカードを渡す",
	}

	t.Run("game-specific reason", func(t *testing.T) {
		assert.Equal(t, "リスクの高いカードを渡す", lookupHintReason("pass_high_risk_cards", gameReasons))
	})
	t.Run("shared reason", func(t *testing.T) {
		assert.Equal(t, "リードスートに追随", lookupHintReason("follow_suit", gameReasons))
	})
	t.Run("game-specific overrides shared", func(t *testing.T) {
		override := map[string]string{"follow_suit": "カスタム"}
		assert.Equal(t, "カスタム", lookupHintReason("follow_suit", override))
	})
	t.Run("unknown reason returns key", func(t *testing.T) {
		assert.Equal(t, "unknown_reason", lookupHintReason("unknown_reason", gameReasons))
	})
	t.Run("nil game reasons uses shared", func(t *testing.T) {
		assert.Equal(t, "低いカードでリード", lookupHintReason("lead_low", nil))
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
