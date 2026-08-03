package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockGrandfathersClockGame() *interfaces.MockGrandfathersClockGame {
	return new(interfaces.MockGrandfathersClockGame)
}

func newMockGrandfathersClockPresenter() *presenter.MockGrandfathersClockPresenter {
	return new(presenter.MockGrandfathersClockPresenter)
}

func TestNewGrandfathersClockInteractor(t *testing.T) {
	assert.NotNil(t, NewGrandfathersClockInteractor(newMockGrandfathersClockGame(), newMockGrandfathersClockPresenter()))
}

func TestNewGrandfathersClockInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewGrandfathersClockInteractor(nil, newMockGrandfathersClockPresenter()) })
	assert.Panics(t, func() { NewGrandfathersClockInteractor(newMockGrandfathersClockGame(), nil) })
}

func TestGrandfathersClockInteractorReset(t *testing.T) {
	g := newMockGrandfathersClockGame()
	p := newMockGrandfathersClockPresenter()
	i := NewGrandfathersClockInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every move method forwards its arguments and routes the error to the presenter.
func TestGrandfathersClockInteractorMoves(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*GrandfathersClockInteractor) string
	}{
		{"tableau to foundation", "MoveTableauToFoundation", []any{2, 7}, func(i *GrandfathersClockInteractor) string { return i.MoveTableauToFoundation(2, 7) }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 5}, func(i *GrandfathersClockInteractor) string { return i.MoveTableauToTableau(0, 5) }},
		{"autocomplete", "AutoComplete", nil, func(i *GrandfathersClockInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *GrandfathersClockInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *GrandfathersClockInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockGrandfathersClockGame()
			p := newMockGrandfathersClockPresenter()
			i := NewGrandfathersClockInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockGrandfathersClockGame()
			p := newMockGrandfathersClockPresenter()
			i := NewGrandfathersClockInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestGrandfathersClockInteractorGiveUp(t *testing.T) {
	g := newMockGrandfathersClockGame()
	p := newMockGrandfathersClockPresenter()
	i := NewGrandfathersClockInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestGrandfathersClockInteractorHintAndActionLog(t *testing.T) {
	g := newMockGrandfathersClockGame()
	p := newMockGrandfathersClockPresenter()
	i := NewGrandfathersClockInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestGrandfathersClockInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultGrandfathersClock()
	d.Reset()
	i := NewGrandfathersClockInteractor(d, newMockGrandfathersClockPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreGrandfathersClockInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultGrandfathersClock()
		d.Reset()
		p := newMockGrandfathersClockPresenter()
		data, err := NewGrandfathersClockInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreGrandfathersClockInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreGrandfathersClockInteractor([]byte("not json"), newMockGrandfathersClockPresenter())
		assert.Error(t, err)
	})
}
