//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubOasisPokerPresenter is a minimal presenter for snapshot tests.
type stubOasisPokerPresenter struct{}

func (s *stubOasisPokerPresenter) Output(_ interfaces.OasisPokerGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubOasisPokerPresenter) ActionLogOutput(_ interfaces.OasisPokerGame) string {
	return `{"log":[]}`
}

func (s *stubOasisPokerPresenter) HintOutput(_ interfaces.OasisPokerGame) string {
	return `{"ok":true}`
}

func TestOasisPokerInteractor_SnapshotRestore(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	oi := NewOasisPokerInteractor(op, new(stubOasisPokerPresenter))

	oi.Bet(100, 10)

	data, err := oi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreOasisPokerInteractor(data, new(stubOasisPokerPresenter))
	require.NoError(t, err)

	result := restored.Reset()
	assert.NotEmpty(t, result)
}

func TestOasisPokerInteractor_SnapshotRestore_BetPhase(t *testing.T) {
	op := domain.NewDefaultOasisPoker()
	oi := NewOasisPokerInteractor(op, new(stubOasisPokerPresenter))

	data, err := oi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreOasisPokerInteractor(data, new(stubOasisPokerPresenter))
	require.NoError(t, err)

	result := restored.Bet(100, 0)
	assert.NotEmpty(t, result)
}
