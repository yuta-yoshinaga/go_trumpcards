//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertTeenPattiCodedError(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestTeenPattiDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("already seen", func(t *testing.T) {
		g := newTestTeenPatti(true)
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		require.NoError(t, g.PlayerSee())
		assertTeenPattiCodedError(t, g.PlayerSee(), "teenpatti.errAlreadySeen")
	})

	t.Run("raise must exceed stake", func(t *testing.T) {
		g := newTestTeenPatti(true)
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		assertTeenPattiCodedError(t, g.PlayerRaise(g.GetStake()), "teenpatti.errRaiseMustExceedStake")
	})

	t.Run("show requires two seen players", func(t *testing.T) {
		g := newTestTeenPatti(true)
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		assertTeenPattiCodedError(t, g.PlayerShow(), "teenpatti.errShowRequiresTwoSeen")
	})

	t.Run("side show requires seen players", func(t *testing.T) {
		g := newTestTeenPatti(true)
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		assertTeenPattiCodedError(t, g.PlayerRequestSideShow(), "teenpatti.errSideShowRequiresSeen")
	})

	t.Run("insufficient chips", func(t *testing.T) {
		g := newTestTeenPatti(true)
		g.Reset()
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).SetChips(0)
		assertTeenPattiCodedError(t, g.PlayerRaise(g.GetStake()+1), "teenpatti.errInsufficientChips")
	})
}
