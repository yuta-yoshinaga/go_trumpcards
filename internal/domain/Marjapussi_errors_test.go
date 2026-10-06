//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func assertMarjapussiCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	assert.Equal(t, code, err.(*DomainError).MessageCode())
}

func TestMarjapussiDomainErrorsHaveMessageCodes(t *testing.T) {
	assert.Equal(t, "リードスートに従ってください", i18n.TForLang("ja", "shared.errFollowLeadSuit"))
	assert.Equal(t, "Please follow the lead suit.", i18n.TForLang("en", "shared.errFollowLeadSuit"))
	g := NewDefaultMarjapussi()
	g.Reset()
	g.SetPhase(MarjapussiPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertMarjapussiCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "marjapussi.errCardIndexOutOfRange")

	g.GetPlayer(0).ResetRound()
	g.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 7, false))
	g.GetPlayer(0).AddCard(NewCard(CardDesignClover, 7, false))
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 8, false)}})
	assertMarjapussiCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "shared.errFollowLeadSuit")

	g.GetPlayer(0).ResetRound()
	g.SetTrumpSuit(CardDesignClover)
	g.GetPlayer(0).AddCard(NewCard(CardDesignClover, 7, false))
	g.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 7, false))
	g.SetCurrentTrick([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 8, false)}})
	assertMarjapussiCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "marjapussi.errMustPlayTrump")
}
