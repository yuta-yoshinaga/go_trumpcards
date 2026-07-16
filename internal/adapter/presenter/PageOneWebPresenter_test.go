package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupPageOneWebMock() *interfaces.MockPageOneGame {
	m := new(interfaces.MockPageOneGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PageOnePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultPageOneConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	players := makePageOnePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m
}

func TestPageOneWebPresenter_Output(t *testing.T) {
	p := new(presenter.PageOneWebPresenter)

	t.Run("play phase", func(t *testing.T) {
		m := setupPageOneWebMock()
		result := p.Output(m, nil)
		var out map[string]interface{}
		require := assert.NoError
		require(t, json.Unmarshal([]byte(result), &out))
		assert.Equal(t, float64(0), out["phase"])
		assert.Equal(t, "pageone.playPhase", out["messageCode"])
	})

	t.Run("must declare phase", func(t *testing.T) {
		m := setupPageOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PageOnePhaseMustDeclare)
		result := p.Output(m, nil)
		var out map[string]interface{}
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "pageone.mustDeclarePhase", out["messageCode"])
	})

	t.Run("round end phase", func(t *testing.T) {
		m := setupPageOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PageOnePhaseRoundEnd)
		result := p.Output(m, nil)
		var out map[string]interface{}
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "pageone.roundEnd", out["messageCode"])
	})

	t.Run("game end", func(t *testing.T) {
		m := setupPageOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "pageone")
	})

	t.Run("error passthrough", func(t *testing.T) {
		m := setupPageOneWebMock()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("discard top included", func(t *testing.T) {
		m := setupPageOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "discardTop")
	})
}

func TestPageOneWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PageOneWebPresenter)
	m := new(interfaces.MockPageOneGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

func TestPageOneWebPresenter_HintOutput(t *testing.T) {
	m := setupPageOneWebMock()
	p := new(presenter.PageOneWebPresenter)
	// The web presenter computes hints client-side, so HintOutput mirrors Output.
	assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
}
