package presenter

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestBlackJackCuiPresenter_suggestionStr(t *testing.T) {
	bjp := NewBlackJackCuiPresenter()
	cases := []struct {
		action domain.BJSuggestedAction
		want   string
	}{
		{domain.BJSuggestHit, "HIT"},
		{domain.BJSuggestStand, "STAND"},
		{domain.BJSuggestDouble, "DOUBLE"},
		{domain.BJSuggestSplit, "SPLIT"},
		{domain.BJSuggestSurrender, "SURRENDER"},
		{domain.BJSuggestDeclineInsurance, "DECLINE INSURANCE"},
		{domain.BJSuggestNone, ""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, bjp.suggestionStr(tc.action))
	}
}
