//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestIndianRummyErrorsReturnMessageCodes(t *testing.T) {
	players := []*domain.IndianRummyPlayer{
		domain.NewIndianRummyPlayer(true),
		domain.NewIndianRummyPlayer(false),
	}
	g := domain.NewIndianRummy(domain.NewTrumpCardsWithDecks(0, 0), players, domain.DefaultIndianRummyConfig())
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.IndianRummyPhaseDraw)

	assertErrorCode(t, g.PlayerDrawFromDiscard(), domain.ErrInvalidPlay, "indianrummy.errDiscardPileEmpty")

	g.SetPhase(domain.IndianRummyPhaseDiscard)
	assertErrorCode(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "indianrummy.errDiscardCardIndexOutOfRange")
	assertErrorCode(t, g.PlayerDeclare(-1), domain.ErrInvalidCard, "indianrummy.errDeclareCardIndexOutOfRange")
}
