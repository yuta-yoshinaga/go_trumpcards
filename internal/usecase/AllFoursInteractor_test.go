//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newAllFoursMocks() (*interfaces.MockAllFoursGame, *presenter.MockAllFoursPresenter) {
	return new(interfaces.MockAllFoursGame), new(presenter.MockAllFoursPresenter)
}

func TestAllFoursInteractor_Reset(t *testing.T) {
	g, pp := newAllFoursMocks()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.AllFoursPhaseBeg)
	pp.On("Output", mock.Anything, mock.Anything).Return("ok")

	ai := usecase.NewAllFoursInteractor(g, pp)
	assert.Equal(t, "ok", ai.Reset())
	g.AssertCalled(t, "Reset")
}

func TestAllFoursInteractor_Beg(t *testing.T) {
	t.Run("game end short-circuits", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("GetGameEndFlag").Return(true)
		pp.On("Output", mock.Anything, mock.Anything).Return("ended")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "ended", ai.Beg(true))
		g.AssertNotCalled(t, "PlayerBeg")
	})
	t.Run("error returns presenter output", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerBeg", true).Return(errors.New("oops"))
		pp.On("Output", g, mock.Anything).Return("err")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "err", ai.Beg(true))
	})
	t.Run("valid beg runs cpu turns", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerBeg", false).Return(nil)
		g.On("GetPhase").Return(domain.AllFoursPhaseGift)
		pp.On("Output", g, mock.Anything).Return("ok")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "ok", ai.Beg(false))
	})
}

func TestAllFoursInteractor_RespondBeg(t *testing.T) {
	g, pp := newAllFoursMocks()
	g.On("GetGameEndFlag").Return(false)
	g.On("PlayerRespondBeg", true).Return(nil)
	g.On("GetPhase").Return(domain.AllFoursPhaseBeg)
	pp.On("Output", g, mock.Anything).Return("ok")
	ai := usecase.NewAllFoursInteractor(g, pp)
	assert.Equal(t, "ok", ai.RespondBeg(true))
	g.AssertCalled(t, "PlayerRespondBeg", true)
}

func TestAllFoursInteractor_Play(t *testing.T) {
	t.Run("not playable short-circuits", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("GetGameEndFlag").Return(true)
		pp.On("Output", mock.Anything, mock.Anything).Return("ended")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "ended", ai.Play(0))
		g.AssertNotCalled(t, "PlayerPlay")
	})
	t.Run("invalid play returns error", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		g.On("PlayerPlay", 99).Return(errors.New("oob"))
		pp.On("Output", g, mock.Anything).Return("err")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "err", ai.Play(99))
	})
	t.Run("valid play runs cpu turns", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		g.On("GetPhase").Return(domain.AllFoursPhasePlay)
		g.On("PlayerPlay", 0).Return(nil)
		pp.On("Output", g, mock.Anything).Return("ok")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "ok", ai.Play(0))
	})
}

func TestAllFoursInteractor_NextTrick(t *testing.T) {
	t.Run("resolves trick then advances", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		// First call: trickEnd → resolve. Then play.
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.AllFoursPhaseTrickEnd).Once()
		g.On("GetPhase").Return(domain.AllFoursPhaseTrickEnd).Once()
		g.On("GetPhase").Return(domain.AllFoursPhaseTrickEnd).Once()
		g.On("GetPhase").Return(domain.AllFoursPhasePlay)
		g.On("ResolveTrick").Return()
		g.On("NextTrick").Return()
		g.On("IsHumanTurn").Return(true)
		pp.On("Output", g, mock.Anything).Return("ok")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "ok", ai.NextTrick())
		g.AssertCalled(t, "ResolveTrick")
	})
	t.Run("round end after resolve", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("GetPhase").Return(domain.AllFoursPhaseTrickEnd).Once()
		g.On("GetPhase").Return(domain.AllFoursPhaseRoundEnd)
		g.On("ResolveTrick").Return()
		pp.On("Output", g, mock.Anything).Return("roundend")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "roundend", ai.NextTrick())
	})
}

func TestAllFoursInteractor_NextRound(t *testing.T) {
	t.Run("game ended after score", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("ScoreRound").Return()
		g.On("GetGameEndFlag").Return(true)
		pp.On("Output", mock.Anything, mock.Anything).Return("ended")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "ended", ai.NextRound())
		g.AssertNotCalled(t, "NextRound")
	})
	t.Run("continues to next deal", func(t *testing.T) {
		g, pp := newAllFoursMocks()
		g.On("ScoreRound").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return()
		g.On("GetPhase").Return(domain.AllFoursPhaseBeg)
		pp.On("Output", g, mock.Anything).Return("ok")
		ai := usecase.NewAllFoursInteractor(g, pp)
		assert.Equal(t, "ok", ai.NextRound())
		g.AssertCalled(t, "NextRound")
	})
}

func TestAllFoursInteractor_HintAndLog(t *testing.T) {
	g, pp := newAllFoursMocks()
	pp.On("HintOutput", g).Return("hint")
	pp.On("ActionLogOutput", g).Return("log")
	ai := usecase.NewAllFoursInteractor(g, pp)
	assert.Equal(t, "hint", ai.Hint())
	assert.Equal(t, "log", ai.ActionLog())
}

func TestAllFoursInteractor_GetConfigAndResetWithConfig(t *testing.T) {
	g, pp := newAllFoursMocks()
	cfg := domain.DefaultAllFoursConfig()
	g.On("GetConfig").Return(cfg)
	g.On("SetConfig", cfg).Return()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.AllFoursPhaseBeg)
	pp.On("Output", mock.Anything, mock.Anything).Return("ok")
	ai := usecase.NewAllFoursInteractor(g, pp)
	assert.Equal(t, cfg, ai.GetConfig())
	assert.Equal(t, "ok", ai.ResetWithConfig(cfg))
}

func TestRestoreAllFoursInteractor(t *testing.T) {
	src := domain.NewDefaultAllFours()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)
	pp := new(presenter.MockAllFoursPresenter)
	ai, err := usecase.RestoreAllFoursInteractor(data, pp)
	assert.NoError(t, err)
	assert.NotNil(t, ai)
}
