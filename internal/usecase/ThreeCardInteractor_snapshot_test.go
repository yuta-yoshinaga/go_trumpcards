//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubThreeCardPresenter is a minimal presenter for snapshot tests.
type stubThreeCardPresenter struct{}

func (s *stubThreeCardPresenter) Output(_ interfaces.ThreeCardGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubThreeCardPresenter) ActionLogOutput(_ interfaces.ThreeCardGame) string {
	return `{"log":[]}`
}

func (s *stubThreeCardPresenter) HintOutput(_ interfaces.ThreeCardGame) string {
	return `{"hint":true}`
}

func TestThreeCardInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	ti := NewThreeCardInteractor(tc, new(stubThreeCardPresenter))

	// Play a game to populate state
	ti.Bet(100, 50)

	data, err := ti.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreThreeCardInteractor(data, new(stubThreeCardPresenter))
	require.NoError(t, err)

	// The restored interactor should produce valid Reset output.
	result := restored.Reset()
	assert.NotEmpty(t, result)
}

func TestThreeCardInteractor_SnapshotRestore_BetPhase(t *testing.T) {
	tc := domain.NewDefaultThreeCard()
	ti := NewThreeCardInteractor(tc, new(stubThreeCardPresenter))

	// Snapshot in initial bet phase
	data, err := ti.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreThreeCardInteractor(data, new(stubThreeCardPresenter))
	require.NoError(t, err)

	// The restored interactor should accept a bet
	result := restored.Bet(100, 0)
	assert.NotEmpty(t, result)
}
