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

func setupMississippiStudWebMockDefaults(m *interfaces.MockMississippiStudGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.MississippiStudPhaseAnte).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteAmount").Return(0).Maybe()
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{}).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetTotalBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetHandRank").Return(0).Maybe()
	m.On("GetPayoutMultiplier").Return(0).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{}).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseMississippiStudOutput(t *testing.T, jsonStr string) *controller.MississippiStudWebOutput {
	t.Helper()
	var out controller.MississippiStudWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestMississippiStudWebPresenter_Output_AntePhase(t *testing.T) {
	p := new(MississippiStudWebPresenter)
	m := new(interfaces.MockMississippiStudGame)
	setupMississippiStudWebMockDefaults(m)

	result := parseMississippiStudOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.MississippiStudPhaseAnte, result.Phase)
	assert.Equal(t, 1000, result.Chips)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.CommunityCards)
	assert.Empty(t, result.Message)
}

func TestMississippiStudWebPresenter_Output_Error(t *testing.T) {
	p := new(MississippiStudWebPresenter)
	m := new(interfaces.MockMississippiStudGame)
	setupMississippiStudWebMockDefaults(m)

	result := parseMississippiStudOutput(t, p.Output(m, errors.New("oops")))
	assert.Equal(t, "oops", result.Message)
}

func TestMississippiStudWebPresenter_Output_ThirdSt_MasksCommunity(t *testing.T) {
	p := new(MississippiStudWebPresenter)
	m := new(interfaces.MockMississippiStudGame)

	hole := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, true),
		domain.NewCard(domain.CardDesignHeart, 11, true),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
	}
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.MississippiStudPhaseThirdSt)
	m.On("GetPlayerHand").Return(hole)
	m.On("GetCommunityCards").Return(community)
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetAnteAmount").Return(100)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetFolded").Return(false)
	m.On("GetTotalBet").Return(100)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetHandRank").Return(0)
	m.On("GetPayoutMultiplier").Return(0)
	m.On("GetAntePayout").Return(0)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetTotalPayout").Return(0)

	result := parseMississippiStudOutput(t, p.Output(m, nil))
	assert.Len(t, result.PlayerHand, 2)
	require := assert.New(t)
	require.Len(result.CommunityCards, 3)
	for _, c := range result.CommunityCards {
		require.Equal("", c.Design)
		require.Equal(0, c.Value)
	}
}

func TestMississippiStudWebPresenter_Output_EndWin(t *testing.T) {
	p := new(MississippiStudWebPresenter)
	m := new(interfaces.MockMississippiStudGame)

	hole := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, true),
		domain.NewCard(domain.CardDesignHeart, 11, true),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
	}
	m.On("GetChips").Return(1600)
	m.On("GetPhase").Return(domain.MississippiStudPhaseEnd)
	m.On("GetPlayerHand").Return(hole)
	m.On("GetCommunityCards").Return(community)
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{true, true, true})
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnteAmount").Return(100)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{3, 1, 1})
	m.On("GetFolded").Return(false)
	m.On("GetTotalBet").Return(600)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetHandRank").Return(domain.PokerHandOnePair)
	m.On("GetPayoutMultiplier").Return(domain.MississippiStudPayHighPair)
	m.On("GetAntePayout").Return(200)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{600, 200, 200})
	m.On("GetTotalPayout").Return(1200)

	result := parseMississippiStudOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.MississippiStudPhaseEnd, result.Phase)
	assert.Equal(t, "Player wins!", result.Message)
	assert.Equal(t, "mississippistud.result.playerWins", result.MessageCode)
	assert.Equal(t, 1200, result.TotalPayout)
	assert.Equal(t, []int{3, 1, 1}, result.StreetMultipliers)
	assert.Equal(t, []bool{true, true, true}, result.CommunityRevealed)
}

func TestMississippiStudWebPresenter_Output_EndPush(t *testing.T) {
	p := new(MississippiStudWebPresenter)
	m := new(interfaces.MockMississippiStudGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.MississippiStudPhaseEnd)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil))
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{})
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnteAmount").Return(0)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetFolded").Return(false)
	m.On("GetTotalBet").Return(0)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetHandRank").Return(0)
	m.On("GetPayoutMultiplier").Return(0)
	m.On("GetAntePayout").Return(0)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetTotalPayout").Return(0)

	result := parseMississippiStudOutput(t, p.Output(m, nil))
	assert.Equal(t, "Push.", result.Message)
	assert.Equal(t, "mississippistud.result.push", result.MessageCode)
}

func TestMississippiStudWebPresenter_Output_EndLose(t *testing.T) {
	p := new(MississippiStudWebPresenter)
	m := new(interfaces.MockMississippiStudGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.MississippiStudPhaseEnd)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil))
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{})
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnteAmount").Return(0)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetFolded").Return(true)
	m.On("GetTotalBet").Return(100)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetHandRank").Return(0)
	m.On("GetPayoutMultiplier").Return(0)
	m.On("GetAntePayout").Return(0)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetTotalPayout").Return(0)

	result := parseMississippiStudOutput(t, p.Output(m, nil))
	assert.Equal(t, "Player loses.", result.Message)
	assert.Equal(t, "mississippistud.result.playerLoses", result.MessageCode)
}

func TestMississippiStudWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(MississippiStudWebPresenter)
	m := new(interfaces.MockMississippiStudGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetGameEndFlag").Return(true).Maybe()

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
