//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCalabresellaErrorsReturnMessageCodes(t *testing.T) {
	players := []*domain.CalabresellaPlayer{
		domain.NewCalabresellaPlayer(true),
		domain.NewCalabresellaPlayer(true),
		domain.NewCalabresellaPlayer(true),
	}
	g := domain.NewCalabresella(domain.NewTrumpCards(0), players, domain.DefaultCalabresellaConfig())

	assertErrorCode(t, g.PlayerBid(domain.CalabresellaBid(3)), domain.ErrInvalidPlay, "calabresella.errBidMustExceed")

	g.SetPhase(domain.CalabresellaPhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	g.SetSoloistIdx(0)
	assertErrorCode(t, g.PlayerDiscard(-1), domain.ErrInvalidCard, "calabresella.errDiscardCardIndexOutOfRange")

	g.SetPhase(domain.CalabresellaPhasePlay)
	assertErrorCode(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "calabresella.errPlayCardIndexOutOfRange")
}
