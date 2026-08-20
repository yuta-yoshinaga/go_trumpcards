//go:build test

package presenter

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
	m.On("SwitchPreviewScores").Return(0, 0, false).Maybe()
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
	// Per-hand results resolve via the unified handWin/handLose keys.
	assert.Contains(t, out, "勝ち")
	assert.Contains(t, out, "負け")
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

// #5586: 交換すると得か損かは、`switch` を打って結果を見るまで分からなかった。
// Web はホバーで先読みを出している。
func TestBlackJackSwitchCuiPresenter_ShowsTheSwitchPreview(t *testing.T) {
	i18n.SetLang("ja")
	build := func(phase, first, second int, ok bool) string {
		m := new(interfaces.MockBlackJackSwitchGame)
		setupBlackJackSwitchCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "SwitchPreviewScores")
		m.On("GetPhase").Return(phase)
		m.On("SwitchPreviewScores").Return(first, second, ok)
		return new(BlackJackSwitchCuiPresenter).Output(m, nil)
	}

	out := build(domain.BJSwitchPhaseSwitch, 20, 18, true)
	assert.Contains(t, out, i18n.Tf("blackjackswitch.switchPreviewLine", "first", "20", "second", "18"))

	// **21 超えはバーストと分かる形で出す** (受け入れ条件2)。数字だけでは、
	// 良くなったのか壊れたのかが読み取りにくい。
	bust := build(domain.BJSwitchPhaseSwitch, 23, 18, true)
	assert.Contains(t, bust, i18n.Tf("blackjackswitch.switchPreviewBust", "score", "23"))

	// ちょうど 21 はバーストでない。ここを >= で書くと、最良手が壊れて見える。
	assert.NotContains(t, build(domain.BJSwitchPhaseSwitch, 21, 18, true),
		i18n.Tf("blackjackswitch.switchPreviewBust", "score", "21"))

	// 入れ替えられない局面 (2枚未満) では出さない。
	assert.NotContains(t, build(domain.BJSwitchPhaseSwitch, 0, 0, false),
		strings.SplitN(i18n.T("blackjackswitch.switchPreviewLine"), "{{", 2)[0])
}

// ACTION / END では出さない (受け入れ条件3)。確定した得点と読み違える。
func TestBlackJackSwitchCuiPresenter_HidesThePreviewOutsideTheSwitchPhase(t *testing.T) {
	i18n.SetLang("ja")
	for _, phase := range []int{domain.BJSwitchPhaseBet, domain.BJSwitchPhaseAction, domain.BJSwitchPhaseEnd} {
		m := new(interfaces.MockBlackJackSwitchGame)
		setupBlackJackSwitchCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "SwitchPreviewScores")
		m.On("GetPhase").Return(phase)
		// **呼ばれもしないこと。**フェーズを見ずに計算してから捨てる実装を弾く。
		m.On("SwitchPreviewScores").Return(20, 18, true).Maybe()

		out := new(BlackJackSwitchCuiPresenter).Output(m, nil)
		assert.NotContains(t, out, strings.SplitN(i18n.T("blackjackswitch.switchPreviewLine"), "{{", 2)[0],
			"phase %d", phase)
		m.AssertNotCalled(t, "SwitchPreviewScores")
	}
}
