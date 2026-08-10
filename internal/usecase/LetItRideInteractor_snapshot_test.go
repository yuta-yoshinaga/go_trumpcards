//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubLetItRidePresenter is a minimal presenter for snapshot tests.
type stubLetItRidePresenter struct{}

func (s *stubLetItRidePresenter) Output(_ interfaces.LetItRideGame, _ error) string {
	return `{"ok":true}`
}

func (s *stubLetItRidePresenter) ActionLogOutput(_ interfaces.LetItRideGame) string {
	return `{"log":[]}`
}

func (s *stubLetItRidePresenter) PullConfirmOutput(_ interfaces.LetItRideGame) string {
	return `{}`
}

func TestLetItRideInteractor_SnapshotRestore(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	li := NewLetItRideInteractor(lir, new(stubLetItRidePresenter))

	li.Bet(100)

	data, err := li.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreLetItRideInteractor(data, new(stubLetItRidePresenter))
	require.NoError(t, err)

	result := restored.Reset()
	assert.NotEmpty(t, result)
}

func TestLetItRideInteractor_SnapshotRestore_BetPhase(t *testing.T) {
	lir := domain.NewDefaultLetItRide()
	li := NewLetItRideInteractor(lir, new(stubLetItRidePresenter))

	data, err := li.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreLetItRideInteractor(data, new(stubLetItRidePresenter))
	require.NoError(t, err)

	result := restored.Bet(100)
	assert.NotEmpty(t, result)
}
