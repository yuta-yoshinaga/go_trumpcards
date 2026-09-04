//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBuraCuiPresenter_ShowsWinningCombinationsFromDomainSource(t *testing.T) {
	b := domain.NewDefaultBura()
	b.Reset()

	out := new(BuraCuiPresenter).Output(b, nil)

	// Verify all winning combinations from domain.BuraWinningCombinations() appear in the output.
	combos := domain.BuraWinningCombinations()
	require.NotEmpty(t, combos)
	for _, c := range combos {
		key := c.Key()
		require.NotEmpty(t, key)
		translated := i18n.T("bura.combo." + key)
		require.NotEqual(t, "bura.combo."+key, translated, "translation must exist for domain combo")
		assert.Contains(t, out, translated)
	}

	// Negative control: Ensure untranslated raw combo keys never leak to output.
	assert.NotContains(t, out, "bura.combo.")

	// Ensure existing output elements are preserved.
	assert.Contains(t, out, "トリック")
	assert.Contains(t, out, "切札:")
	assert.Contains(t, out, "あなた")
	assert.Contains(t, out, "CPU")
	assert.Contains(t, out, "----------")
}

func TestBuraCuiPresenter_ShowsWinningCombinationsFromDomainSource_English(t *testing.T) {
	origLang := i18n.Lang()
	i18n.SetLang("en")
	defer i18n.SetLang(origLang)

	b := domain.NewDefaultBura()
	b.Reset()

	out := new(BuraCuiPresenter).Output(b, nil)

	combos := domain.BuraWinningCombinations()
	require.NotEmpty(t, combos)
	for _, c := range combos {
		key := c.Key()
		require.NotEmpty(t, key)
		translated := i18n.T("bura.combo." + key)
		require.NotEqual(t, "bura.combo."+key, translated, "translation must exist for domain combo")
		assert.Contains(t, out, translated)
	}

	assert.NotContains(t, out, "bura.combo.")

	assert.Contains(t, out, "trick")
	assert.Contains(t, out, "trump:")
	assert.Contains(t, out, "You")
	assert.Contains(t, out, "CPU")
	assert.Contains(t, out, "----------")
}

func TestBuraCuiPresenter_WinningCombosLine_DropsUntranslatedSilently(t *testing.T) {
	line := buraWinningCombosLine(domain.BuraWinningCombinations())
	assert.NotEmpty(t, line)
	assert.NotContains(t, line, "bura.combo.")
}

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

// TestBuraWinningCombosLine_DropsWhatItCannotName は落とす側の 2 分岐を直接見る。
//
// Output 経由では、実在する 4 役はどれもキーを持ち訳もあるので、
// 「キーが空」も「訳が 1 つも無い」も一度も通らない。
func TestBuraWinningCombosLine_DropsWhatItCannotName(t *testing.T) {
	i18n.SetLang("ja")

	t.Run("a combination with no key is dropped", func(t *testing.T) {
		line := buraWinningCombosLine([]domain.BuraCombination{
			domain.BuraCombinationNone,
			domain.BuraCombinationBura,
		})
		assert.Contains(t, line, i18n.T("bura.combo.bura"))
		assert.NotContains(t, line, "bura.combo.")
	})

	t.Run("nothing nameable means no line at all", func(t *testing.T) {
		// 役なししか無ければ出す中身が無い。空行を出しても読めない。
		assert.Empty(t, buraWinningCombosLine([]domain.BuraCombination{domain.BuraCombinationNone}))
		assert.Empty(t, buraWinningCombosLine(nil))
	})
}
