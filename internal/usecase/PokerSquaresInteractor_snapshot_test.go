//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

type stubPokerSquaresPresenter struct{}

func (s *stubPokerSquaresPresenter) Output(_ interfaces.PokerSquaresGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubPokerSquaresPresenter) ActionLogOutput(_ interfaces.PokerSquaresGame) string {
	return `{"log":[]}`
}

func (s *stubPokerSquaresPresenter) HintOutput(_ interfaces.PokerSquaresGame) string {
	return `{"hint":true}`
}

func TestPokerSquaresInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewPokerSquares(domain.NewTrumpCards(0))
	pi := NewPokerSquaresInteractor(g, new(stubPokerSquaresPresenter))

	pi.Reset()
	pi.Place(0, 0)

	data, err := pi.Snapshot()
	require.NoError(t, err)

	restored, err := RestorePokerSquaresInteractor(data, new(stubPokerSquaresPresenter))
	require.NoError(t, err)

	result := restored.Place(0, 1)
	assert.NotEmpty(t, result)
}

func TestPokerSquaresInteractor_RestoreInvalidJSON(t *testing.T) {
	_, err := RestorePokerSquaresInteractor([]byte("not json"), new(stubPokerSquaresPresenter))
	assert.Error(t, err)
}
