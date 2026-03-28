//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubBaccaratPresenter is a minimal presenter for snapshot tests.
type stubBaccaratPresenter struct{}

func (s *stubBaccaratPresenter) Output(_ interfaces.BaccaratGame, _ error) string {
	return `{"ok":true}`
}
func (s *stubBaccaratPresenter) ActionLogOutput(_ interfaces.BaccaratGame) string {
	return `{"log":[]}`
}

func TestBaccaratInteractor_SnapshotRestore(t *testing.T) {
	bac := domain.NewDefaultBaccarat()
	bi := NewBaccaratInteractor(bac, new(stubBaccaratPresenter))

	// Play a game to populate state
	bi.Bet(100, domain.BaccaratBetPlayer, 0, 0)

	data, err := bi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreBaccaratInteractor(data, new(stubBaccaratPresenter))
	require.NoError(t, err)

	// The restored interactor should produce valid Reset output.
	result := restored.Reset()
	assert.NotEmpty(t, result)
}

func TestBaccaratInteractor_SnapshotRestore_BetPhase(t *testing.T) {
	bac := domain.NewDefaultBaccarat()
	bi := NewBaccaratInteractor(bac, new(stubBaccaratPresenter))

	// Snapshot in initial bet phase
	data, err := bi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreBaccaratInteractor(data, new(stubBaccaratPresenter))
	require.NoError(t, err)

	// The restored interactor should accept a bet
	result := restored.Bet(100, domain.BaccaratBetBanker, 0, 0)
	assert.NotEmpty(t, result)
}
