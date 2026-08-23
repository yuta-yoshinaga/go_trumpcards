package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupMadrassoCuiMock() *interfaces.MockMadrassoGame {
	m := new(interfaces.MockMadrassoGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTeamScores").Return([domain.MadrassoTeamCnt]int{0, 0})
	m.On("GetTeamRoundPoints").Return([domain.MadrassoTeamCnt]int{0, 0})
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MadrassoPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// 合法手の目印 (#5633)。既定は「制限なし」= 印を出さない状態。
	m.On("GetPlayableIndices", mock.Anything).Return([]int(nil)).Maybe()
	return m
}

func setupMadrassoCuiMockWithPlayers() (*interfaces.MockMadrassoGame, []*domain.MadrassoPlayer) {
	m := setupMadrassoCuiMock()
	players := makeMadrassoPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMadrassoCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MadrassoCuiPresenter)

	t.Run("play phase shows player info and prompt", func(t *testing.T) {
		m, players := setupMadrassoCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Madrasso")
		assert.NotEmpty(t, result)
		// The thirds-conversion rule is explained alongside the score line.
		rule := i18n.T("madrasso.thirdsRule")
		require.NotEqual(t, "madrasso.thirdsRule", rule, "thirdsRule が未翻訳")
		assert.Contains(t, result, rule)
		// **切り札は配りで決まる。** クローン元 (トレセッテ) に無い概念なので、
		// 行ごと落ちても他のテストは気づかない。
		assert.Contains(t, result, i18n.Tf("madrasso.trumpLine", "suit", "SPADE"))
		// 1/3 点の言い回しが残っていないこと。
		assert.NotContains(t, result, "/3)")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupMadrassoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MadrassoPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows team breakdown and last-trick team", func(t *testing.T) {
		m, _ := setupMadrassoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamRoundPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
		m.On("GetPhase").Return(domain.MadrassoPhaseRoundEnd)
		m.On("GetTeamRoundPoints").Return([domain.MadrassoTeamCnt]int{7, 4})
		m.On("GetLeadPlayerIdx").Return(1) // team B took the last trick
		result := p.Output(m, nil)
		assert.Contains(t, result, strings.Split(i18n.T("madrasso.roundBreakdown"), "{{")[0])
		// **整数点の内訳。** クローン元は "thirds7" のような 1/3 点表記だった。
		assert.Contains(t, result, i18n.Tf("madrasso.roundBreakdown",
			"a", "A", "athird", "7", "b", "B", "bthird", "4", "lastteam", "B"))
		// 1/3 点の言い回しが残っていないこと。
		assert.NotContains(t, result, "thirds")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupMadrassoCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMadrassoCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestMadrassoCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MadrassoCuiPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, players := setupMadrassoCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetHint").Return(&domain.MadrassoHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMadrassoCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.MadrassoHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestMadrassoCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MadrassoCuiPresenter)
	m := new(interfaces.MockMadrassoGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠3"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewMadrassoPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5633: Web は playableIndices で出せる札をリング表示しているのに、CUI は
// 素の一覧だけで、番号を打ってエラーを踏むまで分からなかった。
func TestMadrassoCuiPresenterMarksThePlayableCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MadrassoCuiPresenter)

	// **既定 (.Maybe()) を先に消す。**testify は最初に一致した期待値を返すので、
	// 消さずに足すと上書きしたつもりのケースが何も確かめない。
	setup := func(t *testing.T, phase domain.MadrassoPhase, turn int, playable []int) *interfaces.MockMadrassoGame {
		t.Helper()
		m, players := setupMadrassoCuiMockWithPlayers()
		for _, c := range []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		} {
			players[0].AddCard(c)
		}
		m.ExpectedCalls = madrassoMockWithout(m.ExpectedCalls, "GetPhase", "GetCurrentPlayerIdx", "GetPlayableIndices")
		m.On("GetPhase").Return(phase)
		m.On("GetCurrentPlayerIdx").Return(turn)
		m.On("GetPlayableIndices", mock.Anything).Return(playable)
		return m
	}

	t.Run("marks only what the follow rule allows", func(t *testing.T) {
		m := setup(t, domain.MadrassoPhasePlay, 0, []int{0, 2})
		out := p.Output(m, nil)
		assert.Equal(t, 2, strings.Count(out, presenter.CuiLegalMark))
	})

	t.Run("marks nothing on another player's turn", func(t *testing.T) {
		m := setup(t, domain.MadrassoPhasePlay, 1, []int{0, 2})
		assert.NotContains(t, p.Output(m, nil), presenter.CuiLegalMark)
	})

	t.Run("marks nothing outside the play phase", func(t *testing.T) {
		m := setup(t, domain.MadrassoPhaseRoundEnd, 0, []int{0, 2})
		assert.NotContains(t, p.Output(m, nil), presenter.CuiLegalMark)
	})
}

// madrassoMockWithout drops the listed expectations so a test can override them.
func madrassoMockWithout(calls []*mock.Call, methods ...string) []*mock.Call {
	drop := make(map[string]bool, len(methods))
	for _, m := range methods {
		drop[m] = true
	}
	out := make([]*mock.Call, 0, len(calls))
	for _, c := range calls {
		if !drop[c.Method] {
			out = append(out, c)
		}
	}
	return out
}
