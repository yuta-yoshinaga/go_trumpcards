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

func newMockRoyalCotillionGame() *interfaces.MockRoyalCotillionGame {
	return new(interfaces.MockRoyalCotillionGame)
}

func newMockRoyalCotillionPresenter() *presenter.MockRoyalCotillionPresenter {
	return new(presenter.MockRoyalCotillionPresenter)
}

func TestNewRoyalCotillionInteractor(t *testing.T) {
	assert.NotNil(t, NewRoyalCotillionInteractor(newMockRoyalCotillionGame(), newMockRoyalCotillionPresenter()))
}

func TestNewRoyalCotillionInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewRoyalCotillionInteractor(nil, newMockRoyalCotillionPresenter()) })
	assert.Panics(t, func() { NewRoyalCotillionInteractor(newMockRoyalCotillionGame(), nil) })
}

func TestRoyalCotillionInteractorReset(t *testing.T) {
	g := newMockRoyalCotillionGame()
	p := newMockRoyalCotillionPresenter()
	i := NewRoyalCotillionInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestRoyalCotillionInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*RoyalCotillionInteractor) string
	}{
		{"draw", "Draw", nil, func(i *RoyalCotillionInteractor) string { return i.Draw() }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{3}, func(i *RoyalCotillionInteractor) string { return i.MoveTableauToFoundation(3) }},
		{"reserve to foundation", "MoveReserveToFoundation", []any{2}, func(i *RoyalCotillionInteractor) string { return i.MoveReserveToFoundation(2) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *RoyalCotillionInteractor) string { return i.MoveWasteToFoundation() }},
		{"waste to tableau", "MoveWasteToTableau", []any{2}, func(i *RoyalCotillionInteractor) string { return i.MoveWasteToTableau(2) }},
		{"stock to tableau", "MoveStockToTableau", []any{4}, func(i *RoyalCotillionInteractor) string { return i.MoveStockToTableau(4) }},
		{"autocomplete", "AutoComplete", nil, func(i *RoyalCotillionInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *RoyalCotillionInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *RoyalCotillionInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockRoyalCotillionGame()
			p := newMockRoyalCotillionPresenter()
			i := NewRoyalCotillionInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockRoyalCotillionGame()
			p := newMockRoyalCotillionPresenter()
			i := NewRoyalCotillionInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestRoyalCotillionInteractorGiveUp(t *testing.T) {
	g := newMockRoyalCotillionGame()
	p := newMockRoyalCotillionPresenter()
	i := NewRoyalCotillionInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestRoyalCotillionInteractorHintAndActionLog(t *testing.T) {
	g := newMockRoyalCotillionGame()
	p := newMockRoyalCotillionPresenter()
	i := NewRoyalCotillionInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestRoyalCotillionInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultRoyalCotillion()
	d.Reset()
	i := NewRoyalCotillionInteractor(d, newMockRoyalCotillionPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreRoyalCotillionInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultRoyalCotillion()
		d.Reset()
		p := newMockRoyalCotillionPresenter()
		data, err := NewRoyalCotillionInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreRoyalCotillionInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreRoyalCotillionInteractor([]byte("not json"), newMockRoyalCotillionPresenter())
		assert.Error(t, err)
	})
}
