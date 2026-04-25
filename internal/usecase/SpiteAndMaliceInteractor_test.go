package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockSpiteAndMaliceGame() *interfaces.MockSpiteAndMaliceGame {
	return new(interfaces.MockSpiteAndMaliceGame)
}

func newMockSpiteAndMalicePresenter() *presenter.MockSpiteAndMalicePresenter {
	return new(presenter.MockSpiteAndMalicePresenter)
}

func TestNewSpiteAndMaliceInteractor(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	assert.NotNil(t, NewSpiteAndMaliceInteractor(g, p))
}

func TestNewSpiteAndMaliceInteractor_PanicsOnNil(t *testing.T) {
	p := newMockSpiteAndMalicePresenter()
	assert.Panics(t, func() { NewSpiteAndMaliceInteractor(nil, p) })
	g := newMockSpiteAndMaliceGame()
	assert.Panics(t, func() { NewSpiteAndMaliceInteractor(g, nil) })
}

func TestSpiteAndMaliceInteractor_Reset(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	i := NewSpiteAndMaliceInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_out")

	assert.Equal(t, "reset_out", i.Reset())
	g.AssertCalled(t, "Reset")
}

func TestSpiteAndMaliceInteractor_PlayFromHand(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockSpiteAndMaliceGame()
		p := newMockSpiteAndMalicePresenter()
		i := NewSpiteAndMaliceInteractor(g, p)
		g.On("PlayFromHand", 1, 2).Return(nil)
		p.On("Output", g, nil).Return("ok")
		assert.Equal(t, "ok", i.PlayFromHand(1, 2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockSpiteAndMaliceGame()
		p := newMockSpiteAndMalicePresenter()
		i := NewSpiteAndMaliceInteractor(g, p)
		err := errors.New("bad")
		g.On("PlayFromHand", 0, 0).Return(err)
		p.On("Output", g, err).Return("err")
		assert.Equal(t, "err", i.PlayFromHand(0, 0))
	})
}

func TestSpiteAndMaliceInteractor_PlayFromGoal(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	i := NewSpiteAndMaliceInteractor(g, p)
	g.On("PlayFromGoal", 3).Return(nil)
	p.On("Output", g, nil).Return("g_out")
	assert.Equal(t, "g_out", i.PlayFromGoal(3))
}

func TestSpiteAndMaliceInteractor_PlayFromSide(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	i := NewSpiteAndMaliceInteractor(g, p)
	g.On("PlayFromSide", 0, 1).Return(nil)
	p.On("Output", g, nil).Return("s_out")
	assert.Equal(t, "s_out", i.PlayFromSide(0, 1))
}

func TestSpiteAndMaliceInteractor_Discard(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	i := NewSpiteAndMaliceInteractor(g, p)
	g.On("Discard", 2, 1).Return(nil)
	p.On("Output", g, nil).Return("d_out")
	assert.Equal(t, "d_out", i.Discard(2, 1))
}

func TestSpiteAndMaliceInteractor_CpuStep(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	i := NewSpiteAndMaliceInteractor(g, p)
	g.On("CpuStep").Return(nil)
	p.On("Output", g, nil).Return("cpu_out")
	assert.Equal(t, "cpu_out", i.CpuStep())
}

func TestSpiteAndMaliceInteractor_Hint(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	i := NewSpiteAndMaliceInteractor(g, p)
	p.On("HintOutput", g).Return("hint_out")
	assert.Equal(t, "hint_out", i.Hint())
}

func TestSpiteAndMaliceInteractor_ActionLog(t *testing.T) {
	g := newMockSpiteAndMaliceGame()
	p := newMockSpiteAndMalicePresenter()
	i := NewSpiteAndMaliceInteractor(g, p)
	p.On("ActionLogOutput", g).Return("log")
	assert.Equal(t, "log", i.ActionLog())
}

func TestRestoreSpiteAndMaliceInteractor(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		data := []byte(`{"ph":0,"cu":0,"mc":0,"wn":-1,"cf":{"gs":20,"cd":1}}`)
		p := newMockSpiteAndMalicePresenter()
		i, err := RestoreSpiteAndMaliceInteractor(data, p)
		assert.NoError(t, err)
		assert.NotNil(t, i)
	})
	t.Run("invalid", func(t *testing.T) {
		p := newMockSpiteAndMalicePresenter()
		_, err := RestoreSpiteAndMaliceInteractor([]byte("bogus"), p)
		assert.Error(t, err)
	})
}
