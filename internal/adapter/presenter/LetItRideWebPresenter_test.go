package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupLetItRideWebMockDefaults(m *interfaces.MockLetItRideGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.LetItRidePhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetBet1Active").Return(false).Maybe()
	m.On("GetBet2Active").Return(false).Maybe()
	m.On("GetBet3Active").Return(false).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetHandRank").Return(0).Maybe()
	m.On("GetBet1Payout").Return(0).Maybe()
	m.On("GetBet2Payout").Return(0).Maybe()
	m.On("GetBet3Payout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseLetItRideOutput(t *testing.T, jsonStr string) *controller.LetItRideWebOutput {
	t.Helper()
	var out controller.LetItRideWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestLetItRideWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(LetItRideWebPresenter)
	m := new(interfaces.MockLetItRideGame)
	setupLetItRideWebMockDefaults(m)

	result := parseLetItRideOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.LetItRidePhaseBet, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.CommunityCards)
	assert.Empty(t, result.Message)
}

func TestLetItRideWebPresenter_Output_Error(t *testing.T) {
	p := new(LetItRideWebPresenter)
	m := new(interfaces.MockLetItRideGame)
	setupLetItRideWebMockDefaults(m)

	result := parseLetItRideOutput(t, p.Output(m, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
}

func TestLetItRideWebPresenter_Output_FirstDecision_CommunityMasked(t *testing.T) {
	p := new(LetItRideWebPresenter)
	m := new(interfaces.MockLetItRideGame)
	setupLetItRideWebMockDefaults(m)

	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
	}
	m.ExpectedCalls = nil
	setupLetItRideWebMockDefaults(m)
	m.On("GetPhase").Return(domain.LetItRidePhaseFirstDecision).Unset()
	m.On("GetPhase").Return(domain.LetItRidePhaseFirstDecision)
	m.On("GetPlayerHand").Return(cards).Unset()
	m.On("GetPlayerHand").Return(cards)
	m.On("GetCommunityCards").Return(community).Unset()
	m.On("GetCommunityCards").Return(community)

	result := parseLetItRideOutput(t, p.Output(m, nil))
	assert.Len(t, result.PlayerHand, 3)
	assert.Len(t, result.CommunityCards, 2)
	// Both community cards should be masked in first decision
	for _, c := range result.CommunityCards {
		assert.Equal(t, "", c.Design)
		assert.Equal(t, 0, c.Value)
	}
}

func TestLetItRideWebPresenter_Output_SecondDecision_FirstRevealed(t *testing.T) {
	p := new(LetItRideWebPresenter)
	m := new(interfaces.MockLetItRideGame)

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
	}
	m.On("GetChips").Return(700)
	m.On("GetPhase").Return(domain.LetItRidePhaseSecondDecision)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetCommunityCards").Return(community)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetBetAmount").Return(100)
	m.On("GetBet1Active").Return(true)
	m.On("GetBet2Active").Return(true)
	m.On("GetBet3Active").Return(false)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetHandRank").Return(0)
	m.On("GetBet1Payout").Return(0)
	m.On("GetBet2Payout").Return(0)
	m.On("GetBet3Payout").Return(0)
	m.On("GetTotalPayout").Return(0)

	result := parseLetItRideOutput(t, p.Output(m, nil))
	assert.Len(t, result.CommunityCards, 2)
	// First community card revealed
	assert.NotEqual(t, "", result.CommunityCards[0].Design)
	// Second still masked
	assert.Equal(t, "", result.CommunityCards[1].Design)
}

func TestLetItRideWebPresenter_Output_EndPhase_AllRevealed(t *testing.T) {
	p := new(LetItRideWebPresenter)
	m := new(interfaces.MockLetItRideGame)

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
	}
	m.On("GetChips").Return(1600)
	m.On("GetPhase").Return(domain.LetItRidePhaseEnd)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetCommunityCards").Return(community)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBetAmount").Return(100)
	m.On("GetBet1Active").Return(true)
	m.On("GetBet2Active").Return(true)
	m.On("GetBet3Active").Return(true)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetHandRank").Return(domain.PokerHandTwoPair)
	m.On("GetBet1Payout").Return(300)
	m.On("GetBet2Payout").Return(300)
	m.On("GetBet3Payout").Return(300)
	m.On("GetTotalPayout").Return(900)

	result := parseLetItRideOutput(t, p.Output(m, nil))
	// Both community cards revealed
	assert.NotEqual(t, "", result.CommunityCards[0].Design)
	assert.NotEqual(t, "", result.CommunityCards[1].Design)
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "letitride.result.playerWins", result.MessageCode)
	assert.Equal(t, 900, result.TotalPayout)
}

func TestLetItRideWebPresenter_Output_EndPhase_Loss(t *testing.T) {
	p := new(LetItRideWebPresenter)
	m := new(interfaces.MockLetItRideGame)

	m.On("GetChips").Return(700)
	m.On("GetPhase").Return(domain.LetItRidePhaseEnd)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBetAmount").Return(100)
	m.On("GetBet1Active").Return(true)
	m.On("GetBet2Active").Return(true)
	m.On("GetBet3Active").Return(true)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetHandRank").Return(0)
	m.On("GetBet1Payout").Return(0)
	m.On("GetBet2Payout").Return(0)
	m.On("GetBet3Payout").Return(0)
	m.On("GetTotalPayout").Return(0)

	result := parseLetItRideOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player loses.", result.Message)
	assert.Equal(t, "letitride.result.playerLoses", result.MessageCode)
}

func TestLetItRideWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(LetItRideWebPresenter)
	m := new(interfaces.MockLetItRideGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetGameEndFlag").Return(false)

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

func TestLetItRideMaskCommunity_EmptyCards(t *testing.T) {
	m := new(interfaces.MockLetItRideGame)
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil))
	m.On("GetPhase").Return(domain.LetItRidePhaseBet)

	result := letItRideMaskCommunity(m)
	assert.Empty(t, result)
}
