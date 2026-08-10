//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupConquianWebMock(phase domain.ConquianPhase, ended bool, winner, roundWinner int) (*interfaces.MockConquianGame, []*domain.ConquianPlayer) {
	m := new(interfaces.MockConquianGame)
	players := makeConquianPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetExtendableMeldIndices", mock.Anything, mock.Anything).Return(([]int)(nil)).Maybe()
	m.On("GetDrawPileCount").Return(20)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(ended)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetRoundWinnerIdx").Return(roundWinner)
	m.On("GetTookDiscard").Return(false)
	m.On("GetConfig").Return(domain.DefaultConquianConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

// **押せる先だけを押せるように。**♠5 は「5 のセット」も「♠4-6-7 のラン」も
// 延長できる。画面が「どれでも押せる」ように見せると、押した先と実際に足される
// 先が食い違う (#4837)。
func TestConquianWebPresenter_LayoffTargets(t *testing.T) {
	m, players := setupConquianWebMock(domain.ConquianPhaseMeld, false, -1, -1)
	players[0].SetMelds([][]*domain.Card{
		{
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		},
	})
	for players[0].GetCardsSize() > 0 {
		players[0].RemoveCard(0)
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetExtendableMeldIndices")
	m.On("GetExtendableMeldIndices", 0, mock.MatchedBy(func(c *domain.Card) bool {
		return c != nil && c.GetValue() == 5
	})).Return([]int{0})
	m.On("GetExtendableMeldIndices", mock.Anything, mock.Anything).Return(([]int)(nil))

	var out controller.ConquianWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.ConquianWebPresenter).Output(m, nil)), &out))

	// 5 は延長できる。2 はできない — 空配列であって null ではない。
	assert.Equal(t, [][]int{{0}, {}}, out.LayoffTargets)
}

func TestConquianWebPresenter_Output(t *testing.T) {
	p := new(presenter.ConquianWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		m, players := setupConquianWebMock(domain.ConquianPhaseDraw, false, -1, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[0].AddMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignDiamond, 3, false),
		})
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, 0, parsed.Phase)
		assert.Len(t, parsed.Players, 2)
		assert.Len(t, parsed.Players[0].Melds, 1)
		// CPU hand hidden during play
		assert.Empty(t, parsed.Players[1].Cards)
	})

	t.Run("reveals CPU hand at round end", func(t *testing.T) {
		m, players := setupConquianWebMock(domain.ConquianPhaseRoundEnd, false, -1, 0)
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Len(t, parsed.Players[1].Cards, 1)
	})

	t.Run("game end with human winner", func(t *testing.T) {
		m, _ := setupConquianWebMock(domain.ConquianPhaseGameEnd, true, 0, 0)
		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.True(t, parsed.GameEndFlag)
		assert.Equal(t, "conquian.result.humanWin", parsed.MessageCode)
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupConquianWebMock(domain.ConquianPhaseGameEnd, true, -1, -1)
		out := p.Output(m, nil)
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "conquian.draw", parsed.MessageCode)
	})

	t.Run("error is surfaced", func(t *testing.T) {
		m, _ := setupConquianWebMock(domain.ConquianPhaseDraw, false, -1, -1)
		out := p.Output(m, errors.New("boom"))
		var parsed controller.ConquianWebOutput
		require.NoError(t, json.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, "boom", parsed.Message)
	})
}

func TestConquianWebPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupConquianWebMock(domain.ConquianPhaseDraw, false, -1, -1)
	p := new(presenter.ConquianWebPresenter)
	assert.NotPanics(t, func() { p.ActionLogOutput(m) })
}
