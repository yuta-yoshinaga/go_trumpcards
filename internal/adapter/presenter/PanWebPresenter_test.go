//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupPanWebMock() (*interfaces.MockPanGame, []*domain.PanPlayer) {
	m := new(interfaces.MockPanGame)
	players := []*domain.PanPlayer{
		domain.NewPanPlayer(true),
		domain.NewPanPlayer(false),
		domain.NewPanPlayer(false),
	}
	m.On("GetPhase").Return(domain.PanPhaseDraw)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTargetRounds").Return(3)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDrawPileCount").Return(279)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetPanDeclarerIdx").Return(-1)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetConfig").Return(domain.DefaultPanConfig())
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("PlayerHandPoints", 0).Return(5)
	m.On("PlayerHandPoints", 1).Return(7)
	m.On("PlayerHandPoints", 2).Return(9)
	m.On("PlayerMeldedCount", 0).Return(3)
	m.On("PlayerMeldedCount", 1).Return(0)
	m.On("PlayerMeldedCount", 2).Return(0)
	return m, players
}

func TestPanWebPresenter_Output(t *testing.T) {
	p := new(presenter.PanWebPresenter)

	t.Run("draw phase initial state", func(t *testing.T) {
		m, players := setupPanWebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

		out := p.Output(m, nil)
		assert.Contains(t, out, `"phase":0`)
		assert.Contains(t, out, `"roundNumber":1`)
		assert.Contains(t, out, `"drawPileCount":279`)
		assert.Contains(t, out, `"deckSize":320`)
		assert.Contains(t, out, `"winMeldCount":11`)
		assert.Contains(t, out, `"winnerIdx":-1`)
		assert.Contains(t, out, `"messageCode":"pan.drawPhase"`)
		var parsed map[string]any
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Len(t, parsed["players"].([]any), 3)
	})

	t.Run("error message propagated", func(t *testing.T) {
		m, _ := setupPanWebMock()
		out := p.Output(m, errors.New("bad"))
		assert.Contains(t, out, `"message":"bad"`)
	})

	t.Run("play phase message", func(t *testing.T) {
		m, _ := setupPanWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PanPhasePlay)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"messageCode":"pan.playPhase"`)
	})

	t.Run("round end reveals hands and message", func(t *testing.T) {
		m, _ := setupPanWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PanPhaseRoundEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"messageCode":"pan.roundEnd"`)
		assert.Contains(t, out, `"handPoints":7`)
	})

	t.Run("game end shows winner", func(t *testing.T) {
		m, _ := setupPanWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PanPhaseGameEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"gameEndFlag":true`)
		assert.Contains(t, out, `"messageCode":"pan.result.humanWin"`)
	})

	t.Run("discard top exposed", func(t *testing.T) {
		m, _ := setupPanWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		out := p.Output(m, nil)
		assert.Contains(t, out, `"discardTop":`)
	})

	t.Run("laid melds and chips exposed", func(t *testing.T) {
		m, players := setupPanWebMock()
		players[0].AddLaidMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
		})
		players[0].SetChips(2)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"laidMelds":[`)
		assert.Contains(t, out, `"chips":2`)
		assert.Contains(t, out, `"meldedCount":3`)
	})
}

func TestPanWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PanWebPresenter)
	m := new(interfaces.MockPanGame)
	m.On("GetGameEndFlag").Return(false)
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
