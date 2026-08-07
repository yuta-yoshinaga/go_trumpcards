//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubCaribbeanStudPresenter is a minimal presenter for snapshot tests.
type stubCaribbeanStudPresenter struct{}

func (s *stubCaribbeanStudPresenter) Output(_ interfaces.CaribbeanStudGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubCaribbeanStudPresenter) ActionLogOutput(_ interfaces.CaribbeanStudGame) string {
	return `{"log":[]}`
}

func (s *stubCaribbeanStudPresenter) HintOutput(_ interfaces.CaribbeanStudGame) string {
	return `{}`
}

func TestCaribbeanStudInteractor_SnapshotRestore(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	ci := NewCaribbeanStudInteractor(cs, new(stubCaribbeanStudPresenter))

	ci.Bet(100, 10)

	data, err := ci.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreCaribbeanStudInteractor(data, new(stubCaribbeanStudPresenter))
	require.NoError(t, err)

	result := restored.Reset()
	assert.NotEmpty(t, result)
}

func TestCaribbeanStudInteractor_SnapshotRestore_BetPhase(t *testing.T) {
	cs := domain.NewDefaultCaribbeanStud()
	ci := NewCaribbeanStudInteractor(cs, new(stubCaribbeanStudPresenter))

	data, err := ci.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreCaribbeanStudInteractor(data, new(stubCaribbeanStudPresenter))
	require.NoError(t, err)

	result := restored.Bet(100, 0)
	assert.NotEmpty(t, result)
}
