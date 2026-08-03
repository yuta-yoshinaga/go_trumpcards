//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBuraCuiPresenter_ShowsTheHumanHandAndOnlyCountsForTheCpu(t *testing.T) {
	b := domain.NewDefaultBura()
	b.Reset()

	out := new(BuraCuiPresenter).Output(b, nil)

	// The human's cards are indexed so `p <i>` can name one.
	assert.Contains(t, out, "[0]")
	// Exactly ONE indexed hand is printed. If the CPU's cards were rendered
	// they would carry their own [0], so this count is the leak check.
	// (Counting "[" would not work -- ANSI colour escapes contain one.)
	assert.Equal(t, 1, strings.Count(out, "[0]"), "only the human hand may be printed")
	assert.Contains(t, out, "手札3枚", "the CPU's count is public and must still be shown")
}

func TestBuraCuiPresenter_ShowsTheLeadWhileATrickIsOpen(t *testing.T) {
	b := domain.NewDefaultBura()
	b.Reset()
	require.NoError(t, b.PlayCards(0, []int{0}))

	out := new(BuraCuiPresenter).Output(b, nil)
	assert.NotEmpty(t, out)
}

func TestBuraCuiPresenter_ReportsEachEnding(t *testing.T) {
	t.Run("a short claim loses", func(t *testing.T) {
		b := domain.NewDefaultBura()
		b.Reset()
		require.NoError(t, b.Claim(0))
		assert.NotEmpty(t, new(BuraCuiPresenter).Output(b, nil))
	})

	t.Run("a true claim wins", func(t *testing.T) {
		b := domain.NewDefaultBura()
		b.Reset()
		b.SetPlayerPoints(0, domain.BuraWinThreshold)
		require.NoError(t, b.Claim(0))
		assert.NotEmpty(t, new(BuraCuiPresenter).Output(b, nil))
	})
}

func TestBuraCuiPresenter_RendersAnError(t *testing.T) {
	b := domain.NewDefaultBura()
	b.Reset()
	out := new(BuraCuiPresenter).Output(b, assert.AnError)
	assert.Contains(t, out, assert.AnError.Error())
}

func TestBuraCuiPresenter_HintResolvesItsReasonKey(t *testing.T) {
	// The Web presenter ships the reason identifier and the frontend looks it
	// up; the CUI has to resolve it here, so an unmapped identifier would print
	// the raw key.
	for range 100 {
		b := domain.NewDefaultBura()
		b.Reset()
		out := new(BuraCuiPresenter).HintOutput(b)
		assert.NotContains(t, out, "bura.hint.", "the reason identifier must be translated, not printed raw")
		assert.NotEmpty(t, strings.TrimSpace(out))
	}
}

func TestBuraCuiPresenter_ActionLogRenders(t *testing.T) {
	b := domain.NewDefaultBura()
	b.Reset()
	require.NoError(t, b.Claim(0))
	assert.NotEmpty(t, new(BuraCuiPresenter).ActionLogOutput(b))
}
