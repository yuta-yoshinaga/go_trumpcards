//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupYanivCuiMock() (*interfaces.MockYanivGame, []*domain.YanivPlayer) {
	m := new(interfaces.MockYanivGame)
	players := makeYanivPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(39)
	m.On("GetPickupCards").Return([]*domain.Card{})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.YanivPhaseDiscard)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetCallerIdx").Return(-1)
	m.On("GetIsAsaf").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetConfig").Return(domain.DefaultYanivConfig())
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestYanivCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.YanivCuiPresenter)

	t.Run("discard phase shows hand and commands", func(t *testing.T) {
		m, players := setupYanivCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "ヤニブ")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "山札: 39")
		assert.Contains(t, result, "ディスカードフェーズ")
		assert.Contains(t, result, "[0]SPADE 9")
		// Header shows the configured score limit; no one is near out yet.
		assert.Contains(t, result, strings.Split(i18n.T("yaniv.limitLine"), "{{")[0])
		assert.NotContains(t, result, i18n.T("yaniv.nearOut"))
	})

	t.Run("player near the score limit is warned", func(t *testing.T) {
		m, players := setupYanivCuiMock()
		// Default limit is 200 → 80% = 160; a score above that flags near-out.
		players[1].SetScore(180)
		result := p.Output(m, nil)
		assert.Contains(t, result, i18n.T("yaniv.nearOut"))
	})

	t.Run("yaniv help shown when hand low", func(t *testing.T) {
		m, players := setupYanivCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false)) // total 2 <= 5
		result := p.Output(m, nil)
		assert.Contains(t, result, "Yaniv")
	})

	t.Run("pickup line shown", func(t *testing.T) {
		m, _ := setupYanivCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPickupCards")
		m.On("GetPickupCards").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 7, false),
		})
		assert.Contains(t, p.Output(m, nil), "HEART 7")
	})

	t.Run("draw phase commands", func(t *testing.T) {
		m, _ := setupYanivCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.YanivPhaseDraw)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ドローフェーズ")
		assert.Contains(t, result, "ds")
		assert.Contains(t, result, "dp")
	})

	t.Run("round end yaniv result", func(t *testing.T) {
		m, _ := setupYanivCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCallerIdx")
		m.On("GetPhase").Return(domain.YanivPhaseRoundEnd)
		m.On("GetCallerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("round end asaf result", func(t *testing.T) {
		m, _ := setupYanivCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCallerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsAsaf")
		m.On("GetPhase").Return(domain.YanivPhaseRoundEnd)
		m.On("GetCallerIdx").Return(0)
		m.On("GetIsAsaf").Return(true)
		assert.Contains(t, p.Output(m, nil), "アサフ")
	})

	t.Run("eliminated player shown OUT", func(t *testing.T) {
		m, players := setupYanivCuiMock()
		players[3].SetEliminated(true)
		assert.Contains(t, p.Output(m, nil), "OUT")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupYanivCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid play")), "invalid play")
	})

	t.Run("game ended human winner", func(t *testing.T) {
		m, _ := setupYanivCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		assert.Contains(t, p.Output(m, nil), "ゲーム終了")
	})
}

func TestYanivCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.YanivCuiPresenter)

	m := new(interfaces.MockYanivGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "yaniv", Detail: "You call Yaniv"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	assert.Contains(t, p.ActionLogOutput(m), "yaniv")
}
