//go:build !js || !wasm || extra

package presenter

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupOichoKabuWebMockDefaults(m *interfaces.MockOichoKabuGame) {
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

func parseOichoKabuOutput(t *testing.T, jsonStr string) *controller.OichoKabuWebOutput {
	t.Helper()
	var out controller.OichoKabuWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestOichoKabuWebPresenter_Output_BetPhase(t *testing.T) {
	p := new(OichoKabuWebPresenter)
	m := new(interfaces.MockOichoKabuGame)
	setupOichoKabuWebMockDefaults(m)

	r := parseOichoKabuOutput(t, p.Output(m, nil))
	assert.Equal(t, domain.OichoKabuPhaseBet, r.Phase)
	assert.Equal(t, 1000, r.Chips)
	assert.Empty(t, r.PlayerHand)
	assert.Empty(t, r.BankerHand)
	assert.Empty(t, r.Message)
}

func TestOichoKabuWebPresenter_Output_Error(t *testing.T) {
	p := new(OichoKabuWebPresenter)
	m := new(interfaces.MockOichoKabuGame)
	setupOichoKabuWebMockDefaults(m)
	r := parseOichoKabuOutput(t, p.Output(m, errors.New("oops")))
	assert.Equal(t, "oops", r.Message)
}

func TestOichoKabuWebPresenter_Output_DrawPhaseHidesBanker(t *testing.T) {
	p := new(OichoKabuWebPresenter)
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

	out := p.Output(m, nil)
	r := parseOichoKabuOutput(t, out)
	assert.Len(t, r.PlayerHand, 2)
	assert.Empty(t, r.BankerHand, "banker hand must be hidden until the result")
	assert.Equal(t, 0, r.BankerRank, "banker rank must be hidden until the result")
	assert.Equal(t, 9, r.PlayerRank)

	// Every kabu card must render procedurally: deck="kabu", numeric label.
	assert.True(t, strings.Contains(out, `"deck":"kabu"`), "expected deck kabu in %s", out)
	assert.True(t, strings.Contains(out, `"label":"7"`), "expected numeric label in %s", out)
}

func TestOichoKabuWebPresenter_Output_EndRevealsBanker(t *testing.T) {
	p := new(OichoKabuWebPresenter)
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

	r := parseOichoKabuOutput(t, p.Output(m, nil))
	assert.Len(t, r.BankerHand, 1)
	assert.Equal(t, 8, r.BankerRank)
	assert.Equal(t, 200, r.TotalPayout)
	assert.Equal(t, "oichokabu.result.playerWins", r.MessageCode)
}

func TestOichoKabuWebPresenter_Output_EndMessages(t *testing.T) {
	for _, tt := range []struct {
		name    string
		result  domain.OichoKabuResult
		wantKey string
	}{
		{"win", domain.OichoKabuResultWin, "oichokabu.result.playerWins"},
		{"lose", domain.OichoKabuResultLose, "oichokabu.result.bankerWins"},
		{"push", domain.OichoKabuResultDraw, "oichokabu.result.push"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := new(OichoKabuWebPresenter)
			m := new(interfaces.MockOichoKabuGame)
			m.On("GetChips").Return(1000)
			m.On("GetPhase").Return(domain.OichoKabuPhaseEnd)
			m.On("GetPlayerHand").Return(([]*domain.Card)(nil))
			m.On("GetBankerHand").Return(([]*domain.Card)(nil))
			m.On("GetPlayerRank").Return(5)
			m.On("GetBankerRank").Return(5)
			m.On("GetGameEndFlag").Return(true)
			m.On("GetBet").Return(100)
			m.On("GetResult").Return(tt.result)
			m.On("GetTotalPayout").Return(0)

			r := parseOichoKabuOutput(t, p.Output(m, nil))
			assert.Equal(t, tt.wantKey, r.MessageCode)
		})
	}
}

func TestOichoKabuWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(OichoKabuWebPresenter)
	m := new(interfaces.MockOichoKabuGame)
	m.On("GetGameEndFlag").Return(false)
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}
