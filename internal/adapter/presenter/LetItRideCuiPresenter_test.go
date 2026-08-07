package presenter

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
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
