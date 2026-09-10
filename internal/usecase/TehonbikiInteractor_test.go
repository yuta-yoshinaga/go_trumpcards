//go:build test

package usecase

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

type tehonbikiSilentPresenter struct{}

func (*tehonbikiSilentPresenter) Output(interfaces.TehonbikiGame, error) string   { return "ok" }
func (*tehonbikiSilentPresenter) ActionLogOutput(interfaces.TehonbikiGame) string { return "log" }
func (*tehonbikiSilentPresenter) HintOutput(interfaces.TehonbikiGame) string      { return "hint" }

func TestTehonbikiInteractorDelegatesWagerAndSnapshot(t *testing.T) {
	g := domain.NewDefaultTehonbiki()
	require.NoError(t, g.SetParentCard(4))
	ci := NewTehonbikiInteractor(g, new(tehonbikiSilentPresenter))
	require.Equal(t, "ok", ci.PlaceBet([]int{2, 4}, domain.TehonbikiBetDouble, 50))
	require.Equal(t, domain.TehonbikiResultWin, g.GetResult())
	b, err := ci.Snapshot()
	require.NoError(t, err)
	require.True(t, json.Valid(b))
	restored, err := RestoreTehonbikiInteractor(b, new(tehonbikiSilentPresenter))
	require.NoError(t, err)
	require.Equal(t, g.GetResult(), restored.Game.GetResult())
}

func TestTehonbikiInteractorDoesNotLetGameEndThrough(t *testing.T) {
	g := domain.NewDefaultTehonbikiWithPlayer(domain.NewTehonbikiPlayer(50), domain.DefaultTehonbikiConfig())
	require.NoError(t, g.SetParentCard(6))
	ci := NewTehonbikiInteractor(g, new(tehonbikiSilentPresenter))
	require.Equal(t, "ok", ci.PlaceBet([]int{1}, domain.TehonbikiBetSingle, 50))
	require.Equal(t, "ok", ci.NextRound())
	require.True(t, g.GetGameEndFlag())
	before := g.GetChips()
	ci.PlaceBet([]int{1}, domain.TehonbikiBetSingle, 50)
	require.Equal(t, before, g.GetChips())
}
