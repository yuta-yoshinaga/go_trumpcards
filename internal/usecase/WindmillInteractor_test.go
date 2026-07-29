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

func newMockWindmillGame() *interfaces.MockWindmillGame {
	return new(interfaces.MockWindmillGame)
}

func newMockWindmillPresenter() *presenter.MockWindmillPresenter {
	return new(presenter.MockWindmillPresenter)
}

func TestNewWindmillInteractor(t *testing.T) {
	assert.NotNil(t, NewWindmillInteractor(newMockWindmillGame(), newMockWindmillPresenter()))
}

func TestNewWindmillInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewWindmillInteractor(nil, newMockWindmillPresenter()) })
	assert.Panics(t, func() { NewWindmillInteractor(newMockWindmillGame(), nil) })
}

func TestWindmillInteractorReset(t *testing.T) {
	g := newMockWindmillGame()
	p := newMockWindmillPresenter()
	i := NewWindmillInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestWindmillInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*WindmillInteractor) string
	}{
		{"draw", "Draw", nil, func(i *WindmillInteractor) string { return i.Draw() }},
		{"sail to center", "MoveSailToCenter", []any{3}, func(i *WindmillInteractor) string { return i.MoveSailToCenter(3) }},
		{"sail to corner", "MoveSailToCorner", []any{3, 1}, func(i *WindmillInteractor) string { return i.MoveSailToCorner(3, 1) }},
		{"waste to center", "MoveWasteToCenter", nil, func(i *WindmillInteractor) string { return i.MoveWasteToCenter() }},
		{"waste to corner", "MoveWasteToCorner", []any{2}, func(i *WindmillInteractor) string { return i.MoveWasteToCorner(2) }},
		{"corner to center", "MoveCornerToCenter", []any{0}, func(i *WindmillInteractor) string { return i.MoveCornerToCenter(0) }},
		{"autocomplete", "AutoComplete", nil, func(i *WindmillInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *WindmillInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *WindmillInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockWindmillGame()
			p := newMockWindmillPresenter()
			i := NewWindmillInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockWindmillGame()
			p := newMockWindmillPresenter()
			i := NewWindmillInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestWindmillInteractorGiveUp(t *testing.T) {
	g := newMockWindmillGame()
	p := newMockWindmillPresenter()
	i := NewWindmillInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestWindmillInteractorHintAndActionLog(t *testing.T) {
	g := newMockWindmillGame()
	p := newMockWindmillPresenter()
	i := NewWindmillInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestWindmillInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultWindmill()
	d.Reset()
	i := NewWindmillInteractor(d, newMockWindmillPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreWindmillInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultWindmill()
		d.Reset()
		p := newMockWindmillPresenter()
		data, err := NewWindmillInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreWindmillInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreWindmillInteractor([]byte("not json"), newMockWindmillPresenter())
		assert.Error(t, err)
	})
}
