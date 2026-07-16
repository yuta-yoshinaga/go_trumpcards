//go:build !js || !wasm || extra

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupOichoKabuCuiMockDefaults(m *interfaces.MockOichoKabuGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.OichoKabuPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetBankerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetPlayerRank").Return(0).Maybe()
	m.On("GetBankerRank").Return(0).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.OichoKabuResult(0)).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestOichoKabuCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	m := new(interfaces.MockOichoKabuGame)
	setupOichoKabuCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "チップ: 1000")
	assert.Contains(t, result, "BET")
	assert.Contains(t, result, "伏せ札")
	// Betting hint spells out the ceiling (current chips).
	assert.Contains(t, result, i18n.Tf("oichokabu.maxBetHint", "max", "1000"))
}

func TestOichoKabuCuiPresenter_Output_Error(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	m := new(interfaces.MockOichoKabuGame)
	setupOichoKabuCuiMockDefaults(m)
	result := p.Output(m, errors.New("oops"))
	assert.Contains(t, result, "oops")
}

func TestOichoKabuCuiPresenter_Output_DrawPhase(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	m := new(interfaces.MockOichoKabuGame)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.OichoKabuPhaseDraw)
	m.On("GetPlayerHand").Return([]*domain.Card{domain.NewCard(1, 7, true), domain.NewCard(2, 2, true)})
	m.On("GetBankerHand").Return([]*domain.Card{domain.NewCard(3, 5, true), domain.NewCard(4, 4, true)})
	m.On("GetPlayerRank").Return(9)
	m.On("GetBankerRank").Return(9)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetBet").Return(100)
	m.On("GetResult").Return(domain.OichoKabuResult(0))
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "掛け金: 100")
	assert.Contains(t, result, "手札: 7,2")
	assert.Contains(t, result, "伏せ札", "banker hand hidden during draw")
	// Draw hint spells out the draw/stand choice.
	assert.Contains(t, result, i18n.T("oichokabu.drawHint"))
}

func TestOichoKabuCuiPresenter_Output_EndWin(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	m := new(interfaces.MockOichoKabuGame)
	m.On("GetChips").Return(1100)
	m.On("GetPhase").Return(domain.OichoKabuPhaseEnd)
	m.On("GetPlayerHand").Return([]*domain.Card{domain.NewCard(1, 9, true)})
	m.On("GetBankerHand").Return([]*domain.Card{domain.NewCard(2, 8, true)})
	m.On("GetPlayerRank").Return(9)
	m.On("GetBankerRank").Return(8)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBet").Return(100)
	m.On("GetResult").Return(domain.OichoKabuResultWin)
	m.On("GetTotalPayout").Return(200)

	result := p.Output(m, nil)
	assert.Contains(t, result, "子の勝ち")
	assert.Contains(t, result, "手札: 8") // banker revealed
	assert.Contains(t, result, "合計払戻し: 200")
}

func TestOichoKabuCuiPresenter_Output_EndLose(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	m := new(interfaces.MockOichoKabuGame)
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.OichoKabuPhaseEnd)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetBankerHand").Return(([]*domain.Card)(nil))
	m.On("GetPlayerRank").Return(7)
	m.On("GetBankerRank").Return(8)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBet").Return(100)
	m.On("GetResult").Return(domain.OichoKabuResultLose)
	m.On("GetTotalPayout").Return(0)
	result := p.Output(m, nil)
	assert.Contains(t, result, "親の勝ち")
}

func TestOichoKabuCuiPresenter_Output_EndPush(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	m := new(interfaces.MockOichoKabuGame)
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.OichoKabuPhaseEnd)
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
	m.On("GetBankerHand").Return(([]*domain.Card)(nil))
	m.On("GetPlayerRank").Return(5)
	m.On("GetBankerRank").Return(5)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetBet").Return(100)
	m.On("GetResult").Return(domain.OichoKabuResultDraw)
	m.On("GetTotalPayout").Return(100)
	result := p.Output(m, nil)
	assert.Contains(t, result, "プッシュ")
}

func TestOichoKabuCuiPresenter_PhaseStr_AllBranches(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	for phase, expect := range map[int]string{
		domain.OichoKabuPhaseBet:  "BET",
		domain.OichoKabuPhaseDraw: "DRAW",
		domain.OichoKabuPhaseEnd:  "END",
		999:                       "UNKNOWN",
	} {
		assert.Equal(t, expect, p.phaseStr(phase))
	}
}

func TestOichoKabuCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(OichoKabuCuiPresenter)
	m := new(interfaces.MockOichoKabuGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
