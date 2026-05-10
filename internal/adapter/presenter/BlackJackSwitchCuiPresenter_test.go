//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// setupBlackJackSwitchCuiMockDefaults registers permissive Maybe defaults so the
// presenter can call any getter without exhaustive On() declarations per case.
func setupBlackJackSwitchCuiMockDefaults(m *interfaces.MockBlackJackSwitchGame) {
	dealer := domain.NewBlackJackPlayer()
	player := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	hands := []*domain.BlackJackHand{domain.NewBlackJackHand(), domain.NewBlackJackHand()}
	m.On("GetPlayer").Return(player).Maybe()
	m.On("GetDealer").Return(dealer).Maybe()
	m.On("GetHands").Return(hands).Maybe()
	m.On("GetCurrentHandIdx").Return(0).Maybe()
	m.On("GetPhase").Return(domain.BJSwitchPhaseBet).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("IsSwitched").Return(false).Maybe()
	m.On("IsDealerPushed22").Return(false).Maybe()
	m.On("GetHandResults").Return(([]domain.GameResult)(nil)).Maybe()
	m.On("GetHandPayouts").Return(([]int)(nil)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetOverallResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestBlackJackSwitchCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(BlackJackSwitchCuiPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	setupBlackJackSwitchCuiMockDefaults(m)
	out := p.Output(m, nil)
	assert.Contains(t, out, "1000")
	assert.Contains(t, out, "BET")
}

func TestBlackJackSwitchCuiPresenter_Output_Error(t *testing.T) {
	p := new(BlackJackSwitchCuiPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	setupBlackJackSwitchCuiMockDefaults(m)
	out := p.Output(m, errors.New("invalid"))
	assert.Contains(t, out, "invalid")
}

func TestBlackJackSwitchCuiPresenter_Output_DealerHoleHidden_DuringAction(t *testing.T) {
	p := new(BlackJackSwitchCuiPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	dealer := domain.NewBlackJackPlayer()
	dealer.AddCard(domain.NewCard(domain.CardDesignSpade, 10, true))
	dealer.AddCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	player := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	hands := []*domain.BlackJackHand{domain.NewBlackJackHand(), domain.NewBlackJackHand()}
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 5, true))
	hands[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, true))
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, true))
	hands[1].AddCard(domain.NewCard(domain.CardDesignDiamond, 8, true))

	m.On("GetPlayer").Return(player).Maybe()
	m.On("GetDealer").Return(dealer).Maybe()
	m.On("GetHands").Return(hands).Maybe()
	m.On("GetCurrentHandIdx").Return(0).Maybe()
	m.On("GetPhase").Return(domain.BJSwitchPhaseAction).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("IsSwitched").Return(false).Maybe()
	m.On("IsDealerPushed22").Return(false).Maybe()
	m.On("GetHandResults").Return(([]domain.GameResult)(nil)).Maybe()
	m.On("GetHandPayouts").Return(([]int)(nil)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetOverallResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	out := p.Output(m, nil)
	assert.Contains(t, out, "??", "dealer hole card should be hidden during action phase")
	assert.Contains(t, out, "ACTION")
}

func TestBlackJackSwitchCuiPresenter_Output_EndPhaseShowsResults(t *testing.T) {
	p := new(BlackJackSwitchCuiPresenter)
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
	m.On("IsSwitched").Return(true).Maybe()
	m.On("IsDealerPushed22").Return(false).Maybe()
	m.On("GetHandResults").Return([]domain.GameResult{domain.GameResultWin, domain.GameResultLose}).Maybe()
	m.On("GetHandPayouts").Return([]int{200, 0}).Maybe()
	m.On("GetTotalPayout").Return(200).Maybe()
	m.On("GetOverallResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	out := p.Output(m, nil)
	assert.Contains(t, out, "200") // total payout / hand0 payout
	assert.NotContains(t, out, "??", "dealer hole card should be revealed at end phase")
}

func TestBlackJackSwitchCuiPresenter_Output_Dealer22ShowsPushBanner(t *testing.T) {
	p := new(BlackJackSwitchCuiPresenter)
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
	m.On("IsDealerPushed22").Return(true).Maybe()
	m.On("GetHandResults").Return([]domain.GameResult{domain.GameResultDraw, domain.GameResultDraw}).Maybe()
	m.On("GetHandPayouts").Return([]int{100, 100}).Maybe()
	m.On("GetTotalPayout").Return(200).Maybe()
	m.On("GetOverallResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	out := p.Output(m, nil)
	// Either ja "ディーラー22" or en "dealer 22" depending on locale; just look for "22"
	assert.Contains(t, out, "22")
}

func TestBlackJackSwitchCuiPresenter_PhaseStr(t *testing.T) {
	p := new(BlackJackSwitchCuiPresenter)
	assert.Equal(t, "BET", p.phaseStr(domain.BJSwitchPhaseBet))
	assert.Equal(t, "SWITCH", p.phaseStr(domain.BJSwitchPhaseSwitch))
	assert.Equal(t, "ACTION", p.phaseStr(domain.BJSwitchPhaseAction))
	assert.Equal(t, "END", p.phaseStr(domain.BJSwitchPhaseEnd))
	assert.Equal(t, "UNKNOWN", p.phaseStr(99))
}

func TestBlackJackSwitchCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(BlackJackSwitchCuiPresenter)
	m := new(interfaces.MockBlackJackSwitchGame)
	m.On("GetGameEndFlag").Return(false)
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
