package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockTrashGame() *interfaces.MockTrashGame {
	return new(interfaces.MockTrashGame)
}

func newMockTrashPresenter() *presenter.MockTrashPresenter {
	return new(presenter.MockTrashPresenter)
}

func TestNewTrashInteractor(t *testing.T) {
	tg := newMockTrashGame()
	tp := newMockTrashPresenter()
	ti := NewTrashInteractor(tg, tp)
	assert.NotNil(t, ti)
}

func TestNewTrashInteractorPanicsOnNil(t *testing.T) {
	tp := newMockTrashPresenter()
	assert.Panics(t, func() { NewTrashInteractor(nil, tp) })
	tg := newMockTrashGame()
	assert.Panics(t, func() { NewTrashInteractor(tg, nil) })
}

func TestTrashInteractorReset(t *testing.T) {
	tg := newMockTrashGame()
	tp := newMockTrashPresenter()
	ti := NewTrashInteractor(tg, tp)

	tg.On("Reset").Return()
	tp.On("Output", tg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", ti.Reset())
	tg.AssertCalled(t, "Reset")
}

func TestTrashInteractorHint(t *testing.T) {
	tg := newMockTrashGame()
	tp := newMockTrashPresenter()
	ti := NewTrashInteractor(tg, tp)

	tp.On("HintOutput", tg).Return("hint_output")

	assert.Equal(t, "hint_output", ti.Hint())
	tp.AssertCalled(t, "HintOutput", tg)
}

func TestTrashInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tg := newMockTrashGame()
		tp := newMockTrashPresenter()
		ti := NewTrashInteractor(tg, tp)

		tg.On("Draw").Return(nil)
		tp.On("Output", tg, nil).Return("draw_output")

		assert.Equal(t, "draw_output", ti.Draw())
	})

	t.Run("error", func(t *testing.T) {
		tg := newMockTrashGame()
		tp := newMockTrashPresenter()
		ti := NewTrashInteractor(tg, tp)

		err := errors.New("no cards")
		tg.On("Draw").Return(err)
		tp.On("Output", tg, err).Return("err_output")

		assert.Equal(t, "err_output", ti.Draw())
	})
}

func TestTrashInteractorPlaceWild(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tg := newMockTrashGame()
		tp := newMockTrashPresenter()
		ti := NewTrashInteractor(tg, tp)

		tg.On("PlaceWild", 3).Return(nil)
		tp.On("Output", tg, nil).Return("place_output")

		assert.Equal(t, "place_output", ti.PlaceWild(3))
	})

	t.Run("error", func(t *testing.T) {
		tg := newMockTrashGame()
		tp := newMockTrashPresenter()
		ti := NewTrashInteractor(tg, tp)

		err := errors.New("invalid position")
		tg.On("PlaceWild", 0).Return(err)
		tp.On("Output", tg, err).Return("err_output")

		assert.Equal(t, "err_output", ti.PlaceWild(0))
	})
}

func TestTrashInteractorCpuStep(t *testing.T) {
	tg := newMockTrashGame()
	tp := newMockTrashPresenter()
	ti := NewTrashInteractor(tg, tp)

	tg.On("CpuStep").Return(nil)
	tp.On("Output", tg, nil).Return("cpu_output")

	assert.Equal(t, "cpu_output", ti.CpuStep())
}

func TestTrashInteractorActionLog(t *testing.T) {
	tg := newMockTrashGame()
	tp := newMockTrashPresenter()
	ti := NewTrashInteractor(tg, tp)

	tp.On("ActionLogOutput", tg).Return("log_output")

	assert.Equal(t, "log_output", ti.ActionLog())
}

func TestRestoreTrashInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"st":[],"ds":[],"pl":[{"sl":[]},{"sl":[]}],"cu":0,"ph":0,"pe":null,"mc":0,"wn":-1,"al":[]}`)
		tp := newMockTrashPresenter()
		ti, err := RestoreTrashInteractor(data, tp)
		assert.NoError(t, err)
		assert.NotNil(t, ti)
	})

	t.Run("invalid data", func(t *testing.T) {
		tp := newMockTrashPresenter()
		_, err := RestoreTrashInteractor([]byte("bogus"), tp)
		assert.Error(t, err)
	})
}
