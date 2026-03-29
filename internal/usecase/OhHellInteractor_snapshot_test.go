//go:build test

package usecase

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubOhHellPresenter is a minimal presenter for snapshot tests.
type stubOhHellPresenter struct{}

func (s *stubOhHellPresenter) Output(_ interfaces.OhHellGame, _ error) string {
	return `{}`
}
func (s *stubOhHellPresenter) ActionLogOutput(_ interfaces.OhHellGame) string { return `{}` }
func (s *stubOhHellPresenter) HintOutput(_ interfaces.OhHellGame) string      { return `{}` }

func TestOhHellInteractor_SnapshotRestore(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
	o := domain.NewOhHell(tc, players, domain.DefaultOhHellConfig())
	oi := NewOhHellInteractor(o, new(stubOhHellPresenter))

	data, err := oi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreOhHellInteractor(data, new(stubOhHellPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}

func TestOhHellInteractor_SnapshotRestore_BidPhase(t *testing.T) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
	o := domain.NewOhHell(tc, players, domain.DefaultOhHellConfig())
	oi := NewOhHellInteractor(o, new(stubOhHellPresenter))
	_ = oi.Reset()

	data, err := oi.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreOhHellInteractor(data, new(stubOhHellPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)
}
