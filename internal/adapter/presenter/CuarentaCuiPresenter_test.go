package presenter_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func buildPlayedCuarenta(t *testing.T) *domain.Cuarenta {
	t.Helper()
	g := domain.NewDefaultCuarenta()
	g.Reset()
	return g
}

func TestCuarentaCuiPresenter_Output(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	g := buildPlayedCuarenta(t)
	out := p.Output(g, nil)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	// i18n (ja) is loaded in this build; assert on rendered Japanese text.
	if !strings.Contains(out, "チーム") {
		t.Errorf("expected localized team text in output, got: %s", out)
	}
}

func TestCuarentaCuiPresenter_Error(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	g := buildPlayedCuarenta(t)
	out := p.Output(g, cuarentaAssertErr{})
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error text in output, got: %s", out)
	}
}

type cuarentaAssertErr struct{}

func (cuarentaAssertErr) Error() string { return "boom" }

func TestCuarentaCuiPresenter_GameEnd(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	cfg := domain.DefaultCuarentaConfig()
	cfg.TargetScore = 6
	players := make([]*domain.CuarentaPlayer, domain.CuarentaPlayerCnt)
	for i := 0; i < domain.CuarentaPlayerCnt; i++ {
		players[i] = domain.NewCuarentaPlayer(false)
	}
	g := domain.NewCuarenta(domain.NewTrumpCardsScopa(), players, cfg)
	g.Reset()
	// run to completion to exercise the game-end branch.
	for i := 0; i < 100000 && !g.GetGameEndFlag(); i++ {
		g.CpuPlay()
		if !g.GetGameEndFlag() && g.GetPhase() != int(domain.CuarentaPhasePlay) {
			g.NextRound()
		}
	}
	out := p.Output(g, nil)
	if !strings.Contains(out, "ゲーム終了") {
		t.Errorf("expected game-end text, got: %s", out)
	}
}

func TestCuarentaCuiPresenter_ActionLogOutput(t *testing.T) {
	p := &presenter.CuarentaCuiPresenter{}
	g := buildPlayedCuarenta(t)
	if out := p.ActionLogOutput(g); out == "" {
		t.Error("expected non-empty action log output")
	}
}
