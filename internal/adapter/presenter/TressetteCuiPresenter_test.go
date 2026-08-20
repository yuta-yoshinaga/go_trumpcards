package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupTressetteCuiMock() *interfaces.MockTressetteGame {
	m := new(interfaces.MockTressetteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTeamScores").Return([domain.TressetteTeamCnt]int{0, 0})
	m.On("GetTeamRoundThirds").Return([domain.TressetteTeamCnt]int{0, 0})
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TressettePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// 合法手の目印 (#5633)。既定は「制限なし」= 印を出さない状態。
	m.On("GetPlayableIndices", mock.Anything).Return([]int(nil)).Maybe()
	return m
}

func setupTressetteCuiMockWithPlayers() (*interfaces.MockTressetteGame, []*domain.TressettePlayer) {
	m := setupTressetteCuiMock()
	players := makeTressettePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestTressetteCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TressetteCuiPresenter)

	t.Run("play phase shows player info and prompt", func(t *testing.T) {
		m, players := setupTressetteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Tressette")
		assert.NotEmpty(t, result)
		// The thirds-conversion rule is explained alongside the score line.
		assert.Contains(t, result, "サード: 3トリック")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TressettePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows team breakdown and last-trick team", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamRoundThirds")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
		m.On("GetPhase").Return(domain.TressettePhaseRoundEnd)
		m.On("GetTeamRoundThirds").Return([domain.TressetteTeamCnt]int{7, 4})
		m.On("GetLeadPlayerIdx").Return(1) // team B took the last trick
		result := p.Output(m, nil)
		assert.Contains(t, result, strings.Split(i18n.T("tressette.roundBreakdown"), "{{")[0])
		// Each team's thirds appear (7 and 4).
		assert.Contains(t, result, "thirds7")
		assert.Contains(t, result, "thirds4")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestTressetteCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TressetteCuiPresenter)

	t.Run("hint available", func(t *testing.T) {
		m, players := setupTressetteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetHint").Return(&domain.TressetteHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTressetteCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.TressetteHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})
}

func TestTressetteCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TressetteCuiPresenter)
	m := new(interfaces.MockTressetteGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays ♠3"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewTressettePlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5633: Web は playableIndices で出せる札をリング表示しているのに、CUI は
// 素の一覧だけで、番号を打ってエラーを踏むまで分からなかった。
func TestTressetteCuiPresenterMarksThePlayableCards(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TressetteCuiPresenter)

	// **既定 (.Maybe()) を先に消す。**testify は最初に一致した期待値を返すので、
	// 消さずに足すと上書きしたつもりのケースが何も確かめない。
	setup := func(t *testing.T, phase domain.TressettePhase, turn int, playable []int) *interfaces.MockTressetteGame {
		t.Helper()
		m, players := setupTressetteCuiMockWithPlayers()
		for _, c := range []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		} {
			players[0].AddCard(c)
		}
		m.ExpectedCalls = tressetteMockWithout(m.ExpectedCalls, "GetPhase", "GetCurrentPlayerIdx", "GetPlayableIndices")
		m.On("GetPhase").Return(phase)
		m.On("GetCurrentPlayerIdx").Return(turn)
		m.On("GetPlayableIndices", mock.Anything).Return(playable)
		return m
	}

	t.Run("marks only what the follow rule allows", func(t *testing.T) {
		m := setup(t, domain.TressettePhasePlay, 0, []int{0, 2})
		out := p.Output(m, nil)
		assert.Equal(t, 2, strings.Count(out, presenter.CuiLegalMark))
	})

	t.Run("marks nothing on another player's turn", func(t *testing.T) {
		m := setup(t, domain.TressettePhasePlay, 1, []int{0, 2})
		assert.NotContains(t, p.Output(m, nil), presenter.CuiLegalMark)
	})

	t.Run("marks nothing outside the play phase", func(t *testing.T) {
		m := setup(t, domain.TressettePhaseRoundEnd, 0, []int{0, 2})
		assert.NotContains(t, p.Output(m, nil), presenter.CuiLegalMark)
	})
}

// tressetteMockWithout drops the listed expectations so a test can override them.
func tressetteMockWithout(calls []*mock.Call, methods ...string) []*mock.Call {
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
