//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertThreeCardBragCodedError(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestThreeCardBragDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("already seen", func(t *testing.T) {
		g := newTestBrag(true)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).SetSeen(true)
		assertThreeCardBragCodedError(t, g.PlayerSee(), "threecardbrag.errAlreadySeen")
	})

	t.Run("raise must exceed stake", func(t *testing.T) {
		g := newTestBrag(true)
		g.SetCurrentPlayerIdx(0)
		g.SetStake(5)
		assertThreeCardBragCodedError(t, g.PlayerRaise(5), "threecardbrag.errRaiseMustExceedStake")
	})

	t.Run("show unavailable", func(t *testing.T) {
		g := newTestBrag(true)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).SetSeen(true)
		assertThreeCardBragCodedError(t, g.PlayerShow(), "threecardbrag.errShowUnavailable")
	})

	t.Run("insufficient chips", func(t *testing.T) {
		g := newTestBrag(true)
		g.SetCurrentPlayerIdx(0)
		g.GetPlayer(0).SetChips(0)
		assertThreeCardBragCodedError(t, g.PlayerRaise(g.GetStake()+1), "threecardbrag.errInsufficientChips")
	})
}
