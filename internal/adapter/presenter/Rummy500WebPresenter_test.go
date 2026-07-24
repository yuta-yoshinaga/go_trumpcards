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

func setupRummy500WebMock() (*interfaces.MockRummy500Game, []*domain.Rummy500Player) {
	m := new(interfaces.MockRummy500Game)
	players := []*domain.Rummy500Player{
		domain.NewRummy500Player(true),
		domain.NewRummy500Player(false),
	}
	m.On("GetPhase").Return(domain.Rummy500PhaseDraw)
	m.On("GetRoundNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDrawPileCount").Return(25)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetRoundEnderIdx").Return(-1)
	m.On("GetDiscardPile").Return(([]*domain.Card)(nil))
	m.On("GetConfig").Return(domain.DefaultRummy500Config())
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestRummy500WebPresenter_Output(t *testing.T) {
	p := new(presenter.Rummy500WebPresenter)

	t.Run("draw phase initial state", func(t *testing.T) {
		m, players := setupRummy500WebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

		out := p.Output(m, nil)
		assert.Contains(t, out, `"phase":0`)
		assert.Contains(t, out, `"roundNumber":1`)
		assert.Contains(t, out, `"drawPileCount":25`)
		assert.Contains(t, out, `"winnerIdx":-1`)
		assert.Contains(t, out, `"messageCode":"rummy500.drawPhase"`)
		var parsed map[string]any
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		players2 := parsed["players"].([]any)
		assert.Len(t, players2, 2)
	})

	t.Run("error message propagated", func(t *testing.T) {
		m, _ := setupRummy500WebMock()
		out := p.Output(m, errors.New("bad"))
		assert.Contains(t, out, `"message":"bad"`)
	})

	t.Run("play phase message", func(t *testing.T) {
		m, _ := setupRummy500WebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.Rummy500PhasePlay)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"messageCode":"rummy500.playPhase"`)
	})

	t.Run("round end message", func(t *testing.T) {
		m, _ := setupRummy500WebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.Rummy500PhaseRoundEnd)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"messageCode":"rummy500.roundEnd"`)
	})

	t.Run("game end shows winner", func(t *testing.T) {
		m, _ := setupRummy500WebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetWinnerIdx").Return(0)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"gameEndFlag":true`)
	})

	t.Run("discard pile cards exposed", func(t *testing.T) {
		m, _ := setupRummy500WebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardPile")
		pile := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		}
		m.On("GetDiscardPile").Return(pile)
		out := p.Output(m, nil)
		assert.Contains(t, out, `"discardPile":[`)
	})

	t.Run("player laid melds exposed", func(t *testing.T) {
		m, players := setupRummy500WebMock()
		players[0].AddLaidMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		})
		out := p.Output(m, nil)
		assert.Contains(t, out, `"laidMelds":[`)
	})

	t.Run("CPU cards hidden during play", func(t *testing.T) {
		m, players := setupRummy500WebMock()
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		out := p.Output(m, nil)
		// CPU cards should not leak in their direct ranks during draw phase
		assert.Contains(t, out, `"cardCount":1`)
	})
}

func TestRummy500WebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.Rummy500WebPresenter)
	m := new(interfaces.MockRummy500Game)
	m.On("GetGameEndFlag").Return(false)
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}

func TestRummy500WebPresenter_HintOutput(t *testing.T) {
	m, players := setupRummy500WebMock()
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	p := new(presenter.Rummy500WebPresenter)
	// The web presenter computes hints client-side, so HintOutput mirrors Output.
	assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
}
