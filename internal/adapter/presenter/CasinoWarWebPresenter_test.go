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

func setupCasinoWarWebMockDefaults(m *interfaces.MockCasinoWarGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.CasinoWarPhaseBet).Maybe()
	m.On("GetPlayerCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetDealerCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetPlayerWarCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetDealerWarCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetBurnCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnte").Return(0).Maybe()
	m.On("GetWarBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func parseCasinoWarOutput(t *testing.T, jsonStr string) *controller.CasinoWarWebOutput {
	t.Helper()
	var out controller.CasinoWarWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestCasinoWarWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	setupCasinoWarWebMockDefaults(m)

	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.CasinoWarPhaseBet, r.Phase)
	assert.Equal(t, 1000, r.Chips)
	assert.Nil(t, r.PlayerCard)
	assert.Nil(t, r.DealerCard)
	assert.Empty(t, r.BurnCards)
	assert.Empty(t, r.Message)
}

func TestCasinoWarWebPresenter_Output_Error(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	setupCasinoWarWebMockDefaults(m)
	r := parseCasinoWarOutput(t, p.Output(m, errors.New("oops")))
	assert.Equal(t, "oops", r.Message)
}

func TestCasinoWarWebPresenter_Output_PlayerWins(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	pc := domain.NewCard(domain.CardDesignSpade, 13, false)
	dc := domain.NewCard(domain.CardDesignClover, 7, false)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.CasinoWarPhaseEnd)
	m.On("GetPlayerCard").Return(pc)
	m.On("GetDealerCard").Return(dc)
	m.On("GetPlayerWarCard").Return((*domain.Card)(nil))
	m.On("GetDealerWarCard").Return((*domain.Card)(nil))
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(0)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetTotalPayout").Return(200)

	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Empty(t, r.Message)
	assert.Equal(t, "casinowar.result.playerWins", r.MessageCode)
	assert.NotNil(t, r.PlayerCard)
	assert.NotNil(t, r.DealerCard)
	assert.Equal(t, 200, r.TotalPayout)
}

func TestCasinoWarWebPresenter_Output_DealerWins(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	setupCasinoWarWebMockDefaults(m)
	m.ExpectedCalls = nil
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.CasinoWarPhaseEnd)
	m.On("GetPlayerCard").Return((*domain.Card)(nil))
	m.On("GetDealerCard").Return((*domain.Card)(nil))
	m.On("GetPlayerWarCard").Return((*domain.Card)(nil))
	m.On("GetDealerWarCard").Return((*domain.Card)(nil))
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(0)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetTotalPayout").Return(0)
	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Empty(t, r.Message)
	assert.Equal(t, "casinowar.result.dealerWins", r.MessageCode)
}

func TestCasinoWarWebPresenter_Output_Surrender(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	setupCasinoWarWebMockDefaults(m)
	m.ExpectedCalls = nil
	m.On("GetChips").Return(950)
	m.On("GetPhase").Return(domain.CasinoWarPhaseEnd)
	m.On("GetPlayerCard").Return((*domain.Card)(nil))
	m.On("GetDealerCard").Return((*domain.Card)(nil))
	m.On("GetPlayerWarCard").Return((*domain.Card)(nil))
	m.On("GetDealerWarCard").Return((*domain.Card)(nil))
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(0)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetTotalPayout").Return(50)
	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Empty(t, r.Message)
	assert.Equal(t, "casinowar.result.surrender", r.MessageCode)
}

func TestCasinoWarWebPresenter_Output_WarWin(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	pw := domain.NewCard(domain.CardDesignHeart, 13, false)
	dw := domain.NewCard(domain.CardDesignDiamond, 5, false)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.CasinoWarPhaseEnd)
	m.On("GetPlayerCard").Return((*domain.Card)(nil))
	m.On("GetDealerCard").Return((*domain.Card)(nil))
	m.On("GetPlayerWarCard").Return(pw)
	m.On("GetDealerWarCard").Return(dw)
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(100)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetTotalPayout").Return(300)
	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Equal(t, "casinowar.result.warWin", r.MessageCode)
}

func TestCasinoWarWebPresenter_Output_WarTieWin(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	pw := domain.NewCard(domain.CardDesignHeart, 9, false)
	dw := domain.NewCard(domain.CardDesignDiamond, 9, false)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.CasinoWarPhaseEnd)
	m.On("GetPlayerCard").Return((*domain.Card)(nil))
	m.On("GetDealerCard").Return((*domain.Card)(nil))
	m.On("GetPlayerWarCard").Return(pw)
	m.On("GetDealerWarCard").Return(dw)
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(100)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetTotalPayout").Return(300)
	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Equal(t, "casinowar.result.warTieWin", r.MessageCode)
}

func TestCasinoWarWebPresenter_Output_WarLoss(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	pw := domain.NewCard(domain.CardDesignHeart, 5, false)
	dw := domain.NewCard(domain.CardDesignDiamond, 13, false)
	m.On("GetChips").Return(800)
	m.On("GetPhase").Return(domain.CasinoWarPhaseEnd)
	m.On("GetPlayerCard").Return((*domain.Card)(nil))
	m.On("GetDealerCard").Return((*domain.Card)(nil))
	m.On("GetPlayerWarCard").Return(pw)
	m.On("GetDealerWarCard").Return(dw)
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(100)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetTotalPayout").Return(0)
	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Equal(t, "casinowar.result.warLoss", r.MessageCode)
}

// TestCasinoWarWebPresenter_Output_NonWinFallsBackToLose は、Casino War では発生し得ない
// GameResultDraw が万一返ってきた場合に dealerWins メッセージへフォールバックすることを保証する。
func TestCasinoWarWebPresenter_Output_NonWinFallsBackToLose(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.CasinoWarPhaseEnd)
	m.On("GetPlayerCard").Return((*domain.Card)(nil))
	m.On("GetDealerCard").Return((*domain.Card)(nil))
	m.On("GetPlayerWarCard").Return((*domain.Card)(nil))
	m.On("GetDealerWarCard").Return((*domain.Card)(nil))
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(0)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetTotalPayout").Return(0)
	r := parseCasinoWarOutput(t, p.Output(m, nil))
	assert.Equal(t, "casinowar.result.dealerWins", r.MessageCode)
}

func TestCasinoWarWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(CasinoWarWebPresenter)
	m := new(interfaces.MockCasinoWarGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
