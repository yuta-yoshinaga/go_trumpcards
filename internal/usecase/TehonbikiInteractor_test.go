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

type tehonbikiRecordingPresenter struct {
	outputs int
	hints   int
	logs    int
	lastErr error
}

func (p *tehonbikiRecordingPresenter) Output(_ interfaces.TehonbikiGame, err error) string {
	p.outputs++
	p.lastErr = err
	return "output"
}
func (p *tehonbikiRecordingPresenter) ActionLogOutput(interfaces.TehonbikiGame) string {
	p.logs++
	return "action-log"
}
func (p *tehonbikiRecordingPresenter) HintOutput(interfaces.TehonbikiGame) string {
	p.hints++
	return "hint"
}

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

func TestTehonbikiInteractorCallsPresenterForAllMethods(t *testing.T) {
	g := domain.NewDefaultTehonbiki()
	_ = g.SetParentCard(2)
	p := new(tehonbikiRecordingPresenter)
	ci := NewTehonbikiInteractor(g, p)

	if got := ci.Reset(); got != "output" || p.outputs != 1 {
		t.Fatalf("reset output=%q calls=%d", got, p.outputs)
	}
	if got := ci.GetConfig(); got != domain.DefaultTehonbikiConfig() {
		t.Fatalf("config=%+v", got)
	}
	if got := ci.Hint(); got != "hint" || p.hints != 1 {
		t.Fatalf("hint=%q calls=%d", got, p.hints)
	}
	if got := ci.ActionLog(); got != "action-log" || p.logs != 1 {
		t.Fatalf("log=%q calls=%d", got, p.logs)
	}
	if got := ci.PlaceBet([]int{1}, domain.TehonbikiBetSingle, 50); got != "output" || p.outputs != 2 || p.lastErr != nil {
		t.Fatalf("bet output=%q calls=%d err=%v", got, p.outputs, p.lastErr)
	}
	if got := ci.NextRound(); got != "output" || p.outputs != 3 || p.lastErr != nil {
		t.Fatalf("next output=%q calls=%d err=%v", got, p.outputs, p.lastErr)
	}
	if got := ci.ResetWithConfig(domain.TehonbikiConfig{InitialChips: 500, DefaultBet: 20}); got != "output" || p.outputs != 4 || g.GetConfig().DefaultBet != 20 {
		t.Fatalf("reset config output=%q calls=%d config=%+v", got, p.outputs, g.GetConfig())
	}
	if got := ci.PlaceBet([]int{1, 2}, domain.TehonbikiBetSingle, 50); got != "output" || p.lastErr == nil {
		t.Fatalf("invalid bet did not reach presenter: output=%q err=%v", got, p.lastErr)
	}
	if got := ci.ResetWithConfig(domain.TehonbikiConfig{InitialChips: 1, DefaultBet: 20}); got != "output" || p.lastErr == nil {
		t.Fatalf("invalid config did not reach presenter: output=%q err=%v", got, p.lastErr)
	}
}

func TestTehonbikiInteractorRestoreRejectsInvalidJSON(t *testing.T) {
	if _, err := RestoreTehonbikiInteractor([]byte("not json"), new(tehonbikiSilentPresenter)); err == nil {
		t.Fatal("invalid snapshot accepted")
	}
}
