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

func tcbMakePlayers() []*domain.ThreeCardBragPlayer {
	return []*domain.ThreeCardBragPlayer{
		domain.NewThreeCardBragPlayer(true, 30),
		domain.NewThreeCardBragPlayer(false, 30),
		domain.NewThreeCardBragPlayer(false, 30),
		domain.NewThreeCardBragPlayer(false, 30),
	}
}

func tcbSetupBaseMock() *interfaces.MockThreeCardBragGame {
	m := new(interfaces.MockThreeCardBragGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetPot").Return(4)
	m.On("GetStake").Return(1)
	m.On("GetPhase").Return(domain.ThreeCardBragPhaseBetting)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetMatchWinnerIdx").Return(-1)
	m.On("IsShowdown").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func tcbSetupMockWithPlayers() (*interfaces.MockThreeCardBragGame, []*domain.ThreeCardBragPlayer) {
	m := tcbSetupBaseMock()
	players := tcbMakePlayers()
	m.On("GetPlayerCnt").Return(domain.ThreeCardBragPlayerCnt)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestThreeCardBragCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ThreeCardBragCuiPresenter)

	t.Run("betting phase shows prompt and human cards", func(t *testing.T) {
		m, players := tcbSetupMockWithPlayers()
		players[0].SetSeen(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "SPADE")
	})

	t.Run("showdown reveals all non-folded hands", func(t *testing.T) {
		m, players := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseShowdown)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsShowdown")
		m.On("IsShowdown").Return(true)
		for i := range players {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
			players[i].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
			players[i].AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		}
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end shows winner prompt", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThreeCardBragPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWinnerIdx")
		m.On("GetRoundWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMatchWinnerIdx")
		m.On("GetMatchWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("folded and out statuses render", func(t *testing.T) {
		m, players := tcbSetupMockWithPlayers()
		players[1].SetFolded(true)
		players[2].SetOut(true)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestThreeCardBragCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ThreeCardBragCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.On("GetHint").Return((*domain.ThreeCardBragHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("see hint", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.On("GetHint").Return(&domain.ThreeCardBragHint{Action: "see", Reason: "see_first"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "threecardbrag.hint")
	})

	t.Run("raise hint", func(t *testing.T) {
		m, _ := tcbSetupMockWithPlayers()
		m.On("GetHint").Return(&domain.ThreeCardBragHint{Action: "raise", Reason: "strong_hand"})
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "threecardbrag.hint")
	})
}

func TestThreeCardBragCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ThreeCardBragCuiPresenter)
	m := new(interfaces.MockThreeCardBragGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "You bets 1"},
	})
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "bet")
}
