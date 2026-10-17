package presenter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestCuiHandNames_Localized(t *testing.T) {
	original := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(original) })

	for _, tc := range []struct {
		lang, poker, four string
	}{
		{"ja", "フォーカード", "フォーカード"},
		{"en", "Four of a Kind", "Four of a Kind"},
	} {
		t.Run(tc.lang, func(t *testing.T) {
			i18n.SetLang(tc.lang)
			require.Equal(t, tc.poker, cuiPokerHandName(7))
			require.Equal(t, tc.four, cuiFourCardHandName(8))
		})
	}
}
