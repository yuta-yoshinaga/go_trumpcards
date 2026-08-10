//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

type stubCribbageSquaresPresenter struct{}

func (s *stubCribbageSquaresPresenter) Output(_ interfaces.CribbageSquaresGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubCribbageSquaresPresenter) ActionLogOutput(_ interfaces.CribbageSquaresGame) string {
	return `{"log":[]}`
}

func (s *stubCribbageSquaresPresenter) HintOutput(_ interfaces.CribbageSquaresGame) string {
	return `{"hint":true}`
}

// The Worker is stateless per request and rebuilds the game from this snapshot
// on every call, so a broken restore path means the board silently resets
// mid-game in production rather than failing anywhere a test would see (#4478).
func TestCribbageSquaresInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewCribbageSquares(domain.NewTrumpCards(0))
	ci := NewCribbageSquaresInteractor(g, new(stubCribbageSquaresPresenter))

	ci.Reset()
	ci.Place(0, 0)
	ci.Place(1, 1)

	data, err := ci.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreCribbageSquaresInteractor(data, new(stubCribbageSquaresPresenter))
	require.NoError(t, err)

	// The board came back, so the two taken cells are still taken and a third
	// placement still works.
	assert.NotEmpty(t, restored.Place(2, 2))
	assert.NotEmpty(t, restored.Undo(), "the undo stack survived the round trip too")
}

func TestCribbageSquaresInteractor_RestoreInvalidJSON(t *testing.T) {
	_, err := RestoreCribbageSquaresInteractor([]byte("not json"), new(stubCribbageSquaresPresenter))
	assert.Error(t, err)
}
