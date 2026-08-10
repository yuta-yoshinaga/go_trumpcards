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

func newMockColoradoGame() *interfaces.MockColoradoGame {
	return new(interfaces.MockColoradoGame)
}

func newMockColoradoPresenter() *presenter.MockColoradoPresenter {
	return new(presenter.MockColoradoPresenter)
}

func TestNewColoradoInteractor(t *testing.T) {
	assert.NotNil(t, NewColoradoInteractor(newMockColoradoGame(), newMockColoradoPresenter()))
}

func TestNewColoradoInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewColoradoInteractor(nil, newMockColoradoPresenter()) })
	assert.Panics(t, func() { NewColoradoInteractor(newMockColoradoGame(), nil) })
}

func TestColoradoInteractorReset(t *testing.T) {
	g := newMockColoradoGame()
	p := newMockColoradoPresenter()
	i := NewColoradoInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestColoradoInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*ColoradoInteractor) string
	}{
		{"draw", "Draw", nil, func(i *ColoradoInteractor) string { return i.Draw() }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{3}, func(i *ColoradoInteractor) string { return i.MoveTableauToFoundation(3) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *ColoradoInteractor) string { return i.MoveWasteToFoundation() }},
		{"waste to tableau", "MoveWasteToTableau", []any{2}, func(i *ColoradoInteractor) string { return i.MoveWasteToTableau(2) }},
		{"stock to tableau", "MoveStockToTableau", []any{4}, func(i *ColoradoInteractor) string { return i.MoveStockToTableau(4) }},
		{"autocomplete", "AutoComplete", nil, func(i *ColoradoInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *ColoradoInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *ColoradoInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockColoradoGame()
			p := newMockColoradoPresenter()
			i := NewColoradoInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockColoradoGame()
			p := newMockColoradoPresenter()
			i := NewColoradoInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestColoradoInteractorGiveUp(t *testing.T) {
	g := newMockColoradoGame()
	p := newMockColoradoPresenter()
	i := NewColoradoInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestColoradoInteractorHintAndActionLog(t *testing.T) {
	g := newMockColoradoGame()
	p := newMockColoradoPresenter()
	i := NewColoradoInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestColoradoInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultColorado()
	d.Reset()
	i := NewColoradoInteractor(d, newMockColoradoPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreColoradoInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultColorado()
		d.Reset()
		p := newMockColoradoPresenter()
		data, err := NewColoradoInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreColoradoInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreColoradoInteractor([]byte("not json"), newMockColoradoPresenter())
		assert.Error(t, err)
	})
}
