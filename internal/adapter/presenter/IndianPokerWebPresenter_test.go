package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestIndianPokerWebPresenter_Output(t *testing.T) {
	p := new(presenter.IndianPokerWebPresenter)

	setup := func() (*domain.IndianPoker, []*domain.IndianPokerPlayer) {
		tc := domain.NewTrumpCards(0)
		players := domain.NewIndianPokerPlayers()
		ip := domain.NewIndianPoker(tc, players, domain.DefaultIndianPokerConfig())
		return ip, players
	}

	t.Run("initial state", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, domain.IndianPokerPhaseBetting, out.Phase)
		assert.Equal(t, 4, len(out.Players))
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, "", out.Message)
		assert.Len(t, out.SidePots, 0)
		assert.Len(t, out.CpuActions, 0)
		assert.Len(t, out.RoundResults, 0)
	})

	t.Run("human card null during betting", func(t *testing.T) {
		ip, players := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		human := out.Players[0]
		assert.True(t, human.IsHuman)
		assert.Nil(t, human.Card)
	})

	t.Run("human card visible at showdown", func(t *testing.T) {
		ip, players := setup()
		ip.SetPhase(domain.IndianPokerPhaseShowdown)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		human := out.Players[0]
		assert.NotNil(t, human.Card)
		assert.Equal(t, "SPADE", human.Card.Design)
		assert.Equal(t, 10, human.Card.Value)
	})

	t.Run("human card visible at end", func(t *testing.T) {
		ip, players := setup()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		human := out.Players[0]
		assert.NotNil(t, human.Card)
		assert.Equal(t, "HEART", human.Card.Design)
	})

	t.Run("CPU card always visible", func(t *testing.T) {
		ip, players := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.NotNil(t, cpu.Card)
		assert.Equal(t, "CLOVER", cpu.Card.Design)
		assert.Equal(t, 7, cpu.Card.Value)
	})

	t.Run("player without cards", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		// No cards added, Card should be nil
		assert.Nil(t, out.Players[0].Card)
	})

	t.Run("player fields", func(t *testing.T) {
		ip, players := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		players[0].SetChips(500)
		players[0].SetCurrentBet(20)
		players[0].SetFolded(false)
		players[0].SetAllIn(false)
		players[2].SetFolded(true)
		players[3].SetAllIn(true)

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.Players[0].ID)
		assert.True(t, out.Players[0].IsHuman)
		assert.Equal(t, 500, out.Players[0].Chips)
		assert.Equal(t, 20, out.Players[0].CurrentBet)
		assert.False(t, out.Players[0].Folded)
		assert.False(t, out.Players[0].AllIn)

		assert.True(t, out.Players[2].Folded)
		assert.True(t, out.Players[3].AllIn)
	})

	t.Run("side pots", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetSidePots([]domain.SidePot{
			{Amount: 100, EligiblePlayers: []int{0, 1}},
			{Amount: 50, EligiblePlayers: []int{0}},
		})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.SidePots, 2)
		assert.Equal(t, 100, out.SidePots[0].Amount)
		assert.Equal(t, []int{0, 1}, out.SidePots[0].EligiblePlayers)
		assert.Equal(t, 50, out.SidePots[1].Amount)
	})

	t.Run("CPU actions", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetCpuActions([]domain.IndianPokerCpuAction{
			{PlayerIdx: 1, Action: domain.IndianPokerActionCall, Amount: 10},
			{PlayerIdx: 2, Action: domain.IndianPokerActionFold, Amount: 0},
		})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CpuActions, 2)
		assert.Equal(t, 1, out.CpuActions[0].PlayerIdx)
		assert.Equal(t, domain.IndianPokerActionCall, out.CpuActions[0].Action)
		assert.Equal(t, 10, out.CpuActions[0].Amount)
		assert.Equal(t, 2, out.CpuActions[1].PlayerIdx)
	})

	t.Run("round results with card", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{
				PlayerIdx: 0,
				Card:      domain.NewCard(domain.CardDesignSpade, 14, false),
				CardRank:  14,
				WonAmount: 200,
			},
		})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 1)
		assert.Equal(t, 0, out.RoundResults[0].PlayerIdx)
		assert.NotNil(t, out.RoundResults[0].Card)
		assert.Equal(t, "SPADE", out.RoundResults[0].Card.Design)
		assert.Equal(t, 14, out.RoundResults[0].CardRank)
		assert.Equal(t, 200, out.RoundResults[0].WonAmount)
	})

	t.Run("round results without card", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{
				PlayerIdx: 1,
				Card:      nil,
				WonAmount: 80,
			},
		})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 1)
		assert.Nil(t, out.RoundResults[0].Card)
	})

	t.Run("pot and dealer fields", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)
		ip.SetPot(300)
		ip.SetDealerIdx(2)
		ip.SetCurrentTurn(1)
		ip.SetLastBet(50)
		ip.SetMinRaise(100)
		ip.SetRaiseCount(2)
		ip.SetHandCount(5)

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 300, out.Pot)
		assert.Equal(t, 2, out.DealerIdx)
		assert.Equal(t, 1, out.CurrentTurn)
		assert.Equal(t, 50, out.LastBet)
		assert.Equal(t, 100, out.MinRaise)
		assert.Equal(t, 2, out.RaiseCount)
		assert.Equal(t, 5, out.HandCount)
	})

	t.Run("error message", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, errors.New("invalid action"))
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "invalid action", out.Message)
	})

	t.Run("error message takes priority over game end", func(t *testing.T) {
		ip, _ := setup()
		ip.SetGameEndFlag(true)

		result := p.Output(ip, errors.New("some error"))
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "some error", out.Message)
	})

	t.Run("game end message - human wins", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		ip.SetGameEndFlag(true)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{PlayerIdx: 0, WonAmount: 100},
		})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "indianpoker.result.win", out.MessageCode)
	})

	t.Run("game end message - human loses", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		ip.SetGameEndFlag(true)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{PlayerIdx: 0, WonAmount: 0},
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "indianpoker.result.lose", out.MessageCode)
	})

	t.Run("game end message - human folded", func(t *testing.T) {
		ip, players := setup()
		ip.SetPhase(domain.IndianPokerPhaseEnd)
		ip.SetGameEndFlag(true)
		players[0].SetFolded(true)
		ip.SetRoundResults([]domain.IndianPokerResult{
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "indianpoker.result.folded", out.MessageCode)
	})

	t.Run("game end message - no results", func(t *testing.T) {
		ip, _ := setup()
		ip.SetGameEndFlag(true)
		ip.SetRoundResults([]domain.IndianPokerResult{})

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "indianpoker.result.gameOver", out.MessageCode)
	})

	t.Run("no game end no error returns empty message", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "", out.Message)
		assert.Equal(t, "", out.MessageCode)
	})

	t.Run("config fields in output", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, int(domain.BettingLimitNoLimit), out.BettingLimit)
		assert.Equal(t, 10, out.Ante)
	})

	t.Run("play style name included", func(t *testing.T) {
		ip, _ := setup()
		ip.SetPhase(domain.IndianPokerPhaseBetting)

		result := p.Output(ip, nil)
		var out controller.IndianPokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		// CPU players should have play style names
		assert.NotEmpty(t, out.Players[1].PlayStyleName)
	})
}

func TestIndianPokerWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.IndianPokerWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockIndianPokerGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "bet 100"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		assert.NoError(t, err)
		mockGame.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		mockGame := new(interfaces.MockIndianPokerGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		assert.NoError(t, err)
		mockGame.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		mockGame := new(interfaces.MockIndianPokerGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(result), &parsed)
		assert.NoError(t, err)
		mockGame.AssertExpectations(t)
	})
}

func TestIndianPokerWebPresenter_Output_WithMetaAIProfile(t *testing.T) {
	p := new(presenter.IndianPokerWebPresenter)
	tc := domain.NewTrumpCards(0)
	players := []*domain.IndianPokerPlayer{
		domain.NewIndianPokerPlayer(true, domain.HoldemStyleTAG),
		domain.NewIndianPokerPlayer(false, domain.HoldemStyleLAP),
		domain.NewIndianPokerPlayer(false, domain.HoldemStyleTAP),
		domain.NewIndianPokerPlayer(false, domain.HoldemStyleGTO),
	}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := domain.DefaultIndianPokerConfig()
	cfg.CpuMetaAI = true
	ip := domain.NewIndianPoker(tc, players, cfg)
	_ = ip.Reset()

	result := p.Output(ip, nil)
	var out controller.IndianPokerWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.MetaAI)
	assert.True(t, out.MetaAI.Enabled)
	assert.NotNil(t, out.Profile)
}
