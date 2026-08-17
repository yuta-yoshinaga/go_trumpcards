package presenter

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupLetItRideCuiMockDefaults(m *interfaces.MockLetItRideGame) {
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

func TestLetItRideCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)
	setupLetItRideCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "BET")
}

func TestLetItRideCuiPresenter_Output_Error(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)
	setupLetItRideCuiMockDefaults(m)

	result := p.Output(m, errors.New("test error"))
	assert.Contains(t, result, "test error")
}

func TestLetItRideCuiPresenter_Output_FirstDecision(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)

	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
	}
	m.On("GetChips").Return(700)
	m.On("GetPhase").Return(domain.LetItRidePhaseFirstDecision)
	m.On("GetPlayerHand").Return(cards)
	m.On("GetCommunityCards").Return(community)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetBetAmount").Return(100)
	m.On("GetBet1Active").Return(true)
	m.On("GetBet2Active").Return(true)
	m.On("GetBet3Active").Return(true)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetHandRank").Return(0)
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "FIRST DECISION")
	assert.Contains(t, result, "PLAYER")
	assert.Contains(t, result, "COMMUNITY")
	assert.Contains(t, result, "??")
}

func TestLetItRideCuiPresenter_Output_SecondDecision(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)

	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
	}
	m.On("GetChips").Return(800)
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
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "SECOND DECISION")
	// First community card should be shown, second masked
	lines := strings.Split(result, "\n")
	communityFound := false
	for _, line := range lines {
		if strings.Contains(line, "??") && !strings.HasPrefix(line, "---") {
			communityFound = true
		}
	}
	assert.True(t, communityFound)
}

func TestLetItRideCuiPresenter_Output_EndPhase_Win(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)

	m.On("GetChips").Return(1600)
	m.On("GetPhase").Return(domain.LetItRidePhaseEnd)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBetAmount").Return(100)
	m.On("GetBet1Active").Return(true)
	m.On("GetBet2Active").Return(true)
	m.On("GetBet3Active").Return(true)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetHandRank").Return(domain.PokerHandTwoPair)
	m.On("GetTotalPayout").Return(900)

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーの勝ち！")
	assert.Contains(t, result, "合計払戻し: 900")
}

func TestLetItRideCuiPresenter_Output_EndPhase_Loss(t *testing.T) {
	p := new(LetItRideCuiPresenter)
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
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "プレイヤーの負け。")
}

func TestLetItRideCuiPresenter_Output_UnknownPhase(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)
	setupLetItRideCuiMockDefaults(m)
	m.On("GetPhase").Return(99).Unset()
	m.On("GetPhase").Return(99)
	m.On("GetCommunityCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}).Unset()
	m.On("GetCommunityCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})

	result := p.Output(m, nil)
	assert.Contains(t, result, "UNKNOWN")
}

func TestLetItRideCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetGameEndFlag").Return(false)

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

func TestLetItRideCuiPresenter_PullConfirmOutput(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	withPreview := func(pv *domain.LetItRidePullPreview) *interfaces.MockLetItRideGame {
		m := new(interfaces.MockLetItRideGame)
		m.On("GetPullPreview").Return(pv)
		return m
	}

	// **金額が出ないと確認する意味がない。**取り消せない操作の前に、戻る額と
	// 場に残る額の前後を見せる (#4699)。
	t.Run("names the amount returned and the stake before and after", func(t *testing.T) {
		out := p.PullConfirmOutput(withPreview(&domain.LetItRidePullPreview{
			Returned: 100, RiskBefore: 300, RiskAfter: 200,
		}))
		assert.Contains(t, out, "100")
		assert.Contains(t, out, "300")
		assert.Contains(t, out, "200")
	})

	t.Run("says the action cannot be undone and how to go ahead", func(t *testing.T) {
		out := p.PullConfirmOutput(withPreview(&domain.LetItRidePullPreview{
			Returned: 100, RiskBefore: 300, RiskAfter: 200,
		}))
		assert.Contains(t, out, "取り消せません")
		assert.Contains(t, out, "y")
	})

	t.Run("reports when pulling is not possible", func(t *testing.T) {
		assert.Contains(t, p.PullConfirmOutput(withPreview(nil)), "引き下げられません")
	})
}

// #5537: Web は「いまリスクにさらされている総額」を常設で出しているのに、
// CUI は 1 口の額と口数しか出さず、判断のたびに暗算が要った。
func TestLetItRideCuiPresenter_Output_TotalRisk(t *testing.T) {
	p := new(LetItRideCuiPresenter)

	build := func(phase int, active [3]bool) *interfaces.MockLetItRideGame {
		m := new(interfaces.MockLetItRideGame)
		m.On("GetPhase").Return(phase)
		m.On("GetBetAmount").Return(100)
		m.On("GetBet1Active").Return(active[0])
		m.On("GetBet2Active").Return(active[1])
		m.On("GetBet3Active").Return(active[2])
		setupLetItRideCuiMockDefaults(m)
		return m
	}
	riskLine := func(total int) string {
		return i18n.Tf("letitride.totalRiskLine", "risk", strconv.Itoa(total))
	}

	// 3口すべて有効 = 300。
	assert.Contains(t, p.Output(build(domain.LetItRidePhaseFirstDecision, [3]bool{true, true, true}), nil), riskLine(300))
	// 1口引いた後 = 200。
	assert.Contains(t, p.Output(build(domain.LetItRidePhaseSecondDecision, [3]bool{false, true, true}), nil), riskLine(200))
	assert.Contains(t, p.Output(build(domain.LetItRidePhaseEnd, [3]bool{false, false, true}), nil), riskLine(100))

	// **BET フェーズでは出さない。**まだ賭けが確定していない。
	// i18n.T() で比べると {{risk}} のままの文字列と比べることになり、
	// 何を出していても通ってしまう。実際に出るはずの行で比べる。
	betOut := p.Output(build(domain.LetItRidePhaseBet, [3]bool{true, true, true}), nil)
	assert.NotContains(t, betOut, riskLine(300))
	assert.NotContains(t, betOut, strings.SplitN(i18n.T("letitride.totalRiskLine"), "{{", 2)[0])
}

// **Pull 確認画面の RiskBefore と同じ値であること。**二重に計算していると、
// 同じ局面で違う総額を見せることになる。
func TestLetItRideCuiPresenter_TotalRiskMatchesThePullPreview(t *testing.T) {
	p := new(LetItRideCuiPresenter)
	m := new(interfaces.MockLetItRideGame)
	m.On("GetPhase").Return(domain.LetItRidePhaseFirstDecision)
	m.On("GetBetAmount").Return(50)
	m.On("GetBet1Active").Return(true)
	m.On("GetBet2Active").Return(true)
	m.On("GetBet3Active").Return(false)
	m.On("GetPullPreview").Return(&domain.LetItRidePullPreview{Returned: 50, RiskBefore: 100, RiskAfter: 50})
	setupLetItRideCuiMockDefaults(m)

	assert.Contains(t, p.Output(m, nil), i18n.Tf("letitride.totalRiskLine", "risk", "100"))
	assert.Contains(t, p.PullConfirmOutput(m), "100")
}
