package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBlackJackCuiPresenter_suggestionStr(t *testing.T) {
	bjp := new(BlackJackCuiPresenter)
	cases := []struct {
		action domain.BJSuggestedAction
		want   string
	}{
		{domain.BJSuggestHit, "ヒット"},
		{domain.BJSuggestStand, "スタンド"},
		{domain.BJSuggestDouble, "ダブル"},
		{domain.BJSuggestDoubleStand, "ダブル"},
		{domain.BJSuggestSplit, "スプリット"},
		{domain.BJSuggestSurrender, "サレンダー"},
		{domain.BJSuggestDeclineInsurance, "インシュランスを断る"},
		{domain.BJSuggestNone, ""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, bjp.suggestionStr(tc.action))
	}
}
