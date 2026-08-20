package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
	assert.Contains(t, result, "チップ: 1000")
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
	assert.Contains(t, result, "アンテ: 100")
	assert.Contains(t, result, "プレイヤー:")
	assert.Contains(t, result, "ディーラー:")
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
	assert.Contains(t, result, "ウォーベット: 100")
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
	assert.Contains(t, result, "プレイヤーの勝ち")
	assert.Contains(t, result, "合計払戻し: 200")
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
	assert.Contains(t, result, "プレイヤーの負け")
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
	assert.Contains(t, result, "プッシュ")
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

// タイの局面でチップが足りないとき、選ぶ前に分かること。
//
// Web の CasinoWarPage は insufficientChips を role="alert" で出し、War ボタンを
// disabled にしている。CUI には同じ比較が無く、"war" を打って初めてドメインの
// エラーで気づく (#5583)。
func TestCasinoWarCuiPresenter_WarnsBeforeAnUnaffordableWar(t *testing.T) {
	render := func(chips, ante int, phase int) string {
		m := new(interfaces.MockCasinoWarGame)
		setupCasinoWarCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetChips")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetAnte")
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.On("GetChips").Return(chips)
		m.On("GetAnte").Return(ante)
		m.On("GetPhase").Return(phase)
		return new(CasinoWarCuiPresenter).Output(m, nil)
	}

	warn := i18n.Tf("casinowar.warInsufficientChips", "ante", "100")

	t.Run("chips below the ante warn", func(t *testing.T) {
		out := render(50, 100, domain.CasinoWarPhaseTieDecision)
		assert.Contains(t, out, warn)
	})

	// ちょうど足りるのは足りる。ここを > で書くと、払えるのに警告が出る。
	t.Run("chips exactly equal to the ante do not warn", func(t *testing.T) {
		assert.NotContains(t, render(100, 100, domain.CasinoWarPhaseTieDecision), warn)
	})

	t.Run("ample chips do not warn", func(t *testing.T) {
		assert.NotContains(t, render(900, 100, domain.CasinoWarPhaseTieDecision), warn)
	})

	// タイ以外の局面ではウォーを選べないので、警告する相手がいない。
	t.Run("other phases stay silent even when chips are short", func(t *testing.T) {
		for _, ph := range []int{
			domain.CasinoWarPhaseBet,
			domain.CasinoWarPhaseWarDealt,
			domain.CasinoWarPhaseEnd,
		} {
			assert.NotContains(t, render(50, 100, ph), warn, "phase %v", ph)
		}
	})
}
