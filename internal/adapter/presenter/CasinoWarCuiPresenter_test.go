package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCasinoWarCuiMockDefaults(m *interfaces.MockCasinoWarGame) {
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

func TestCasinoWarCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	m := new(interfaces.MockCasinoWarGame)
	setupCasinoWarCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "chips: 1000")
	assert.Contains(t, result, "BET")
}

func TestCasinoWarCuiPresenter_Output_Error(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	m := new(interfaces.MockCasinoWarGame)
	setupCasinoWarCuiMockDefaults(m)
	result := p.Output(m, errors.New("oops"))
	assert.Contains(t, result, "oops")
}

func TestCasinoWarCuiPresenter_Output_TieDecision(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	m := new(interfaces.MockCasinoWarGame)
	pc := domain.NewCard(domain.CardDesignSpade, 7, false)
	dc := domain.NewCard(domain.CardDesignHeart, 7, false)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.CasinoWarPhaseTieDecision)
	m.On("GetPlayerCard").Return(pc)
	m.On("GetDealerCard").Return(dc)
	m.On("GetPlayerWarCard").Return((*domain.Card)(nil))
	m.On("GetDealerWarCard").Return((*domain.Card)(nil))
	m.On("GetBurnCards").Return(([]*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(0)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "TIE DECISION")
	assert.Contains(t, result, "INITIAL")
	assert.Contains(t, result, "ante: 100")
	assert.Contains(t, result, "player:")
	assert.Contains(t, result, "dealer:")
}

func TestCasinoWarCuiPresenter_Output_WarDealt(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	m := new(interfaces.MockCasinoWarGame)
	pc := domain.NewCard(domain.CardDesignSpade, 7, false)
	dc := domain.NewCard(domain.CardDesignHeart, 7, false)
	pw := domain.NewCard(domain.CardDesignClover, 13, false)
	dw := domain.NewCard(domain.CardDesignDiamond, 5, false)
	burn := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
		domain.NewCard(domain.CardDesignSpade, 4, false),
	}
	m.On("GetChips").Return(800)
	m.On("GetPhase").Return(domain.CasinoWarPhaseWarDealt)
	m.On("GetPlayerCard").Return(pc)
	m.On("GetDealerCard").Return(dc)
	m.On("GetPlayerWarCard").Return(pw)
	m.On("GetDealerWarCard").Return(dw)
	m.On("GetBurnCards").Return(burn)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetAnte").Return(100)
	m.On("GetWarBet").Return(100)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "WAR DEALT")
	assert.Contains(t, result, "BURN")
	assert.Contains(t, result, "WAR")
	assert.Contains(t, result, "warBet: 100")
}

func TestCasinoWarCuiPresenter_Output_EndWin(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	m := new(interfaces.MockCasinoWarGame)
	pc := domain.NewCard(domain.CardDesignSpade, 13, false)
	dc := domain.NewCard(domain.CardDesignHeart, 7, false)
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

	result := p.Output(m, nil)
	assert.Contains(t, result, "Player wins")
	assert.Contains(t, result, "total payout: 200")
}

func TestCasinoWarCuiPresenter_Output_EndLose(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	m := new(interfaces.MockCasinoWarGame)
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
	result := p.Output(m, nil)
	assert.Contains(t, result, "Player loses")
}

func TestCasinoWarCuiPresenter_Output_EndPush(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
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
	m.On("GetTotalPayout").Return(100)
	result := p.Output(m, nil)
	assert.Contains(t, result, "Push")
}

func TestCasinoWarCuiPresenter_PhaseStr_AllBranches(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	for phase, expect := range map[int]string{
		domain.CasinoWarPhaseBet:          "BET",
		domain.CasinoWarPhaseInitialDealt: "INITIAL DEALT",
		domain.CasinoWarPhaseTieDecision:  "TIE DECISION",
		domain.CasinoWarPhaseWarDealt:     "WAR DEALT",
		domain.CasinoWarPhaseEnd:          "END",
		999:                               "UNKNOWN",
	} {
		assert.Equal(t, expect, p.phaseStr(phase))
	}
}

func TestCasinoWarCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(CasinoWarCuiPresenter)
	m := new(interfaces.MockCasinoWarGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
