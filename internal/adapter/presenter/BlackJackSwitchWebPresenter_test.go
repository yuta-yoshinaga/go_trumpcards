//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBlackJackSwitchWebMockBet(m *interfaces.MockBlackJackSwitchGame) {
	dealer := domain.NewBlackJackPlayer()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 10, true))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	player := domain.NewBlackJackPlayer()
	player.SetChips(800)
	hands := []*domain.BlackJackHand{domain.NewBlackJackHand(), domain.NewBlackJackHand()}
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, true))
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, true))
	hands[0].SetBet(100)
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, true))
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 8, true))
	hands[1].SetBet(100)

	m.On("GetPlayer").Return(player).Maybe()
	m.On("GetDealer").Return(dealer).Maybe()
	m.On("GetHands").Return(hands).Maybe()
	m.On("GetCurrentHandIdx").Return(0).Maybe()
	m.On("GetPhase").Return(domain.BJSwitchPhaseSwitch).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("IsSwitched").Return(false).Maybe()
	m.On("IsDealerPushed22").Return(false).Maybe()
	m.On("GetHandResults").Return(([]domain.GameResult)(nil)).Maybe()
	m.On("GetHandPayouts").Return(([]int)(nil)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetOverallResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestBlackJackSwitchWebPresenter_Output_HidesHoleCard_BeforeEnd(t *testing.T) {
	p := new(BlackJackSwitchWebPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	setupBlackJackSwitchWebMockBet(m)
	out := p.Output(m, nil)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	dealerCards, ok := parsed["dealerCards"].([]any)
	require.True(t, ok)
	require.Len(t, dealerCards, 2)
	assert.Nil(t, dealerCards[1], "second dealer card must be hidden until end phase")
	// Visible score equals up-card only (10).
	assert.Equal(t, float64(10), parsed["dealerScore"])
}

func TestBlackJackSwitchWebPresenter_Output_RevealsAtEnd(t *testing.T) {
	p := new(BlackJackSwitchWebPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	dealer := domain.NewBlackJackPlayer()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 10, true))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	player := domain.NewBlackJackPlayer()
	player.SetChips(1200)
	hands := []*domain.BlackJackHand{domain.NewBlackJackHand(), domain.NewBlackJackHand()}
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 10, true))
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 9, true))
	hands[0].SetBet(100)
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, true))
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 8, true))
	hands[1].SetBet(100)

	m.On("GetPlayer").Return(player).Maybe()
	m.On("GetDealer").Return(dealer).Maybe()
	m.On("GetHands").Return(hands).Maybe()
	m.On("GetCurrentHandIdx").Return(0).Maybe()
	m.On("GetPhase").Return(domain.BJSwitchPhaseEnd).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("IsSwitched").Return(false).Maybe()
	m.On("IsDealerPushed22").Return(false).Maybe()
	m.On("GetHandResults").Return([]domain.GameResult{domain.GameResultWin, domain.GameResultLose}).Maybe()
	m.On("GetHandPayouts").Return([]int{200, 0}).Maybe()
	m.On("GetTotalPayout").Return(200).Maybe()
	m.On("GetOverallResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	out := p.Output(m, nil)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	dealerCards := parsed["dealerCards"].([]any)
	require.Len(t, dealerCards, 2)
	assert.NotNil(t, dealerCards[1], "hole card should be revealed at end phase")
	assert.Equal(t, float64(17), parsed["dealerScore"])
	assert.Equal(t, float64(200), parsed["totalPayout"])
}

func TestBlackJackSwitchWebPresenter_Output_Dealer22Banner(t *testing.T) {
	p := new(BlackJackSwitchWebPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	dealer := domain.NewBlackJackPlayer()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 5, true))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, true)) // 22
	player := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	hands := []*domain.BlackJackHand{domain.NewBlackJackHand(), domain.NewBlackJackHand()}
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 10, true))
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 8, true))
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, true))
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 8, true))

	m.On("GetPlayer").Return(player).Maybe()
	m.On("GetDealer").Return(dealer).Maybe()
	m.On("GetHands").Return(hands).Maybe()
	m.On("GetCurrentHandIdx").Return(0).Maybe()
	m.On("GetPhase").Return(domain.BJSwitchPhaseEnd).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("IsSwitched").Return(false).Maybe()
	m.On("IsDealerPushed22").Return(true).Maybe()
	m.On("GetHandResults").Return([]domain.GameResult{domain.GameResultDraw, domain.GameResultDraw}).Maybe()
	m.On("GetHandPayouts").Return([]int{100, 100}).Maybe()
	m.On("GetTotalPayout").Return(200).Maybe()
	m.On("GetOverallResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	out := p.Output(m, nil)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	assert.Equal(t, true, parsed["dealerPushed22"])
	assert.Equal(t, "blackjackswitch.result.dealer22Push", parsed["messageCode"])
}

func TestBlackJackSwitchWebPresenter_Output_Error(t *testing.T) {
	p := new(BlackJackSwitchWebPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	setupBlackJackSwitchWebMockBet(m)
	out := p.Output(m, errors.New("invalid bet"))
	assert.Contains(t, out, "invalid bet")
}

func TestBlackJackSwitchWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(BlackJackSwitchWebPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	m.On("GetGameEndFlag").Return(false)
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
