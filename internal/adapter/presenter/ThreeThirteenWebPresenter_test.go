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

func setupThreeThirteenWebMock(phase domain.ThreeThirteenPhase, gameEnd bool) (*interfaces.MockThreeThirteenGame, []*domain.ThreeThirteenPlayer) {
	m := new(interfaces.MockThreeThirteenGame)
	players := []*domain.ThreeThirteenPlayer{
		domain.NewThreeThirteenPlayer(true),
		domain.NewThreeThirteenPlayer(false),
	}
	m.On("GetRound").Return(2)
	m.On("WildRank").Return(4)
	m.On("GetDealCount").Return(4)
	m.On("GetDrawPileCount").Return(80)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 3, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetKnockerIdx").Return(-1)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultThreeThirteenConfig())
	m.On("GetPlayerCnt").Return(2)
	for i := 0; i < 2; i++ {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetPlayerDeadwoodValue", i).Return(7)
	}
	return m, players
}

func unmarshalThreeThirteen(t *testing.T, s string) controller.ThreeThirteenWebOutput {
	t.Helper()
	var out controller.ThreeThirteenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestThreeThirteenWebPresenter_Output(t *testing.T) {
	p := new(presenter.ThreeThirteenWebPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, players := setupThreeThirteenWebMock(domain.ThreeThirteenPhaseDraw, false)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		out := unmarshalThreeThirteen(t, p.Output(m, nil))
		assert.Len(t, out.Players, 2)
		assert.Equal(t, 2, out.Round)
		assert.Equal(t, 4, out.WildRank)
		assert.Equal(t, "threethirteen.drawPhase", out.MessageCode)
		assert.NotNil(t, out.DiscardTop)
	})

	t.Run("discard phase", func(t *testing.T) {
		m, _ := setupThreeThirteenWebMock(domain.ThreeThirteenPhaseDiscard, false)
		out := unmarshalThreeThirteen(t, p.Output(m, nil))
		assert.Equal(t, "threethirteen.discardPhase", out.MessageCode)
	})

	t.Run("round end reveals all", func(t *testing.T) {
		m, _ := setupThreeThirteenWebMock(domain.ThreeThirteenPhaseRoundEnd, false)
		out := unmarshalThreeThirteen(t, p.Output(m, nil))
		assert.Equal(t, "threethirteen.roundEnd", out.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupThreeThirteenWebMock(domain.ThreeThirteenPhaseDraw, false)
		out := unmarshalThreeThirteen(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("game ended", func(t *testing.T) {
		m, _ := setupThreeThirteenWebMock(domain.ThreeThirteenPhaseGameEnd, true)
		out := unmarshalThreeThirteen(t, p.Output(m, nil))
		assert.True(t, out.GameEndFlag)
	})
}

func TestThreeThirteenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ThreeThirteenWebPresenter)
	m, _ := setupThreeThirteenWebMock(domain.ThreeThirteenPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	assert.NotEmpty(t, p.ActionLogOutput(m))
}
