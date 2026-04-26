//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockNertzGame() *interfaces.MockNertzGame {
	return new(interfaces.MockNertzGame)
}

func newMockNertzPresenter() *presenter.MockNertzPresenter {
	return new(presenter.MockNertzPresenter)
}

func TestNewNertzInteractor(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	assert.NotNil(t, NewNertzInteractor(g, p))
}

func TestNewNertzInteractor_PanicsOnNil(t *testing.T) {
	p := newMockNertzPresenter()
	assert.Panics(t, func() { NewNertzInteractor(nil, p) })
	g := newMockNertzGame()
	assert.Panics(t, func() { NewNertzInteractor(g, nil) })
}

func TestNertzInteractor_Reset(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_out")
	assert.Equal(t, "reset_out", i.Reset())
	g.AssertCalled(t, "Reset")
}

func TestNertzInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		g := newMockNertzGame()
		p := newMockNertzPresenter()
		i := NewNertzInteractor(g, p)
		cfg := domain.DefaultNertzConfig()
		g.On("ResetWithConfig", cfg).Return()
		p.On("Output", g, nil).Return("ok")
		assert.Equal(t, "ok", i.ResetWithConfig(cfg))
	})
	t.Run("invalid", func(t *testing.T) {
		g := newMockNertzGame()
		p := newMockNertzPresenter()
		i := NewNertzInteractor(g, p)
		bad := domain.DefaultNertzConfig()
		bad.PlayerCount = 0
		p.On("Output", g, mock.MatchedBy(func(err error) bool { return err != nil })).Return("err")
		assert.Equal(t, "err", i.ResetWithConfig(bad))
		g.AssertNotCalled(t, "ResetWithConfig", bad)
	})
}

func TestNertzInteractor_NextRound(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("NextRound").Return()
	p.On("Output", g, nil).Return("nr_out")
	assert.Equal(t, "nr_out", i.NextRound())
}

func TestNertzInteractor_Draw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockNertzGame()
		p := newMockNertzPresenter()
		i := NewNertzInteractor(g, p)
		g.On("DrawStock", 0).Return(nil)
		p.On("Output", g, nil).Return("draw_ok")
		assert.Equal(t, "draw_ok", i.Draw(0))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockNertzGame()
		p := newMockNertzPresenter()
		i := NewNertzInteractor(g, p)
		err := errors.New("empty")
		g.On("DrawStock", 0).Return(err)
		p.On("Output", g, err).Return("draw_err")
		assert.Equal(t, "draw_err", i.Draw(0))
	})
}

func TestNertzInteractor_MoveNertzToFoundation(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("MoveNertzToFoundation", 0, 1).Return(nil)
	p.On("Output", g, nil).Return("nf")
	assert.Equal(t, "nf", i.MoveNertzToFoundation(0, 1))
}

func TestNertzInteractor_MoveNertzToTableau(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("MoveNertzToTableau", 0, 2).Return(nil)
	p.On("Output", g, nil).Return("nt")
	assert.Equal(t, "nt", i.MoveNertzToTableau(0, 2))
}

func TestNertzInteractor_MoveWasteToFoundation(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("MoveWasteToFoundation", 0, 3).Return(nil)
	p.On("Output", g, nil).Return("wf")
	assert.Equal(t, "wf", i.MoveWasteToFoundation(0, 3))
}

func TestNertzInteractor_MoveWasteToTableau(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("MoveWasteToTableau", 0, 0).Return(nil)
	p.On("Output", g, nil).Return("wt")
	assert.Equal(t, "wt", i.MoveWasteToTableau(0, 0))
}

func TestNertzInteractor_MoveTableauToFoundation(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("MoveTableauToFoundation", 0, 1, 2).Return(nil)
	p.On("Output", g, nil).Return("tf")
	assert.Equal(t, "tf", i.MoveTableauToFoundation(0, 1, 2))
}

func TestNertzInteractor_MoveTableauToTableau(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("MoveTableauToTableau", 0, 1, 0, 2).Return(nil)
	p.On("Output", g, nil).Return("tt")
	assert.Equal(t, "tt", i.MoveTableauToTableau(0, 1, 0, 2))
}

func TestNertzInteractor_Tick(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	g.On("Tick").Return([]*domain.NertzAction{})
	p.On("Output", g, nil).Return("tick_out")
	assert.Equal(t, "tick_out", i.Tick())
}

func TestNertzInteractor_Hint(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	p.On("HintOutput", g).Return("hint_out")
	assert.Equal(t, "hint_out", i.Hint())
}

func TestNertzInteractor_Undo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockNertzGame()
		p := newMockNertzPresenter()
		i := NewNertzInteractor(g, p)
		g.On("Undo").Return(nil)
		p.On("Output", g, nil).Return("undo_ok")
		assert.Equal(t, "undo_ok", i.Undo())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockNertzGame()
		p := newMockNertzPresenter()
		i := NewNertzInteractor(g, p)
		err := errors.New("nothing")
		g.On("Undo").Return(err)
		p.On("Output", g, err).Return("undo_err")
		assert.Equal(t, "undo_err", i.Undo())
	})
}

func TestNertzInteractor_ActionLog(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	p.On("ActionLogOutput", g).Return("log")
	assert.Equal(t, "log", i.ActionLog())
}

func TestNertzInteractor_GetConfig(t *testing.T) {
	g := newMockNertzGame()
	p := newMockNertzPresenter()
	i := NewNertzInteractor(g, p)
	cfg := domain.DefaultNertzConfig()
	g.On("GetConfig").Return(cfg)
	assert.Equal(t, cfg, i.GetConfig())
}

func TestRestoreNertzInteractor(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		data := []byte(`{"ph":0,"rn":0,"wi":-1,"mw":-1,"mc":0,"cf":{"pc":4,"dc":3,"ts":100,"cd":1,"tm":0}}`)
		p := newMockNertzPresenter()
		i, err := RestoreNertzInteractor(data, p)
		assert.NoError(t, err)
		assert.NotNil(t, i)
	})
	t.Run("invalid", func(t *testing.T) {
		p := newMockNertzPresenter()
		_, err := RestoreNertzInteractor([]byte("bogus"), p)
		assert.Error(t, err)
	})
}
