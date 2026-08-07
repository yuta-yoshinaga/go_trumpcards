//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubMississippiStudPresenter は snapshot/restore テスト用の最小プレゼンター。
type stubMississippiStudPresenter struct{}

func (s *stubMississippiStudPresenter) Output(_ interfaces.MississippiStudGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubMississippiStudPresenter) ActionLogOutput(_ interfaces.MississippiStudGame) string {
	return `{"log":[]}`
}

func (s *stubMississippiStudPresenter) HintOutput(_ interfaces.MississippiStudGame) string {
	return `{}`
}

func TestMississippiStudInteractor_SnapshotRestore(t *testing.T) {
	g := domain.NewDefaultMississippiStud()
	mi := NewMississippiStudInteractor(g, new(stubMississippiStudPresenter))

	mi.Bet(100)
	mi.Play(2)

	data, err := mi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreMississippiStudInteractor(data, new(stubMississippiStudPresenter))
	require.NoError(t, err)

	result := restored.Reset()
	assert.NotEmpty(t, result)
}

func TestMississippiStudInteractor_SnapshotRestore_AntePhase(t *testing.T) {
	g := domain.NewDefaultMississippiStud()
	mi := NewMississippiStudInteractor(g, new(stubMississippiStudPresenter))

	data, err := mi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreMississippiStudInteractor(data, new(stubMississippiStudPresenter))
	require.NoError(t, err)

	result := restored.Bet(100)
	assert.NotEmpty(t, result)
}
