//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupAllFoursCuiMock() *interfaces.MockAllFoursGame {
	m := new(interfaces.MockAllFoursGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(0)
	m.On("GetDealerIdx").Return(1)
	m.On("GetNonDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetTurnUp").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetRunCount").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.AllFoursPhaseBeg)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLeadPlayerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultAllFoursConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupAllFoursCuiMockWithPlayers() (*interfaces.MockAllFoursGame, []*domain.AllFoursPlayer) {
	m := setupAllFoursCuiMock()
	players := makeAllFoursPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestAllFoursCuiPresenter_Output_Beg(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.AllFoursCuiPresenter)

	m, players := setupAllFoursCuiMockWithPlayers()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

	result := p.Output(m, nil)
	assert.Contains(t, result, "All Fours")
	assert.Contains(t, result, "ラウンド: 1")
	assert.Contains(t, result, "親: CPU 1")
	assert.Contains(t, result, "ベグフェーズ")
}

func TestAllFoursCuiPresenter_Output_Play(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.AllFoursCuiPresenter)

	m, _ := setupAllFoursCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
	m.On("GetPhase").Return(domain.AllFoursPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "手番: あなた")
}

func TestAllFoursCuiPresenter_Output_GameEnd(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.AllFoursCuiPresenter)

	m, _ := setupAllFoursCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "ゲーム終了")
}

func TestAllFoursCuiPresenter_Output_Error(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.AllFoursCuiPresenter)

	m, _ := setupAllFoursCuiMockWithPlayers()
	result := p.Output(m, errors.New("bad play"))
	assert.Contains(t, result, "bad play")
}

func TestAllFoursCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.AllFoursCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m, _ := setupAllFoursCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.AllFoursHint)(nil))
		out := p.HintOutput(m)
		assert.Contains(t, out, "ヒントはありません")
	})

	t.Run("beg hint", func(t *testing.T) {
		m, _ := setupAllFoursCuiMockWithPlayers()
		beg := true
		m.On("GetHint").Return(&domain.AllFoursHint{Beg: &beg, Reason: "beg_beg"})
		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
	})

	t.Run("run hint", func(t *testing.T) {
		m, _ := setupAllFoursCuiMockWithPlayers()
		run := false
		m.On("GetHint").Return(&domain.AllFoursHint{Run: &run, Reason: "gift_gift"})
		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupAllFoursCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		idx := 0
		m.On("GetHint").Return(&domain.AllFoursHint{CardIndex: &idx, Reason: "trump_cut"})
		out := p.HintOutput(m)
		assert.Contains(t, out, "HINT")
	})
}

func TestAllFoursCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.AllFoursCuiPresenter)
	m := new(interfaces.MockAllFoursGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "stand", Detail: "You stand"},
	})
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "You stand")
}
