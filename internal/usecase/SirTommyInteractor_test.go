package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockSirTommyGame() *interfaces.MockSirTommyGame {
	return new(interfaces.MockSirTommyGame)
}

func newMockSirTommyPresenter() *presenter.MockSirTommyPresenter {
	return new(presenter.MockSirTommyPresenter)
}

func TestNewSirTommyInteractor(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)
	assert.NotNil(t, ci)
}

func TestNewSirTommyInteractor_PanicsOnNil(t *testing.T) {
	p := newMockSirTommyPresenter()
	assert.Panics(t, func() { NewSirTommyInteractor(nil, p) })
	g := newMockSirTommyGame()
	assert.Panics(t, func() { NewSirTommyInteractor(g, nil) })
}

func TestSirTommyInteractor_Reset(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_out")

	assert.Equal(t, "reset_out", ci.Reset())
	g.AssertCalled(t, "Reset")
}

func TestSirTommyInteractor_PlayStockToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockSirTommyGame()
		p := newMockSirTommyPresenter()
		ci := NewSirTommyInteractor(g, p)

		g.On("PlayStockToFoundation", 2).Return(nil)
		p.On("Output", g, nil).Return("ok")

		assert.Equal(t, "ok", ci.PlayStockToFoundation(2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockSirTommyGame()
		p := newMockSirTommyPresenter()
		ci := NewSirTommyInteractor(g, p)

		e := errors.New("nope")
		g.On("PlayStockToFoundation", 1).Return(e)
		p.On("Output", g, e).Return("err")

		assert.Equal(t, "err", ci.PlayStockToFoundation(1))
	})
}

func TestSirTommyInteractor_PlayStockToWaste(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	g.On("PlayStockToWaste", 3).Return(nil)
	p.On("Output", g, nil).Return("ok")

	assert.Equal(t, "ok", ci.PlayStockToWaste(3))
}

func TestSirTommyInteractor_PlayWasteToFoundation(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	g.On("PlayWasteToFoundation", 1, 0).Return(nil)
	p.On("Output", g, nil).Return("ok")

	assert.Equal(t, "ok", ci.PlayWasteToFoundation(1, 0))
}

func TestSirTommyInteractor_GiveUp(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup")

	assert.Equal(t, "giveup", ci.GiveUp())
}

func TestSirTommyInteractor_Undo(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	g.On("Undo").Return(nil)
	p.On("Output", g, nil).Return("undo")

	assert.Equal(t, "undo", ci.Undo())
}

func TestSirTommyInteractor_UndoN(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	g.On("UndoN", 2).Return(nil)
	p.On("Output", g, nil).Return("undon")

	assert.Equal(t, "undon", ci.UndoN(2))
}

func TestSirTommyInteractor_AutoComplete(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	g.On("AutoComplete").Return(nil)
	p.On("Output", g, nil).Return("auto")

	assert.Equal(t, "auto", ci.AutoComplete())
}

func TestSirTommyInteractor_Hint(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	assert.Equal(t, "hint", ci.Hint())
}

func TestSirTommyInteractor_ActionLog(t *testing.T) {
	g := newMockSirTommyGame()
	p := newMockSirTommyPresenter()
	ci := NewSirTommyInteractor(g, p)

	p.On("ActionLogOutput", g).Return("log")
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreSirTommyInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"fd":[[],[],[],[]],"wa":[[],[],[],[]],"st":[],"ps":0,"mc":0,"al":[],"sl":false}`)
		p := newMockSirTommyPresenter()
		ci, err := RestoreSirTommyInteractor(data, p)
		assert.NoError(t, err)
		assert.NotNil(t, ci)
	})
	t.Run("invalid data", func(t *testing.T) {
		p := newMockSirTommyPresenter()
		_, err := RestoreSirTommyInteractor([]byte("invalid"), p)
		assert.Error(t, err)
	})
}
