package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupVideoPokerWebMockDefaults(m *interfaces.MockVideoPokerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseBet).Maybe()
	m.On("GetHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHandRank").Return(0).Maybe()
	m.On("GetHandName").Return("").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("jacksorbetter").Maybe()
}

func parseVideoPokerOutput(t *testing.T, jsonStr string) *controller.VideoPokerWebOutput {
	t.Helper()
	var out controller.VideoPokerWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestVideoPokerWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(VideoPokerWebPresenter)
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerWebMockDefaults(m)

	result := parseVideoPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.VideoPokerPhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.Hand)
	assert.Empty(t, result.Message)
}

func TestVideoPokerWebPresenter_Output_Win(t *testing.T) {
	p := new(VideoPokerWebPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(1025).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseResult).Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(1).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(25).Maybe()
	m.On("GetHandRank").Return(domain.PokerHandFourOfAKind).Maybe()
	m.On("GetHandName").Return("Four of a Kind").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{true, true, true, true, false}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("jacksorbetter").Maybe()

	result := parseVideoPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, "Four of a Kind! You win!", result.Message)
	assert.Equal(t, "videopoker.result.win", result.MessageCode)
	assert.Equal(t, "Four of a Kind", result.MessageParams["handName"])
	assert.Equal(t, "25", result.MessageParams["payout"])
	assert.Equal(t, 25, result.Payout)
	assert.Len(t, result.Hand, 5)
}

func TestVideoPokerWebPresenter_Output_Lose(t *testing.T) {
	p := new(VideoPokerWebPresenter)
	m := new(interfaces.MockVideoPokerGame)
	m.On("GetChips").Return(999).Maybe()
	m.On("GetPhase").Return(domain.VideoPokerPhaseResult).Maybe()
	m.On("GetHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 11, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(1).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHandRank").Return(domain.PokerHandHighCard).Maybe()
	m.On("GetHandName").Return("").Maybe()
	m.On("GetHeldIndices").Return([domain.VideoPokerHandSize]bool{}).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
	m.On("GetVariantName").Return("jacksorbetter").Maybe()

	result := parseVideoPokerOutput(t, p.Output(m, nil))
	assert.Equal(t, "No winning hand.", result.Message)
	assert.Equal(t, "videopoker.result.lose", result.MessageCode)
}

func TestVideoPokerWebPresenter_Output_Error(t *testing.T) {
	p := new(VideoPokerWebPresenter)
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerWebMockDefaults(m)

	result := parseVideoPokerOutput(t, p.Output(m, domain.NewDomainError(domain.ErrInvalidAmount, "Bet must be between 1 and 5 coins.")))
	assert.Equal(t, "Bet must be between 1 and 5 coins.", result.Message)
}

func TestVideoPokerWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(VideoPokerWebPresenter)

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetGameEndFlag").Return(false)
		jsonStr := p.ActionLogOutput(m)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(jsonStr), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})

	t.Run("game ended with log", func(t *testing.T) {
		m := new(interfaces.MockVideoPokerGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "bet 3 coin(s)"},
		})
		jsonStr := p.ActionLogOutput(m)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(jsonStr), &out)
		assert.NoError(t, err)
		assert.Len(t, out.Entries, 1)
	})
}

func TestVideoPokerWebPresenter_HintOutput(t *testing.T) {
	m := new(interfaces.MockVideoPokerGame)
	setupVideoPokerWebMockDefaults(m)
	p := new(VideoPokerWebPresenter)
	// The web presenter computes hints client-side, so HintOutput mirrors Output.
	assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
}
