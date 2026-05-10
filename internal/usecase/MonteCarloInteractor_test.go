package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockMonteCarloGame() *interfaces.MockMonteCarloGame {
	return new(interfaces.MockMonteCarloGame)
}

func newMockMonteCarloPresenter() *presenter.MockMonteCarloPresenter {
	return new(presenter.MockMonteCarloPresenter)
}

func TestNewMonteCarloInteractor(t *testing.T) {
	g := newMockMonteCarloGame()
	mp := newMockMonteCarloPresenter()
	mi := NewMonteCarloInteractor(g, mp)
	assert.NotNil(t, mi)
}

func TestNewMonteCarloInteractor_PanicsOnNil(t *testing.T) {
	mp := newMockMonteCarloPresenter()
	assert.Panics(t, func() { NewMonteCarloInteractor(nil, mp) })
	g := newMockMonteCarloGame()
	assert.Panics(t, func() { NewMonteCarloInteractor(g, nil) })
}

func TestMonteCarloInteractor_Reset(t *testing.T) {
	g := newMockMonteCarloGame()
	mp := newMockMonteCarloPresenter()
	mi := NewMonteCarloInteractor(g, mp)

	g.On("Reset").Return()
	mp.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", mi.Reset())
	g.AssertCalled(t, "Reset")
}

func TestMonteCarloInteractor_Remove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockMonteCarloGame()
		mp := newMockMonteCarloPresenter()
		mi := NewMonteCarloInteractor(g, mp)

		g.On("Remove", 0, 0, 0, 1).Return(nil)
		mp.On("Output", g, nil).Return("remove_output")

		assert.Equal(t, "remove_output", mi.Remove(0, 0, 0, 1))
	})

	t.Run("error", func(t *testing.T) {
		g := newMockMonteCarloGame()
		mp := newMockMonteCarloPresenter()
		mi := NewMonteCarloInteractor(g, mp)

		err := errors.New("not adjacent")
		g.On("Remove", 0, 0, 2, 2).Return(err)
		mp.On("Output", g, err).Return("err_output")

		assert.Equal(t, "err_output", mi.Remove(0, 0, 2, 2))
	})
}

func TestMonteCarloInteractor_Deal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockMonteCarloGame()
		mp := newMockMonteCarloPresenter()
		mi := NewMonteCarloInteractor(g, mp)

		g.On("Deal").Return(nil)
		mp.On("Output", g, nil).Return("deal_output")

		assert.Equal(t, "deal_output", mi.Deal())
	})

	t.Run("error", func(t *testing.T) {
		g := newMockMonteCarloGame()
		mp := newMockMonteCarloPresenter()
		mi := NewMonteCarloInteractor(g, mp)

		err := errors.New("wrong phase")
		g.On("Deal").Return(err)
		mp.On("Output", g, err).Return("deal_err")

		assert.Equal(t, "deal_err", mi.Deal())
	})
}

func TestMonteCarloInteractor_Undo(t *testing.T) {
	g := newMockMonteCarloGame()
	mp := newMockMonteCarloPresenter()
	mi := NewMonteCarloInteractor(g, mp)

	g.On("Undo").Return(nil)
	mp.On("Output", g, nil).Return("undo_output")

	assert.Equal(t, "undo_output", mi.Undo())
}

func TestMonteCarloInteractor_GiveUp(t *testing.T) {
	g := newMockMonteCarloGame()
	mp := newMockMonteCarloPresenter()
	mi := NewMonteCarloInteractor(g, mp)

	g.On("GiveUp").Return()
	mp.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", mi.GiveUp())
}

func TestMonteCarloInteractor_Hint(t *testing.T) {
	g := newMockMonteCarloGame()
	mp := newMockMonteCarloPresenter()
	mi := NewMonteCarloInteractor(g, mp)

	mp.On("HintOutput", g).Return("hint_output")

	assert.Equal(t, "hint_output", mi.Hint())
}

func TestMonteCarloInteractor_ActionLog(t *testing.T) {
	g := newMockMonteCarloGame()
	mp := newMockMonteCarloPresenter()
	mi := NewMonteCarloInteractor(g, mp)

	mp.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "log_output", mi.ActionLog())
}

func TestRestoreMonteCarloInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"bd":[[null,null,null,null,null],[null,null,null,null,null],[null,null,null,null,null],[null,null,null,null,null],[null,null,null,null,null]],"ps":0,"rc":0,"dc":0,"sl":false,"al":[]}`)
		mp := newMockMonteCarloPresenter()
		mi, err := RestoreMonteCarloInteractor(data, mp)
		assert.NoError(t, err)
		assert.NotNil(t, mi)
	})

	t.Run("invalid data", func(t *testing.T) {
		mp := newMockMonteCarloPresenter()
		_, err := RestoreMonteCarloInteractor([]byte("invalid"), mp)
		assert.Error(t, err)
	})
}
