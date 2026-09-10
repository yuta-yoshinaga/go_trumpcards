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

func setupMarriageWebMock(phase domain.MarriagePhase, gameEnd bool) (*interfaces.MockMarriageGame, []*domain.MarriagePlayer) {
	m := new(interfaces.MockMarriageGame)
	players := []*domain.MarriagePlayer{
		domain.NewMarriagePlayer(true),
		domain.NewMarriagePlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRoundNumber").Return(1)
	m.On("GetTargetRounds").Return(3)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetWildJoker").Return(domain.NewCard(domain.CardDesignDiamond, 9, false))
	m.On("GetWildRank").Return(9)
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetDeclarerIdx").Return(-1)
	m.On("GetDeclarationValid").Return(false)
	m.On("GetConfig").Return(domain.DefaultMarriageConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("PlayerDeadwoodValue", 0).Return(5)
	m.On("PlayerDeadwoodValue", 1).Return(80)
	m.On("PlayerMaalValue", 0).Return(3)
	m.On("PlayerMaalValue", 1).Return(4)
	m.On("PlayerHasPureSequence", 0).Return(true)
	m.On("PlayerHasPureSequence", 1).Return(false)
	return m, players
}

func unmarshalMarriage(t *testing.T, s string) controller.MarriageWebOutput {
	t.Helper()
	var out controller.MarriageWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestMarriageWebPresenter_Output(t *testing.T) {
	p := new(presenter.MarriageWebPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupMarriageWebMock(domain.MarriagePhaseDraw, false)
		out := unmarshalMarriage(t, p.Output(m, nil))
		assert.Len(t, out.Players, 2)
		assert.Equal(t, 1, out.RoundNumber)
		assert.Equal(t, 3, out.TargetRounds)
		assert.Equal(t, 9, out.WildRank)
		assert.Equal(t, 3, out.Players[0].Maal)
		assert.NotNil(t, out.WildJoker)
		assert.Equal(t, "marriage.drawPhase", out.MessageCode)
	})

	t.Run("discard phase", func(t *testing.T) {
		m, _ := setupMarriageWebMock(domain.MarriagePhaseDiscard, false)
		out := unmarshalMarriage(t, p.Output(m, nil))
		assert.Equal(t, "marriage.discardPhase", out.MessageCode)
	})

	t.Run("round end reveals cpu and deadwood", func(t *testing.T) {
		m, _ := setupMarriageWebMock(domain.MarriagePhaseRoundEnd, false)
		out := unmarshalMarriage(t, p.Output(m, nil))
		assert.Equal(t, "marriage.roundEnd", out.MessageCode)
		assert.Equal(t, 80, out.Players[1].Deadwood)
		assert.Equal(t, 4, out.Players[1].Maal)
		assert.False(t, out.Players[1].HasPureSequence)
	})

	t.Run("game ended", func(t *testing.T) {
		m, _ := setupMarriageWebMock(domain.MarriagePhaseGameEnd, true)
		out := unmarshalMarriage(t, p.Output(m, nil))
		assert.True(t, out.GameEndFlag)
	})

	t.Run("with error", func(t *testing.T) {
		m, _ := setupMarriageWebMock(domain.MarriagePhaseDraw, false)
		out := unmarshalMarriage(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})
}

// **CUI は毎ターン出しているのに、Web は狭い条件でしか出していなかった (#4824)。**
func TestMarriageWebPresenter_HumanHandStatus(t *testing.T) {
	m, _ := setupMarriageWebMock(domain.MarriagePhaseDiscard, false)

	var out controller.MarriageWebOutput
	assert.NoError(t, json.Unmarshal([]byte(new(presenter.MarriageWebPresenter).Output(m, nil)), &out))

	// 人間 (席 0) の値は公開前でも載る。CPU (席 1) は伏せたまま。
	assert.Equal(t, 5, out.Players[0].Deadwood)
	assert.True(t, out.Players[0].HasPureSequence)
	assert.Equal(t, 0, out.Players[1].Deadwood, "CPU のデッドウッドは伏せる")
	assert.False(t, out.Players[1].HasPureSequence)
}

func TestMarriageWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MarriageWebPresenter)
	m, _ := setupMarriageWebMock(domain.MarriagePhaseDraw, false)
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
