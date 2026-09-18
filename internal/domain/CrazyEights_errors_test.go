//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCrazyEightsErrorsReturnMessageCodes(t *testing.T) {
	players := []*domain.CrazyEightsPlayer{
		domain.NewCrazyEightsPlayer(true),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
	}
	g := domain.NewCrazyEights(domain.NewTrumpCardsWithDecks(0, 0), players, domain.DefaultCrazyEightsConfig())
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.CrazyEightsPhasePlay)

	assertErrorCode(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "crazyeights.errCardIndexOutOfRange")

	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	assertErrorCode(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "crazyeights.errCardCannotBePlayed")

	g.SetPhase(domain.CrazyEightsPhaseChooseSuit)
	assertErrorCode(t, g.PlayerChooseSuit(0), domain.ErrInvalidPlay, "crazyeights.errSuitOutOfRange")
}
