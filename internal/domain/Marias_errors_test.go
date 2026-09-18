//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestMariasErrorsReturnMessageCodes(t *testing.T) {
	players := []*domain.MariasPlayer{
		domain.NewMariasPlayer(true), domain.NewMariasPlayer(false), domain.NewMariasPlayer(false),
	}
	g := domain.NewMarias(domain.NewTrumpCardsBelote(), players, domain.DefaultMariasConfig())
	g.Reset()
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerPlay(-1)
	assertErrorCode(t, err, domain.ErrInvalidCard, "marias.errCardIndexOutOfRange")

	g.SetTrumpSuit(domain.CardDesignDiamond)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 1, false)}})
	players[0].Reset()
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	err = g.PlayerPlay(0)
	assertErrorCode(t, err, domain.ErrInvalidPlay, "marias.errMustPlayTrump")

	players[0].Reset()
	players[0].AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	err = g.PlayerPlay(1)
	assertErrorCode(t, err, domain.ErrInvalidPlay, "marias.errMustFollowSuit")
}
