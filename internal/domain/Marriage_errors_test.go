//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestMarriageErrorsReturnMessageCodes(t *testing.T) {
	g := domain.NewDefaultMarriage()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	g.SetDiscardPile(nil)

	err := g.PlayerDrawFromDiscard()
	assertErrorCode(t, err, domain.ErrInvalidPlay, "marriage.errDiscardPileEmpty")

	g.SetPhase(domain.MarriagePhaseDiscard)
	err = g.PlayerDiscard(-1)
	assertErrorCode(t, err, domain.ErrInvalidCard, "marriage.errDiscardCardIndexOutOfRange")

	err = g.PlayerDeclare(-1)
	assertErrorCode(t, err, domain.ErrInvalidCard, "marriage.errDeclareCardIndexOutOfRange")
}
