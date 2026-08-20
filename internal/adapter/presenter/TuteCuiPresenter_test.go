//go:build test

package presenter_test

import (
	"errors"
	"strconv"
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

func makeTutePlayers() []*domain.TutePlayer {
	return []*domain.TutePlayer{
		domain.NewTutePlayer(true),
		domain.NewTutePlayer(false),
		domain.NewTutePlayer(false),
		domain.NewTutePlayer(false),
	}
}

func setupTuteCuiMock() *interfaces.MockTuteGame {
	m := new(interfaces.MockTuteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TutePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("CanHumanDeclareMarriage").Return(false)
	m.On("GetHumanDeclarableMarriageSuits").Return(([]int)(nil))
	m.On("CanHumanDeclareTute").Return(false)
	m.On("GetRoundTeamPoints").Return([domain.TuteTeamCnt]int{0, 0})
	m.On("GetTeamScores").Return([domain.TuteTeamCnt]int{0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupTuteCuiMockWithPlayers() (*interfaces.MockTuteGame, []*domain.TutePlayer) {
	m := setupTuteCuiMock()
	players := makeTutePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestTuteCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TuteCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupTuteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Tute")
		assert.NotEmpty(t, result)
	})

	t.Run("play phase with marriage prompt", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "CanHumanDeclareMarriage")
		m.On("CanHumanDeclareMarriage").Return(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("marriage prompt lists declarable suits with trump mark", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "CanHumanDeclareMarriage")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHumanDeclarableMarriageSuits")
		m.On("CanHumanDeclareMarriage").Return(true)
		// Trump is spades (default mock); heart is a plain marriage.
		m.On("GetHumanDeclarableMarriageSuits").Return([]int{domain.CardDesignSpade, domain.CardDesignHeart})
		result := p.Output(m, nil)
		assert.Contains(t, result, strings.Split(i18n.T("tute.promptMarriageSuits"), "{{")[0])
		// The trump suit (spades) carries the +40 marker.
		assert.Contains(t, result, i18n.T("tute.marriageTrumpMark"))
	})

	t.Run("play phase with tute prompt", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "CanHumanDeclareTute")
		m.On("CanHumanDeclareTute").Return(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TutePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TutePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestTuteCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.TuteCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.TuteHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupTuteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.TuteHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.TuteHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("marriage hint", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.TuteHint{Marriage: domain.CardDesignSpade, Reason: "declare_marriage"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("tute hint", func(t *testing.T) {
		m, _ := setupTuteCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.TuteHint{Reason: "declare_tute"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestTuteCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TuteCuiPresenter)
	m := new(interfaces.MockTuteGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewTutePlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5641: 結婚 (同スートの K+Q) は宣言した瞬間に加点される。Web は #4722 を受けて
// プレイ中も tute-running-points で今ラウンドの点を出しているのに、CUI は
// GetRoundTeamPoints を RoundEnd でしか読んでおらず、宣言しても何点入ったのか
// ラウンドが終わるまで分からなかった。
func TestTuteCuiPresenter_ShowsTheRunningRoundPoints(t *testing.T) {
	p := new(presenter.TuteCuiPresenter)

	phaseMock := func(phase domain.TutePhase) *interfaces.MockTuteGame {
		m, _ := setupTuteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundTeamPoints")
		m.On("GetPhase").Return(phase)
		m.On("GetRoundTeamPoints").Return([domain.TuteTeamCnt]int{42, 17})
		return m
	}

	for _, phase := range []domain.TutePhase{domain.TutePhasePlay, domain.TutePhaseTrickEnd} {
		t.Run("running points during "+strconv.Itoa(int(phase)), func(t *testing.T) {
			out := p.Output(phaseMock(phase), nil)

			assert.Contains(t, out, i18n.Tf("tute.runningPoints",
				"ptsA", "42", "ptsB", "17"))
		})
	}

	// RoundEnd は自分の行を持っているので、同じ数字を二度出さない。
	t.Run("round end keeps its own single line", func(t *testing.T) {
		out := p.Output(phaseMock(domain.TutePhaseRoundEnd), nil)

		assert.Contains(t, out, i18n.Tf("tute.promptRoundEnd", "ptsA", "42", "ptsB", "17"))
		assert.NotContains(t, out, i18n.Tf("tute.runningPoints", "ptsA", "42", "ptsB", "17"))
	})
}
