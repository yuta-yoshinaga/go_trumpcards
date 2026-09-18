//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBrusquembilleErrorsReturnMessageCodes(t *testing.T) {
	newGame := func() *domain.Brusquembille {
		return domain.NewBrusquembille(
			domain.NewTrumpCardsWithDecks(0, 0),
			domain.NewBrusquembillePlayersForTable(2),
			domain.DefaultBrusquembilleConfig(),
		)
	}

	g := newGame()
	assertErrorCode(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "brusquembille.errCardIndexOutOfRange")

	g = newGame()
	g.GetPlayer(0).AddCard(nil)
	assertErrorCode(t, g.PlayerPlay(0), domain.ErrInvalidCard, "brusquembille.errCardNil")

	g = newGame()
	lead := domain.NewCard(domain.CardDesignSpade, 5, false)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: lead}})
	g.GetPlayer(0).AddCard(lead)
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	assertErrorCode(t, g.PlayerPlay(1), domain.ErrInvalidCard, "brusquembille.errMustFollowSuit")
}
