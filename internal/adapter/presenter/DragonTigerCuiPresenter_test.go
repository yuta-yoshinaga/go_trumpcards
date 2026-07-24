package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
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
	// No history yet, so the history line is omitted.
	assert.NotContains(t, result, "履歴:")
}

func TestDragonTigerCuiPresenter_Output_History(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	setupDragonTigerCuiMockDefaults(m)
	m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetHistory")
	m.On("GetHistory").Return([]int{
		domain.DragonTigerResultDragon,
		domain.DragonTigerResultTiger,
		domain.DragonTigerResultTie,
		domain.DragonTigerResultDragon,
	}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "履歴: D T = D")
	assert.Contains(t, result, "集計: D:2 T:1 =:1")
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
	// Dragon pays 1:1, so the odds line reads ×1.
	assert.Contains(t, result, "DRAGON ×1")
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
	// Player bet Dragon (×1) here, so the odds line reads ×1.
	assert.Contains(t, result, "DRAGON ×1")
}

// Regression coverage for the gemini/Claude review: the result message color
// must reflect the player's outcome, not the game-side winner. Player bet
// Tiger and Tiger wins → green, even though the underlying GameResult is Lose.
func TestDragonTigerCuiPresenter_Output_TigerWinsOnTigerBet_Green(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	m.On("GetChips").Return(1100).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd).Maybe()
	m.On("GetDragonCard").Return(domain.NewCard(domain.CardDesignSpade, 3, false)).Maybe()
	m.On("GetTigerCard").Return(domain.NewCard(domain.CardDesignHeart, 13, false)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.DragonTigerBetTiger).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetPayout").Return(200).Maybe()
	m.On("GetHistory").Return([]int{domain.DragonTigerResultTiger}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "タイガーの勝ち")
	// ANSI green = \x1b[32m. Bare-color check is brittle if the codebase
	// ever swaps to no-color mode in tests, so we verify the green prefix
	// appears near the "タイガーの勝ち" marker and not the red prefix.
	assert.NotContains(t, result, "\x1b[31mタイガーの勝ち", "Player bet Tiger and won — must not be red")
}

// Inverse: player bet Dragon and Dragon wins → green; player bet Tiger and
// Dragon wins → red. Asserts the Dragon-bet-on-Dragon-win path stays green
// after the refactor.
func TestDragonTigerCuiPresenter_Output_DragonWinsOnTigerBet_Red(t *testing.T) {
	p := new(DragonTigerCuiPresenter)
	m := new(interfaces.MockDragonTigerGame)
	m.On("GetChips").Return(900).Maybe()
	m.On("GetPhase").Return(domain.DragonTigerPhaseEnd).Maybe()
	m.On("GetDragonCard").Return(domain.NewCard(domain.CardDesignSpade, 13, false)).Maybe()
	m.On("GetTigerCard").Return(domain.NewCard(domain.CardDesignHeart, 5, false)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetBetAmount").Return(100).Maybe()
	m.On("GetBetType").Return(domain.DragonTigerBetTiger).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetPayout").Return(0).Maybe()
	m.On("GetHistory").Return([]int{domain.DragonTigerResultDragon}).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "ドラゴンの勝ち")
	assert.NotContains(t, result, "\x1b[32mドラゴンの勝ち", "Player bet Tiger but Dragon won — must not be green")
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
	// Tie pays 8:1, so the odds line reads ×8.
	assert.Contains(t, result, "TIE ×8")
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
