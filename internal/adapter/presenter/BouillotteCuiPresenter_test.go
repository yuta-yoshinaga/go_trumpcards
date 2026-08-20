package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBouillotteCuiPresenter_OutputBettingPhase(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	// Title, round line and the shared retourne card should be present.
	assert.Contains(t, out, "ブイヨット")
	assert.Contains(t, out, "ラウンド")
	assert.Contains(t, out, "ルトゥルヌ")
}

func TestBouillotteCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestBouillotteCuiPresenter_OutputResult(t *testing.T) {
	g := bouillotteResultGame(false) // human loses to a CPU brelan (shared helper)
	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "ラウンド")
}

// TestBouillotteCuiPresenter_OutputResultCpuWin drives the cpuWin (result==None)
// branch: the human folds, so a CPU takes the pot. The winner must be shown by
// localized player name, not a raw index.
func TestBouillotteCuiPresenter_OutputResultCpuWin(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	g.SetPhase(domain.BouillottePhaseBetting)
	g.SetRetourne(domain.NewCard(domain.CardDesignDiamond, 8, false))
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetOut(false)
		g.GetPlayer(i).SetFolded(false)
	}
	// Human (seat 0) folds; CPU seat 1 holds the winning brelan.
	g.GetPlayer(0).SetFolded(true)
	bouillotteWebSetHand(g.GetPlayer(1),
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 1, false),
		domain.NewCard(domain.CardDesignHeart, 1, false))
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(true)
	}
	g.SetPot(100)
	g.ResolveForTest()

	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, nil)
	assert.Contains(t, out, "がポットを獲得")
	// Winner named via cuiPlayerName (CPU N), not a raw index.
	assert.Contains(t, out, "CPU")
}

func TestBouillotteCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.BouillottePhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	require.True(t, g.GetGameEndFlag())
	p := new(presenter.BouillotteCuiPresenter)
	out := p.Output(g, nil)
	assert.Contains(t, out, "の勝ち")
	// The winner is shown via cuiPlayerName (あなた / CPU N), not a raw index.
	assert.True(t, strings.Contains(out, "あなた") || strings.Contains(out, "CPU"),
		"game-end banner should name the winner, got: %s", out)
}

func TestBouillotteCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestBouillotteCuiPresenter_HintOutputNone(t *testing.T) {
	g := bouillotteResultGame(false) // result phase → GetHint returns nil
	p := new(presenter.BouillotteCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestBouillotteCuiPresenter_ActionLog(t *testing.T) {
	g := bouillotteResultGame(false)
	p := new(presenter.BouillotteCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// #5698: 同型の Primero は「コールに必要な額」と「レイズ後の到達額」を出すのに、
// Bouillotte は現在のベット額しか出さず、必要額を暗算させていた。
func TestBouillotteCuiPresenter_ShowsCallAndRaiseAmounts(t *testing.T) {
	p := new(presenter.BouillotteCuiPresenter)

	t.Run("spells out the call and the raise total", func(t *testing.T) {
		g := domain.NewDefaultBouillotte()
		g.Reset()
		g.SetPhase(domain.BouillottePhaseBetting)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentBet(12)
		human := g.GetPlayer(0)
		human.AddRoundBet(5 - human.GetRoundBet()) // 既に 5 払っている状態にする

		out := p.Output(g, nil)

		// need = 12 - 5、raiseTo = 12 + アンティ。
		assert.Contains(t, out, i18n.Tf("bouillotte.promptBetting",
			"bet", "12",
			"need", "7",
			"raiseTo", strconv.Itoa(12+g.GetAnte())))
		assert.NotContains(t, out, "{{", "no placeholder may leak into the prompt")
	})

	// 自分の払いが現在のベットに追いついている席では、追加は 0 (負にしない)。
	t.Run("never asks for a negative call", func(t *testing.T) {
		g := domain.NewDefaultBouillotte()
		g.Reset()
		g.SetPhase(domain.BouillottePhaseBetting)
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentBet(4)
		human := g.GetPlayer(0)
		human.AddRoundBet(9 - human.GetRoundBet())

		out := p.Output(g, nil)

		assert.Contains(t, out, i18n.Tf("bouillotte.promptBetting",
			"bet", "4",
			"need", "0",
			"raiseTo", strconv.Itoa(4+g.GetAnte())))
	})
}
