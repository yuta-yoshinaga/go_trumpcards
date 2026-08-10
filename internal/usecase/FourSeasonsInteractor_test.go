package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockFourSeasonsGame() *interfaces.MockFourSeasonsGame {
	return new(interfaces.MockFourSeasonsGame)
}

func newMockFourSeasonsPresenter() *presenter.MockFourSeasonsPresenter {
	return new(presenter.MockFourSeasonsPresenter)
}

func TestNewFourSeasonsInteractor(t *testing.T) {
	assert.NotNil(t, NewFourSeasonsInteractor(newMockFourSeasonsGame(), newMockFourSeasonsPresenter()))
}

func TestNewFourSeasonsInteractor_PanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewFourSeasonsInteractor(nil, newMockFourSeasonsPresenter()) })
	assert.Panics(t, func() { NewFourSeasonsInteractor(newMockFourSeasonsGame(), nil) })
}

func TestFourSeasonsInteractor_Reset(t *testing.T) {
	g, p := newMockFourSeasonsGame(), newMockFourSeasonsPresenter()
	fi := NewFourSeasonsInteractor(g, p)
	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_out")
	assert.Equal(t, "reset_out", fi.Reset())
	g.AssertCalled(t, "Reset")
}

func TestFourSeasonsInteractor_GiveUp(t *testing.T) {
	g, p := newMockFourSeasonsGame(), newMockFourSeasonsPresenter()
	fi := NewFourSeasonsInteractor(g, p)
	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup")
	assert.Equal(t, "giveup", fi.GiveUp())
}

// Every action goes through execAndPresent, so a rejected move must still
// render — carrying the error rather than dropping it.
func TestFourSeasonsInteractor_ActionsPresentBothOutcomes(t *testing.T) {
	boom := errors.New("nope")
	tests := []struct {
		name string
		arm  func(*interfaces.MockFourSeasonsGame, error)
		call func(*FourSeasonsInteractor) string
		asrt string
	}{
		{"Draw", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("Draw").Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.Draw() }, "Draw"},
		{"MoveWasteToTableau", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("MoveWasteToTableau", 2).Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.MoveWasteToTableau(2) }, "MoveWasteToTableau"},
		{"MoveWasteToFoundation", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("MoveWasteToFoundation", 1).Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.MoveWasteToFoundation(1) }, "MoveWasteToFoundation"},
		{"MoveTableauToTableau", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("MoveTableauToTableau", 0, 3).Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.MoveTableauToTableau(0, 3) }, "MoveTableauToTableau"},
		{"MoveTableauToFoundation", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("MoveTableauToFoundation", 4, 0).Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.MoveTableauToFoundation(4, 0) }, "MoveTableauToFoundation"},
		{"AutoComplete", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("AutoComplete").Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.AutoComplete() }, "AutoComplete"},
		{"Undo", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("Undo").Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.Undo() }, "Undo"},
		{"UndoN", func(g *interfaces.MockFourSeasonsGame, e error) { g.On("UndoN", 2).Return(e) },
			func(fi *FourSeasonsInteractor) string { return fi.UndoN(2) }, "UndoN"},
	}
	for _, tt := range tests {
		t.Run(tt.name+" success", func(t *testing.T) {
			g, p := newMockFourSeasonsGame(), newMockFourSeasonsPresenter()
			fi := NewFourSeasonsInteractor(g, p)
			tt.arm(g, nil)
			p.On("Output", g, nil).Return("ok")
			assert.Equal(t, "ok", tt.call(fi))
			g.AssertCalled(t, tt.asrt, g.Calls[0].Arguments...)
		})
		t.Run(tt.name+" error", func(t *testing.T) {
			g, p := newMockFourSeasonsGame(), newMockFourSeasonsPresenter()
			fi := NewFourSeasonsInteractor(g, p)
			tt.arm(g, boom)
			p.On("Output", g, boom).Return("err")
			assert.Equal(t, "err", tt.call(fi))
		})
	}
}

func TestFourSeasonsInteractor_HintAndActionLog(t *testing.T) {
	g, p := newMockFourSeasonsGame(), newMockFourSeasonsPresenter()
	fi := NewFourSeasonsInteractor(g, p)
	p.On("HintOutput", g).Return("hint")
	p.On("ActionLogOutput", g).Return("log")
	assert.Equal(t, "hint", fi.Hint())
	assert.Equal(t, "log", fi.ActionLog())
}

func TestRestoreFourSeasonsInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"tb":[[],[],[],[],[]],"fd":[[],[],[],[]],"st":[],"wa":[],"br":7,"ps":0,"mc":0,"al":[]}`)
		fi, err := RestoreFourSeasonsInteractor(data, newMockFourSeasonsPresenter())
		assert.NoError(t, err)
		assert.NotNil(t, fi)

		// The KV blob is what the Worker rebuilds from, so the restored game must
		// be able to serialise itself straight back out.
		out, err := fi.Snapshot()
		assert.NoError(t, err)
		assert.NotEmpty(t, out)
	})
	t.Run("invalid data", func(t *testing.T) {
		_, err := RestoreFourSeasonsInteractor([]byte("invalid"), newMockFourSeasonsPresenter())
		assert.Error(t, err)
	})
}
