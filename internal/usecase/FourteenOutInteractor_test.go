package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockFourteenOutGame() *interfaces.MockFourteenOutGame {
	return new(interfaces.MockFourteenOutGame)
}

func newMockFourteenOutPresenter() *presenter.MockFourteenOutPresenter {
	return new(presenter.MockFourteenOutPresenter)
}

func TestNewFourteenOutInteractor(t *testing.T) {
	g := newMockFourteenOutGame()
	mp := newMockFourteenOutPresenter()
	mi := NewFourteenOutInteractor(g, mp)
	assert.NotNil(t, mi)
}

func TestNewFourteenOutInteractor_PanicsOnNil(t *testing.T) {
	mp := newMockFourteenOutPresenter()
	assert.Panics(t, func() { NewFourteenOutInteractor(nil, mp) })
	g := newMockFourteenOutGame()
	assert.Panics(t, func() { NewFourteenOutInteractor(g, nil) })
}

func TestFourteenOutInteractor_Reset(t *testing.T) {
	g := newMockFourteenOutGame()
	mp := newMockFourteenOutPresenter()
	mi := NewFourteenOutInteractor(g, mp)

	g.On("Reset").Return()
	mp.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", mi.Reset())
	g.AssertCalled(t, "Reset")
}

func TestFourteenOutInteractor_Remove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockFourteenOutGame()
		mp := newMockFourteenOutPresenter()
		mi := NewFourteenOutInteractor(g, mp)

		g.On("Remove", 0, 1).Return(nil)
		mp.On("Output", g, nil).Return("remove_output")

		assert.Equal(t, "remove_output", mi.Remove(0, 1))
	})

	t.Run("error", func(t *testing.T) {
		g := newMockFourteenOutGame()
		mp := newMockFourteenOutPresenter()
		mi := NewFourteenOutInteractor(g, mp)

		err := errors.New("cards do not sum to 14")
		g.On("Remove", 0, 2).Return(err)
		mp.On("Output", g, err).Return("err_output")

		assert.Equal(t, "err_output", mi.Remove(0, 2))
	})
}

func TestFourteenOutInteractor_Undo(t *testing.T) {
	g := newMockFourteenOutGame()
	mp := newMockFourteenOutPresenter()
	mi := NewFourteenOutInteractor(g, mp)

	g.On("Undo").Return(nil)
	mp.On("Output", g, nil).Return("undo_output")

	assert.Equal(t, "undo_output", mi.Undo())
}

func TestFourteenOutInteractor_GiveUp(t *testing.T) {
	g := newMockFourteenOutGame()
	mp := newMockFourteenOutPresenter()
	mi := NewFourteenOutInteractor(g, mp)

	g.On("GiveUp").Return()
	mp.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", mi.GiveUp())
}

func TestFourteenOutInteractor_Hint(t *testing.T) {
	g := newMockFourteenOutGame()
	mp := newMockFourteenOutPresenter()
	mi := NewFourteenOutInteractor(g, mp)

	mp.On("HintOutput", g).Return("hint_output")

	assert.Equal(t, "hint_output", mi.Hint())
}

func TestFourteenOutInteractor_ActionLog(t *testing.T) {
	g := newMockFourteenOutGame()
	mp := newMockFourteenOutPresenter()
	mi := NewFourteenOutInteractor(g, mp)

	mp.On("ActionLogOutput", g).Return("log_output")

	assert.Equal(t, "log_output", mi.ActionLog())
}

func TestRestoreFourteenOutInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"bd":[[null,null,null,null,null],[null,null,null,null,null],[null,null,null,null,null],[null,null,null,null,null],[null,null,null,null,null]],"ps":0,"rc":0,"dc":0,"sl":false,"al":[]}`)
		mp := newMockFourteenOutPresenter()
		mi, err := RestoreFourteenOutInteractor(data, mp)
		assert.NoError(t, err)
		assert.NotNil(t, mi)
	})

	t.Run("invalid data", func(t *testing.T) {
		mp := newMockFourteenOutPresenter()
		_, err := RestoreFourteenOutInteractor([]byte("invalid"), mp)
		assert.Error(t, err)
	})
}
