//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKalookiWebMock() (*interfaces.MockKalookiGame, []*domain.KalookiPlayer) {
	m := new(interfaces.MockKalookiGame)
	players := []*domain.KalookiPlayer{
		domain.NewKalookiPlayer(true),
		domain.NewKalookiPlayer(false),
		domain.NewKalookiPlayer(false),
		domain.NewKalookiPlayer(false),
	}
	m.On("GetOpeningThreshold").Return(51)
	m.On("GetDrawPileCount").Return(53)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.KalookiPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultKalookiConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func unmarshalKalooki(t *testing.T, s string) controller.KalookiWebOutput {
	t.Helper()
	var out controller.KalookiWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestKalookiWebPresenter_Output(t *testing.T) {
	p := new(presenter.KalookiWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupKalookiWebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
		})
		players[0].SetHasOpened(true)

		out := unmarshalKalooki(t, p.Output(m, nil))
		assert.Len(t, out.Players, 4)
		assert.Equal(t, 51, out.OpeningThreshold)
		assert.True(t, out.Players[0].HasOpened)
		assert.Len(t, out.Players[0].Melds, 1)
		assert.Equal(t, "kalooki.drawPhase", out.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupKalookiWebMock()
		out := unmarshalKalooki(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("meld phase", func(t *testing.T) {
		m := new(interfaces.MockKalookiGame)
		players := []*domain.KalookiPlayer{domain.NewKalookiPlayer(true), domain.NewKalookiPlayer(false)}
		m.On("GetOpeningThreshold").Return(51)
		m.On("GetDrawPileCount").Return(40)
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.KalookiPhaseMeld)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetConfig").Return(domain.DefaultKalookiConfig())
		m.On("GetRoundWinnerIdx").Return(-1)
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		out := unmarshalKalooki(t, p.Output(m, nil))
		assert.Equal(t, "kalooki.meldPhase", out.MessageCode)
		assert.NotNil(t, out.DiscardTop)
	})

	t.Run("game ended", func(t *testing.T) {
		m := new(interfaces.MockKalookiGame)
		players := []*domain.KalookiPlayer{domain.NewKalookiPlayer(true), domain.NewKalookiPlayer(false)}
		m.On("GetOpeningThreshold").Return(51)
		m.On("GetDrawPileCount").Return(0)
		m.On("GetDiscardTop").Return((*domain.Card)(nil))
		m.On("GetGameEndFlag").Return(true)
		m.On("GetPhase").Return(domain.KalookiPhaseGameEnd)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(0)
		m.On("GetConfig").Return(domain.DefaultKalookiConfig())
		m.On("GetRoundWinnerIdx").Return(0)
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		out := unmarshalKalooki(t, p.Output(m, nil))
		assert.True(t, out.GameEndFlag)
	})

	t.Run("cpu hands hidden during play", func(t *testing.T) {
		m, players := setupKalookiWebMock()
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

		out := unmarshalKalooki(t, p.Output(m, nil))
		// Draw phase: CPU card faces are hidden, only the count is exposed.
		assert.Equal(t, 2, out.Players[1].CardCount)
		assert.Empty(t, out.Players[1].Cards)
	})

	t.Run("cpu hands revealed at round end", func(t *testing.T) {
		m := new(interfaces.MockKalookiGame)
		players := []*domain.KalookiPlayer{domain.NewKalookiPlayer(true), domain.NewKalookiPlayer(false)}
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		m.On("GetOpeningThreshold").Return(51)
		m.On("GetDrawPileCount").Return(20)
		m.On("GetDiscardTop").Return((*domain.Card)(nil))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.KalookiPhaseRoundEnd)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetConfig").Return(domain.DefaultKalookiConfig())
		m.On("GetRoundWinnerIdx").Return(0)
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])

		out := unmarshalKalooki(t, p.Output(m, nil))
		// Round end: CPU card faces are revealed so penalty scores can be verified.
		assert.Equal(t, "kalooki.roundEnd", out.MessageCode)
		assert.Len(t, out.Players[1].Cards, 2)
	})
}

func TestKalookiWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KalookiWebPresenter)
	m, _ := setupKalookiWebMock()
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
