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

func makePokerForPresenter() (*domain.Poker, []*domain.PokerPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
		domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
	}
	cfg := domain.DefaultPokerConfig()
	p := domain.NewPoker(tc, players, cfg)
	return p, players
}

func TestPokerWebPresenter_Output(t *testing.T) {
	pres := new(presenter.PokerWebPresenter)

	t.Run("initial state deal phase", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, domain.PokerPhaseDeal, out.Phase)
		assert.Equal(t, 4, len(out.Players))
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, "", out.Message)
		assert.Len(t, out.SidePots, 0)
		assert.Len(t, out.CpuActions, 0)
		assert.Len(t, out.CpuExchanges, 0)
		assert.Len(t, out.RoundResults, 0)
	})

	t.Run("human cards visible in non-end phase", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		human := out.Players[0]
		assert.True(t, human.IsHuman)
		assert.Len(t, human.Cards, 2)
		assert.Equal(t, "SPADE", human.Cards[0].Design)
		assert.Equal(t, 10, human.Cards[0].Value)
		assert.Equal(t, "HEART", human.Cards[1].Design)
		assert.Equal(t, 11, human.Cards[1].Value)
	})

	t.Run("CPU cards hidden before end phase", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.False(t, cpu.IsHuman)
		assert.Len(t, cpu.Cards, 0)
	})

	t.Run("CPU cards visible at end phase when not folded", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].SetHandRank(domain.PokerHandTwoPair)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 5)
		assert.Equal(t, "SPADE", cpu.Cards[0].Design)
		assert.Equal(t, domain.PokerHandTwoPair, cpu.HandRank)
		assert.Equal(t, "Two Pair", cpu.HandName)
	})

	t.Run("folded CPU cards hidden at end phase", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].SetFolded(true)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 0)
		assert.Equal(t, 0, cpu.HandRank)
		assert.Equal(t, "", cpu.HandName)
	})

	t.Run("hand rank and name only shown at end phase", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[1].SetHandRank(domain.PokerHandFlush)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Equal(t, 0, cpu.HandRank)
		assert.Equal(t, "", cpu.HandName)
	})

	t.Run("player fields", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].SetChips(500)
		players[0].SetCurrentBet(20)
		players[0].SetFolded(false)
		players[0].SetAllIn(false)
		players[0].SetExchangeCount(3)
		players[2].SetFolded(true)
		players[3].SetAllIn(true)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.Players[0].ID)
		assert.True(t, out.Players[0].IsHuman)
		assert.Equal(t, 500, out.Players[0].Chips)
		assert.Equal(t, 20, out.Players[0].CurrentBet)
		assert.False(t, out.Players[0].Folded)
		assert.False(t, out.Players[0].AllIn)
		assert.Equal(t, 3, out.Players[0].ExchangeCount)

		assert.True(t, out.Players[2].Folded)
		assert.True(t, out.Players[3].AllIn)
	})

	t.Run("playStyleName included", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.NotEmpty(t, out.Players[1].PlayStyleName)
	})

	t.Run("pot and game fields", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseSecondBet)
		p.SetPot(300)
		p.SetDealerIdx(2)
		p.SetCurrentTurn(1)
		p.SetLastBet(50)
		p.SetMinRaise(100)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 300, out.Pot)
		assert.Equal(t, 2, out.DealerIdx)
		assert.Equal(t, 1, out.CurrentTurn)
		assert.Equal(t, 50, out.LastBet)
		assert.Equal(t, 100, out.MinRaise)
		assert.Equal(t, domain.PokerPhaseSecondBet, out.Phase)
	})

	t.Run("ante and jokerCount from config", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 10, out.Ante)
		assert.Equal(t, 0, out.JokerCount)
	})

	t.Run("side pots", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetSidePots([]domain.SidePot{
			{Amount: 100, EligiblePlayers: []int{0, 1}},
			{Amount: 50, EligiblePlayers: []int{0}},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.SidePots, 2)
		assert.Equal(t, 100, out.SidePots[0].Amount)
		assert.Equal(t, []int{0, 1}, out.SidePots[0].EligiblePlayers)
		assert.Equal(t, 50, out.SidePots[1].Amount)
		assert.Equal(t, []int{0}, out.SidePots[1].EligiblePlayers)
	})

	t.Run("CPU actions", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetCpuActions([]domain.PokerCpuAction{
			{PlayerIdx: 1, Action: domain.PokerActionCall, Amount: 10},
			{PlayerIdx: 2, Action: domain.PokerActionFold, Amount: 0},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CpuActions, 2)
		assert.Equal(t, 1, out.CpuActions[0].PlayerIdx)
		assert.Equal(t, domain.PokerActionCall, out.CpuActions[0].Action)
		assert.Equal(t, 10, out.CpuActions[0].Amount)
		assert.Equal(t, 2, out.CpuActions[1].PlayerIdx)
		assert.Equal(t, domain.PokerActionFold, out.CpuActions[1].Action)
	})

	t.Run("CPU exchanges", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)
		p.SetCpuExchanges([]domain.PokerCpuExchange{
			{PlayerIdx: 1, ExchangeCount: 3},
			{PlayerIdx: 2, ExchangeCount: 0},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CpuExchanges, 2)
		assert.Equal(t, 1, out.CpuExchanges[0].PlayerIdx)
		assert.Equal(t, 3, out.CpuExchanges[0].ExchangeCount)
		assert.Equal(t, 2, out.CpuExchanges[1].PlayerIdx)
		assert.Equal(t, 0, out.CpuExchanges[1].ExchangeCount)
	})

	t.Run("round results", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 200},
			{PlayerIdx: 1, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 12, 10}, WonAmount: 0},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 2)
		assert.Equal(t, 0, out.RoundResults[0].PlayerIdx)
		assert.Equal(t, domain.PokerHandFlush, out.RoundResults[0].HandRank)
		assert.Equal(t, "Flush", out.RoundResults[0].HandName)
		assert.Equal(t, "", out.RoundResults[0].Kickers)
		assert.Equal(t, 200, out.RoundResults[0].WonAmount)
		assert.Equal(t, 1, out.RoundResults[1].PlayerIdx)
		assert.Equal(t, "A, Q, 10", out.RoundResults[1].Kickers)
		assert.Equal(t, 0, out.RoundResults[1].WonAmount)
	})

	t.Run("error message", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, errors.New("invalid action"))
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "invalid action", out.Message)
	})

	t.Run("error message takes priority over game end", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetGameEndFlag(true)

		result := pres.Output(p, errors.New("some error"))
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "some error", out.Message)
	})

	t.Run("game end message - human wins", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetGameEndFlag(true)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, WonAmount: 100},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "You are the winner.", out.Message)
		assert.Equal(t, "poker.result.win", out.MessageCode)
	})

	t.Run("game end message - human loses", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetGameEndFlag(true)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, WonAmount: 0},
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "You lose.", out.Message)
		assert.Equal(t, "poker.result.lose", out.MessageCode)
	})

	t.Run("game end message - human folded", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetGameEndFlag(true)
		players[0].SetFolded(true)
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "You folded.", out.Message)
		assert.Equal(t, "poker.result.folded", out.MessageCode)
	})

	t.Run("game end message - no results", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetGameEndFlag(true)
		p.SetRoundResults([]domain.PokerResult{})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Game over.", out.Message)
		assert.Equal(t, "poker.result.gameOver", out.MessageCode)
	})

	t.Run("game end - human in results with zero won (not winner)", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseEnd)
		p.SetGameEndFlag(true)
		// Human has WonAmount 0 - loop continues, then falls to "You lose."
		p.SetRoundResults([]domain.PokerResult{
			{PlayerIdx: 0, WonAmount: 0},
		})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "You lose.", out.Message)
		assert.Equal(t, "poker.result.lose", out.MessageCode)
	})

	t.Run("no message when not game end and no error", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetGameEndFlag(false)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "", out.Message)
	})

	t.Run("exchange phase", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 4, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, domain.PokerPhaseExchange, out.Phase)
		assert.Len(t, out.Players[0].Cards, 5)
	})

	t.Run("empty side pots and cpu actions", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetSidePots([]domain.SidePot{})
		p.SetCpuActions([]domain.PokerCpuAction{})
		p.SetCpuExchanges([]domain.PokerCpuExchange{})

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.SidePots, 0)
		assert.Len(t, out.CpuActions, 0)
		assert.Len(t, out.CpuExchanges, 0)
	})

	t.Run("human folded but still visible cards", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].SetFolded(true)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		// Human cards visible (isHuman branch is true)
		assert.Len(t, out.Players[0].Cards, 1)
	})

	t.Run("init phase", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseInit)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, domain.PokerPhaseInit, out.Phase)
	})
}

func TestPokerWebPresenter_OutputWithOdds(t *testing.T) {
	pres := new(presenter.PokerWebPresenter)

	t.Run("with odds data", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)

		odds := []domain.PokerDrawOdds{
			{HandRank: 0, HandName: "High Card", Probability: 0.5, Count: 5, Total: 10},
			{HandRank: 1, HandName: "One Pair", Probability: 0.3, Count: 3, Total: 10},
			{HandRank: 5, HandName: "Flush", Probability: 0.2, Count: 2, Total: 10},
		}

		result := pres.OutputWithOdds(p, nil, odds)
		var out controller.PokerWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Len(t, out.Odds, 3)
		assert.Equal(t, "High Card", out.Odds[0].HandName)
		assert.Equal(t, 0.5, out.Odds[0].Probability)
		assert.Equal(t, 5, out.Odds[0].Count)
		assert.Equal(t, 10, out.Odds[0].Total)
		assert.Equal(t, "One Pair", out.Odds[1].HandName)
		assert.Equal(t, "Flush", out.Odds[2].HandName)
	})

	t.Run("with nil odds", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)

		result := pres.OutputWithOdds(p, nil, nil)
		// odds field should be omitted (omitempty)
		assert.NotContains(t, result, `"odds"`)
	})

	t.Run("with empty non-nil odds", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)

		result := pres.OutputWithOdds(p, nil, []domain.PokerDrawOdds{})
		// empty slice: odds field omitted via omitempty
		assert.NotContains(t, result, `"odds"`)
	})

	t.Run("with error", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseExchange)

		odds := []domain.PokerDrawOdds{
			{HandRank: 0, HandName: "High Card", Probability: 1.0, Count: 1, Total: 1},
		}
		testErr := errors.New("test error")

		result := pres.OutputWithOdds(p, testErr, odds)
		var out controller.PokerWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "test error", out.Message)
		assert.Len(t, out.Odds, 1)
	})
}

func TestPokerWebPresenter_Output_IsLowball(t *testing.T) {
	pres := new(presenter.PokerWebPresenter)

	t.Run("isLowball true", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		}
		cfg := domain.DefaultPokerConfig()
		cfg.IsLowball = true
		cfg.CpuCount = 1
		p := domain.NewPoker(tc, players, cfg)
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.True(t, out.IsLowball)
	})

	t.Run("isLowball false", func(t *testing.T) {
		p, _ := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.False(t, out.IsLowball)
	})
}

func TestPokerWebPresenter_Output_BettingLimitFields(t *testing.T) {
	pres := new(presenter.PokerWebPresenter)

	t.Run("default Fixed limit", func(t *testing.T) {
		p, players := makePokerForPresenter()
		p.SetPhase(domain.PokerPhaseDeal)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 0, out.BettingLimit)
		assert.Equal(t, 0, out.RaiseCount)
		assert.Equal(t, 0, out.MaxBetAmount)
	})

	t.Run("PotLimit returns maxBetAmount", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		}
		cfg := domain.DefaultPokerConfig()
		cfg.BettingLimit = domain.BettingLimitPotLimit
		cfg.CpuCount = 1
		p := domain.NewPoker(tc, players, cfg)
		p.SetPhase(domain.PokerPhaseDeal)
		p.SetPot(100)
		p.SetLastBet(20)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

		result := pres.Output(p, nil)
		var out controller.PokerWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 1, out.BettingLimit)
		assert.Equal(t, 120, out.MaxBetAmount) // pot(100) + lastBet(20)
	})
}

func TestPokerWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PokerWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockPokerGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "exchange", Detail: "exchanged 2 cards", Cards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, true)}},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"actionType":"exchange"`)
		assert.Contains(t, result, `"detail":"exchanged 2 cards"`)
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockPokerGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockPokerGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})
}

func TestPokerWebPresenter_Output_WithMetaAIProfile(t *testing.T) {
	pres := new(presenter.PokerWebPresenter)
	tc := domain.NewTrumpCards(0)
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
		domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
	}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := domain.DefaultPokerConfig()
	cfg.CpuMetaAI = true
	pk := domain.NewPoker(tc, players, cfg)
	_ = pk.Reset()

	result := pres.Output(pk, nil)
	var out controller.PokerWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.MetaAI)
	assert.True(t, out.MetaAI.Enabled)
	assert.NotNil(t, out.Profile)
}

// #5475: 交換枚数が読まれていることを Web にも伝える。閾値は domain 側にあり、
// フロントで数え直さない -- 数え直すと CPU の実際の挙動とずれる。
func TestPokerWebPresenter_ExchangeRead(t *testing.T) {
	pwp := new(presenter.PokerWebPresenter)

	decode := func(phase, humanExchange int) controller.PokerWebOutput {
		g := domain.NewDefaultPoker()
		g.SetPhase(phase)
		g.GetPlayers()[0].SetExchangeCount(humanExchange)
		var out controller.PokerWebOutput
		assert.NoError(t, json.Unmarshal([]byte(pwp.Output(g, nil)), &out))
		return out
	}

	assert.True(t, decode(domain.PokerPhaseSecondBet, 1).ExchangeRead)
	assert.False(t, decode(domain.PokerPhaseSecondBet, 3).ExchangeRead, "閾値ちょうどでは読まれない")
	assert.False(t, decode(domain.PokerPhaseExchange, 0).ExchangeRead, "第2ベット以外では出さない")
}
