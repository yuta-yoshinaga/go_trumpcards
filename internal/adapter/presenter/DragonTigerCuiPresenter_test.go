package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupDragonTigerCuiMockDefaults(m *interfaces.MockDragonTigerGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseBet).Maybe()
	m.On("GetDragonCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetTigerCard").Return((*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBetAmount").Return(0).Maybe()
	m.On("GetBetType").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHistory").Return(([]int)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestDragonTigerCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	setupDragonTigerCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "フェーズ: BET")
}

func TestDragonTigerCuiPresenter_Output_DragonWins(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	m.On("GetChips").Return(1100).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd).Maybe()
	m.On("GetDragonCard").Return(domain.NewCard(domain.CardDesignSpade, 13, false)).Maybe()
	m.On("GetTigerCard").Return(domain.NewCard(domain.CardDesignHeart, 5, false)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.DragonTigerBetDragon).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(200).Maybe()
	m.On("GetHistory").Return([]int{domain.DragonTigerResultDragon}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ドラゴン")
	assert.Contains(t, result, "タイガー")
	assert.Contains(t, result, "ドラゴンの勝ち")
	assert.Contains(t, result, "払戻し: 200")
}

func TestDragonTigerCuiPresenter_Output_TigerWins(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd).Maybe()
	m.On("GetDragonCard").Return(domain.NewCard(domain.CardDesignSpade, 3, false)).Maybe()
	m.On("GetTigerCard").Return(domain.NewCard(domain.CardDesignHeart, 13, false)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.DragonTigerBetDragon).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHistory").Return([]int{domain.DragonTigerResultTiger}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "タイガーの勝ち")
}

func TestDragonTigerCuiPresenter_Output_Tie_RefundOnDragonBet(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	m.On("GetChips").Return(950).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd).Maybe()
	m.On("GetDragonCard").Return(domain.NewCard(domain.CardDesignSpade, 7, false)).Maybe()
	m.On("GetTigerCard").Return(domain.NewCard(domain.CardDesignHeart, 7, false)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.DragonTigerBetDragon).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetPayout").Return(50).Maybe()
	m.On("GetHistory").Return([]int{domain.DragonTigerResultTie}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "タイ。ベット額の半分を返還")
	assert.Contains(t, result, "払戻し: 50")
}

func TestDragonTigerCuiPresenter_Output_Tie_TieBetWins(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	m.On("GetChips").Return(1800).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd).Maybe()
	m.On("GetDragonCard").Return(domain.NewCard(domain.CardDesignSpade, 7, false)).Maybe()
	m.On("GetTigerCard").Return(domain.NewCard(domain.CardDesignHeart, 7, false)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.DragonTigerBetTie).Maybe()
	m.On("GetResult").Return(domain.GameResultDraw).Maybe()
	m.On("GetPayout").Return(900).Maybe()
	m.On("GetHistory").Return([]int{domain.DragonTigerResultTie}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "タイベット的中")
	assert.Contains(t, result, "払戻し: 900")
}

func TestDragonTigerCuiPresenter_Output_Error(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	setupDragonTigerCuiMockDefaults(m)
	result := p.Output(m, errors.New("invalid bet"))
	assert.Contains(t, result, "invalid bet")
}

func TestDragonTigerCuiPresenter_PhaseStr(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	assert.Equal(t, "BET", p.phaseStr(domain.DragonTigerPhaseBet))
	assert.Equal(t, "END", p.phaseStr(domain.DragonTigerPhaseEnd))
	assert.Equal(t, "UNKNOWN", p.phaseStr(99))
}

func TestDragonTigerCuiPresenter_BetTypeStr(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	assert.Equal(t, "DRAGON", p.betTypeStr(domain.DragonTigerBetDragon))
	assert.Equal(t, "TIGER", p.betTypeStr(domain.DragonTigerBetTiger))
	assert.Equal(t, "TIE", p.betTypeStr(domain.DragonTigerBetTie))
	assert.Equal(t, "UNKNOWN", p.betTypeStr(99))
}

func TestDragonTigerCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
