//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertHeartsDomainError(t *testing.T, err error, sentinel error, code string, params map[string]string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	actualCode, actualParams := domain.ErrorMessageCode(err)
	assert.Equal(t, code, actualCode)
	assert.Equal(t, params, actualParams)
}

func newHeartsErrorGame() *domain.Hearts {
	players := []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(true),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
	}
	g := domain.NewHearts(domain.NewTrumpCards(0), players, domain.DefaultHeartsConfig())
	g.SetPhase(domain.HeartsPhasePass)
	g.SetCurrentPlayerIdx(0)
	return g
}

func TestHeartsDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		g := newHeartsErrorGame()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		assertHeartsDomainError(t, g.PlayerPass([]int{0}), domain.ErrInvalidPlay, "hearts.errPassCardCount", map[string]string{"count": "3"})
		assert.NoError(t, g.PlayerPass([]int{0, 1, 2}))
		assertHeartsDomainError(t, g.PlayerPass([]int{0, 1, 2}), domain.ErrInvalidPlay, "hearts.errPassAlreadySelected", nil)
	})

	t.Run("duplicate pass card", func(t *testing.T) {
		g := newHeartsErrorGame()
		for i := 0; i < 3; i++ {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2+i, false))
		}
		assertHeartsDomainError(t, g.PlayerPass([]int{0, 0, 1}), domain.ErrInvalidCard, "hearts.errDuplicateCardIndex", nil)
	})

	t.Run("pass card index", func(t *testing.T) {
		g := newHeartsErrorGame()
		for i := 0; i < 3; i++ {
			g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2+i, false))
		}
		assertHeartsDomainError(t, g.PlayerPass([]int{0, 1, 3}), domain.ErrInvalidCard, "hearts.errCardIndexOutOfRange", nil)
	})

	t.Run("play card index", func(t *testing.T) {
		g := newHeartsErrorGame()
		g.SetPhase(domain.HeartsPhasePlay)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assertHeartsDomainError(t, g.PlayerPlay(1), domain.ErrInvalidCard, "hearts.errCardIndexOutOfRange", nil)
	})

	t.Run("first trick must lead two of clubs", func(t *testing.T) {
		g := newHeartsErrorGame()
		g.SetPhase(domain.HeartsPhasePlay)
		g.SetTrickNumber(1)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assertHeartsDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "hearts.errFirstTrickMustLeadTwoOfClubs", nil)
	})

	t.Run("hearts not broken", func(t *testing.T) {
		g := newHeartsErrorGame()
		g.SetPhase(domain.HeartsPhasePlay)
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assertHeartsDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "hearts.errHeartsNotBroken", nil)
	})

	t.Run("follow lead suit", func(t *testing.T) {
		g := newHeartsErrorGame()
		g.SetPhase(domain.HeartsPhasePlay)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 3, false)}})
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assertHeartsDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "hearts.errFollowLeadSuit", nil)
	})

	t.Run("first trick point card", func(t *testing.T) {
		g := newHeartsErrorGame()
		g.SetPhase(domain.HeartsPhasePlay)
		g.SetTrickNumber(1)
		g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 3, false)}})
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		assertHeartsDomainError(t, g.PlayerPlay(0), domain.ErrInvalidPlay, "hearts.errFirstTrickNoPointCards", nil)
	})
}
