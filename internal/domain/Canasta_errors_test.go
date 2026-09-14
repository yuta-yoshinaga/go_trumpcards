//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertCanastaDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func canastaCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func canastaErrorGame(phase domain.CanastaPhase) (*domain.Canasta, *domain.CanastaPlayer) {
	g := newTestCanasta()
	g.Reset()
	g.SetPhase(phase)
	g.SetCurrentPlayerIdx(0)
	p := g.GetPlayer(0)
	p.Reset()
	return g, p
}

func TestCanasta_ValidationErrorsHaveCodes(t *testing.T) {
	t.Run("discard pile", func(t *testing.T) {
		g, _ := canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile(nil)
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard(nil), domain.ErrInvalidPlay, "canasta.errDiscardPileEmpty")

		g, _ = canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 3)})
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "canasta.errBlackThreeCannotTakeDiscardPile")

		g, _ = canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignJoker, 1)})
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "canasta.errWildCardCannotTakeDiscardPile")

		g, p := canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 7)})
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{0}), domain.ErrInvalidPlay, "canasta.errNaturalPairIndicesRequired")

		g, p = canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 7)})
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		p.AddCard(canastaCard(domain.CardDesignDiamond, 7))
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{-1, 1}), domain.ErrInvalidCard, "canasta.errCardIndexOutOfRange")

		g, p = canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 7)})
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{0, 0}), domain.ErrInvalidCard, "canasta.errSameCard")

		g, p = canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 7)})
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		p.AddCard(canastaCard(domain.CardDesignJoker, 1))
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "canasta.errPairMustBeNatural")

		g, p = canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 7)})
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		p.AddCard(canastaCard(domain.CardDesignDiamond, 8))
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "canasta.errPairRankMismatch")

		g, p = canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 4)})
		p.AddCard(canastaCard(domain.CardDesignHeart, 4))
		p.AddCard(canastaCard(domain.CardDesignDiamond, 4))
		assertCanastaDomainError(t, g.PlayerDrawFromDiscard([]int{0, 1}), domain.ErrInvalidPlay, "canasta.errInitialMeldMinimumNotMet")
	})

	t.Run("meld and discard", func(t *testing.T) {
		g, p := canastaErrorGame(domain.CanastaPhaseMeld)
		p.AddCard(canastaCard(domain.CardDesignSpade, 7))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{-1, 0, 1}}), domain.ErrInvalidCard, "canasta.errCardIndexOutOfRange")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		p.AddCard(canastaCard(domain.CardDesignSpade, 7))
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1}}), domain.ErrInvalidPlay, "canasta.errMeldNeedsAtLeastThreeCards")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		p.AddCard(canastaCard(domain.CardDesignSpade, 7))
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		p.AddCard(canastaCard(domain.CardDesignDiamond, 7))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}, {0, 1, 2}}), domain.ErrInvalidCard, "canasta.errDuplicateCardIndex")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		p.AddCard(canastaCard(domain.CardDesignSpade, 7))
		p.AddCard(canastaCard(domain.CardDesignHeart, 8))
		p.AddCard(canastaCard(domain.CardDesignDiamond, 7))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "canasta.errMeldCardsMustHaveSameRank")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		p.AddCard(canastaCard(domain.CardDesignSpade, 3))
		p.AddCard(canastaCard(domain.CardDesignClover, 3))
		p.AddCard(canastaCard(domain.CardDesignHeart, 3))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "canasta.errBlackThreeCannotMeld")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		for i := 0; i < 3; i++ {
			p.AddCard(canastaCard(domain.CardDesignJoker, i+1))
		}
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "canasta.errMeldNeedsAtLeastTwoNaturalCards")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		for _, d := range []int{domain.CardDesignSpade, domain.CardDesignHeart} {
			p.AddCard(canastaCard(d, 7))
		}
		for i := 0; i < 4; i++ {
			p.AddCard(canastaCard(domain.CardDesignJoker, i+1))
		}
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2, 3, 4, 5}}), domain.ErrInvalidPlay, "canasta.errMeldAllowsAtMostThreeWildCards")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		p.AddCard(canastaCard(domain.CardDesignSpade, 7))
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		p.AddCard(canastaCard(domain.CardDesignJoker, 1))
		p.AddCard(canastaCard(domain.CardDesignJoker, 2))
		p.AddCard(canastaCard(domain.CardDesignSpade, 2))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2, 3, 4}}), domain.ErrInvalidPlay, "canasta.errWildCardsCannotExceedNaturalCards")

		g, p = canastaErrorGame(domain.CanastaPhaseMeld)
		p.SetHasInitMeld(true)
		p.SetMelds([]*domain.CanastaMeld{{Cards: []*domain.Card{canastaCard(domain.CardDesignSpade, 7), canastaCard(domain.CardDesignHeart, 7)}, IsNatural: true}})
		p.AddCard(canastaCard(domain.CardDesignDiamond, 7))
		p.AddCard(canastaCard(domain.CardDesignClover, 8))
		p.AddCard(canastaCard(domain.CardDesignHeart, 2))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "canasta.errMeldCardRankMismatch")

		g, p = canastaErrorGame(domain.CanastaPhaseDiscard)
		p.AddCard(canastaCard(domain.CardDesignHeart, 3))
		assertCanastaDomainError(t, g.PlayerDiscard(0), domain.ErrInvalidPlay, "canasta.errRedThreeCannotBeDiscarded")

		g, p = canastaErrorGame(domain.CanastaPhaseDiscard)
		p.SetMelds([]*domain.CanastaMeld{{Cards: make([]*domain.Card, 7), IsNatural: true}})
		p.SetHasInitMeld(true)
		p.AddCard(canastaCard(domain.CardDesignClover, 5))
		p.AddCard(canastaCard(domain.CardDesignClover, 6))
		assertCanastaDomainError(t, g.PlayerGoOut(), domain.ErrInvalidPlay, "canasta.errHandMustHaveAtMostOneCardToGoOut")

		g, _ = canastaErrorGame(domain.CanastaPhaseDiscard)
		cfg := g.GetConfig()
		cfg.UsePozzetto = true
		g.SetConfig(cfg)
		assertCanastaDomainError(t, g.PlayerGoOut(), domain.ErrInvalidPlay, "canasta.errPozzettoRequiredToGoOut")

		g, _ = canastaErrorGame(domain.CanastaPhaseDiscard)
		assertCanastaDomainError(t, g.PlayerGoOut(), domain.ErrInvalidPlay, "canasta.errCanastaRequiredToGoOut")
	})

	t.Run("discard top must be melded", func(t *testing.T) {
		g, p := canastaErrorGame(domain.CanastaPhaseDraw)
		g.SetDiscardPile([]*domain.Card{canastaCard(domain.CardDesignSpade, 7)})
		p.AddCard(canastaCard(domain.CardDesignHeart, 7))
		p.AddCard(canastaCard(domain.CardDesignDiamond, 7))
		p.SetHasInitMeld(true)
		require.NoError(t, g.PlayerDrawFromDiscard([]int{0, 1}))
		assertCanastaDomainError(t, g.PlayerSkipMeld(), domain.ErrInvalidPlay, "canasta.errTopCardMustBeMelded")
	})
}

func TestCanasta_BiribaValidationErrorsHaveCodes(t *testing.T) {
	biribaGame := func() (*domain.Canasta, *domain.CanastaPlayer) {
		g, p := canastaErrorGame(domain.CanastaPhaseMeld)
		cfg := g.GetConfig()
		cfg.UseBiriba = true
		g.SetConfig(cfg)
		p.SetHasInitMeld(true)
		return g, p
	}

	t.Run("same suit", func(t *testing.T) {
		g, p := biribaGame()
		p.AddCard(canastaCard(domain.CardDesignSpade, 4))
		p.AddCard(canastaCard(domain.CardDesignHeart, 5))
		p.AddCard(canastaCard(domain.CardDesignSpade, 6))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "canasta.errSequenceMustUseSameSuit")
	})
	t.Run("duplicate rank", func(t *testing.T) {
		g, p := biribaGame()
		p.AddCard(canastaCard(domain.CardDesignSpade, 4))
		p.AddCard(canastaCard(domain.CardDesignSpade, 4))
		p.AddCard(canastaCard(domain.CardDesignSpade, 5))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "canasta.errSequenceCannotDuplicateRank")
	})
	t.Run("not consecutive", func(t *testing.T) {
		g, p := biribaGame()
		p.AddCard(canastaCard(domain.CardDesignSpade, 4))
		p.AddCard(canastaCard(domain.CardDesignSpade, 6))
		p.AddCard(canastaCard(domain.CardDesignSpade, 7))
		assertCanastaDomainError(t, g.PlayerMeld([][]int{{0, 1, 2}}), domain.ErrInvalidPlay, "canasta.errSequenceRanksNotConsecutive")
	})
	t.Run("wild cards exceed sequence range", func(t *testing.T) {
		g, p := biribaGame()
		for _, rank := range []int{1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12} {
			p.AddCard(canastaCard(domain.CardDesignHeart, rank))
		}
		for i := 0; i < 3; i++ {
			p.AddCard(canastaCard(domain.CardDesignJoker, i+1))
		}
		indices := make([]int, 14)
		for i := range indices {
			indices[i] = i
		}
		assertCanastaDomainError(t, g.PlayerMeld([][]int{indices}), domain.ErrInvalidPlay, "canasta.errWildCardsExceedSequenceRange")
	})
}
