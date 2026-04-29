//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubTexasHoldemBonusPresenter is a minimal presenter for snapshot tests.
type stubTexasHoldemBonusPresenter struct{}

func (s *stubTexasHoldemBonusPresenter) Output(_ interfaces.TexasHoldemBonusGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubTexasHoldemBonusPresenter) ActionLogOutput(_ interfaces.TexasHoldemBonusGame) string {
	return `{"log":[]}`
}

func TestTexasHoldemBonusInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	ti := NewTexasHoldemBonusInteractor(g, new(stubTexasHoldemBonusPresenter))

	ti.Bet(100, 10)

	data, err := ti.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreTexasHoldemBonusInteractor(data, new(stubTexasHoldemBonusPresenter))
	require.NoError(t, err)

	result := restored.Reset()
	assert.NotEmpty(t, result)
}

func TestTexasHoldemBonusInteractor_SnapshotRestore_BetPhase(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	ti := NewTexasHoldemBonusInteractor(g, new(stubTexasHoldemBonusPresenter))

	data, err := ti.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreTexasHoldemBonusInteractor(data, new(stubTexasHoldemBonusPresenter))
	require.NoError(t, err)

	result := restored.Bet(100, 0)
	assert.NotEmpty(t, result)
}
