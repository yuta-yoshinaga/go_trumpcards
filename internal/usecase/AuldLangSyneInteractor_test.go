package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockAuldLangSyneGame() *interfaces.MockAuldLangSyneGame {
	return new(interfaces.MockAuldLangSyneGame)
}

func newMockAuldLangSynePresenter() *presenter.MockAuldLangSynePresenter {
	return new(presenter.MockAuldLangSynePresenter)
}

func TestNewAuldLangSyneInteractor(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)
	assert.NotNil(t, ci)
}

func TestNewAuldLangSyneInteractor_PanicsOnNil(t *testing.T) {
	p := newMockAuldLangSynePresenter()
	assert.Panics(t, func() { NewAuldLangSyneInteractor(nil, p) })
	g := newMockAuldLangSyneGame()
	assert.Panics(t, func() { NewAuldLangSyneInteractor(g, nil) })
}

func TestAuldLangSyneInteractor_Reset(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_out")

	assert.Equal(t, "reset_out", ci.Reset())
	g.AssertCalled(t, "Reset")
}

// Deal is the command that replaces Sir Tommy's stock->waste move: it goes
// through execAndPresent, so a rejected deal must still render, carrying the
// error rather than dropping it.
func TestAuldLangSyneInteractor_Deal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockAuldLangSyneGame()
		p := newMockAuldLangSynePresenter()
		ci := NewAuldLangSyneInteractor(g, p)

		g.On("Deal").Return(nil)
		p.On("Output", g, nil).Return("dealt")

		assert.Equal(t, "dealt", ci.Deal())
		g.AssertCalled(t, "Deal")
	})
	t.Run("error", func(t *testing.T) {
		g := newMockAuldLangSyneGame()
		p := newMockAuldLangSynePresenter()
		ci := NewAuldLangSyneInteractor(g, p)

		e := errors.New("stock is empty")
		g.On("Deal").Return(e)
		p.On("Output", g, e).Return("err")

		assert.Equal(t, "err", ci.Deal())
	})
}

func TestAuldLangSyneInteractor_PlayWasteToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockAuldLangSyneGame()
		p := newMockAuldLangSynePresenter()
		ci := NewAuldLangSyneInteractor(g, p)

		g.On("PlayWasteToFoundation", 1, 0).Return(nil)
		p.On("Output", g, nil).Return("ok")

		assert.Equal(t, "ok", ci.PlayWasteToFoundation(1, 0))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockAuldLangSyneGame()
		p := newMockAuldLangSynePresenter()
		ci := NewAuldLangSyneInteractor(g, p)

		e := errors.New("cannot place card on foundation")
		g.On("PlayWasteToFoundation", 2, 3).Return(e)
		p.On("Output", g, e).Return("err")

		assert.Equal(t, "err", ci.PlayWasteToFoundation(2, 3))
	})
}

func TestAuldLangSyneInteractor_GiveUp(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup")

	assert.Equal(t, "giveup", ci.GiveUp())
}

func TestAuldLangSyneInteractor_Undo(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)

	g.On("Undo").Return(nil)
	p.On("Output", g, nil).Return("undo")

	assert.Equal(t, "undo", ci.Undo())
}

func TestAuldLangSyneInteractor_UndoN(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)

	g.On("UndoN", 2).Return(nil)
	p.On("Output", g, nil).Return("undon")

	assert.Equal(t, "undon", ci.UndoN(2))
}

func TestAuldLangSyneInteractor_AutoComplete(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)

	g.On("AutoComplete").Return(nil)
	p.On("Output", g, nil).Return("auto")

	assert.Equal(t, "auto", ci.AutoComplete())
}

func TestAuldLangSyneInteractor_Hint(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)

	p.On("HintOutput", g).Return("hint")
	assert.Equal(t, "hint", ci.Hint())
}

func TestAuldLangSyneInteractor_ActionLog(t *testing.T) {
	g := newMockAuldLangSyneGame()
	p := newMockAuldLangSynePresenter()
	ci := NewAuldLangSyneInteractor(g, p)

	p.On("ActionLogOutput", g).Return("log")
	assert.Equal(t, "log", ci.ActionLog())
}

func TestRestoreAuldLangSyneInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"fd":[[],[],[],[]],"wa":[[],[],[],[]],"st":[],"ps":0,"mc":0,"al":[],"sl":false}`)
		p := newMockAuldLangSynePresenter()
		ci, err := RestoreAuldLangSyneInteractor(data, p)
		assert.NoError(t, err)
		assert.NotNil(t, ci)
	})
	t.Run("invalid data", func(t *testing.T) {
		p := newMockAuldLangSynePresenter()
		_, err := RestoreAuldLangSyneInteractor([]byte("invalid"), p)
		assert.Error(t, err)
	})
}

// Snapshot is what the Cloudflare Worker persists to KV between requests, so a
// restored interactor must be able to round-trip its own state back out.
func TestAuldLangSyneInteractor_Snapshot(t *testing.T) {
	data := []byte(`{"tc":{},"fd":[[],[],[],[]],"wa":[[],[],[],[]],"st":[],"ps":0,"mc":0,"al":[],"sl":false}`)
	p := newMockAuldLangSynePresenter()
	ci, err := RestoreAuldLangSyneInteractor(data, p)
	assert.NoError(t, err)

	out, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
}
