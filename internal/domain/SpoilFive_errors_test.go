//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSpoilFiveErrorsReturnMessageCodes(t *testing.T) {
	players := make([]*domain.SpoilFivePlayer, domain.SpoilFivePlayerCnt)
	for i := range players {
		players[i] = domain.NewSpoilFivePlayer(i == 0)
	}
	g := domain.NewSpoilFive(domain.NewTrumpCards(0), players, domain.DefaultSpoilFiveConfig())
	g.Reset()
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerPlay(-1)
	assertErrorCode(t, err, domain.ErrInvalidCard, "spoilfive.errCardIndexOutOfRange")

	g.SetTrumpSuit(domain.CardDesignDiamond)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 7, false)}})
	players[0].Reset()
	players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	err = g.PlayerPlay(1)
	assertErrorCode(t, err, domain.ErrInvalidPlay, "spoilfive.errMustFollowTrump")

	players[0].Reset()
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 1, false)}})
	err = g.PlayerPlay(1)
	assertErrorCode(t, err, domain.ErrInvalidPlay, "spoilfive.errMustFollowSuit")
}
