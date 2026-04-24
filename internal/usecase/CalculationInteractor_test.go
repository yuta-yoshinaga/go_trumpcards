package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockCalculationGame() *interfaces.MockCalculationGame {
	return new(interfaces.MockCalculationGame)
}

func newMockCalculationPresenter() *presenter.MockCalculationPresenter {
	return new(presenter.MockCalculationPresenter)
}

func TestNewCalculationInteractor(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)
	assert.NotNil(t, ci)
}

func TestNewCalculationInteractor_PanicsOnNil(t *testing.T) {
	p := newMockCalculationPresenter()
	assert.Panics(t, func() { NewCalculationInteractor(nil, p) })
	g := newMockCalculationGame()
	assert.Panics(t, func() { NewCalculationInteractor(g, nil) })
}

func TestCalculationInteractor_Reset(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_out")

	assert.Equal(t, "reset_out", ci.Reset())
	g.AssertCalled(t, "Reset")
}

func TestCalculationInteractor_PlayStockToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockCalculationGame()
		p := newMockCalculationPresenter()
		ci := NewCalculationInteractor(g, p)

		g.On("PlayStockToFoundation", 2).Return(nil)
		p.On("Output", g, nil).Return("ok")

		assert.Equal(t, "ok", ci.PlayStockToFoundation(2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockCalculationGame()
		p := newMockCalculationPresenter()
		ci := NewCalculationInteractor(g, p)

		e := errors.New("nope")
		g.On("PlayStockToFoundation", 1).Return(e)
		p.On("Output", g, e).Return("err")

		assert.Equal(t, "err", ci.PlayStockToFoundation(1))
	})
}

func TestCalculationInteractor_PlayStockToWaste(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	g.On("PlayStockToWaste", 3).Return(nil)
	p.On("Output", g, nil).Return("ok")

	assert.Equal(t, "ok", ci.PlayStockToWaste(3))
}

func TestCalculationInteractor_PlayWasteToFoundation(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	g.On("PlayWasteToFoundation", 1, 0).Return(nil)
	p.On("Output", g, nil).Return("ok")

	assert.Equal(t, "ok", ci.PlayWasteToFoundation(1, 0))
}

func TestCalculationInteractor_GiveUp(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup")

	assert.Equal(t, "giveup", ci.GiveUp())
}

func TestCalculationInteractor_Undo(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	g.On("Undo").Return(nil)
	p.On("Output", g, nil).Return("undo")

	assert.Equal(t, "undo", ci.Undo())
}

func TestCalculationInteractor_UndoN(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	g.On("UndoN", 2).Return(nil)
	p.On("Output", g, nil).Return("undon")

	assert.Equal(t, "undon", ci.UndoN(2))
}

func TestCalculationInteractor_AutoComplete(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	g.On("AutoComplete").Return(nil)
	p.On("Output", g, nil).Return("auto")

	assert.Equal(t, "auto", ci.AutoComplete())
}

func TestCalculationInteractor_Hint(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	assert.Equal(t, "hint", ci.Hint())
}

func TestCalculationInteractor_ActionLog(t *testing.T) {
	g := newMockCalculationGame()
	p := newMockCalculationPresenter()
	ci := NewCalculationInteractor(g, p)

	p.On("ActionLogOutput", g).Return("log")
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreCalculationInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"fd":[[],[],[],[]],"wa":[[],[],[],[]],"st":[],"ps":0,"mc":0,"al":[],"sl":false}`)
		p := newMockCalculationPresenter()
		ci, err := RestoreCalculationInteractor(data, p)
		assert.NoError(t, err)
		assert.NotNil(t, ci)
	})
	t.Run("invalid data", func(t *testing.T) {
		p := newMockCalculationPresenter()
		_, err := RestoreCalculationInteractor([]byte("invalid"), p)
		assert.Error(t, err)
	})
}
