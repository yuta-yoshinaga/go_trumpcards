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

func newMockAmericanToadGame() *interfaces.MockAmericanToadGame {
	return new(interfaces.MockAmericanToadGame)
}

func newMockAmericanToadPresenter() *presenter.MockAmericanToadPresenter {
	return new(presenter.MockAmericanToadPresenter)
}

func TestNewAmericanToadInteractor(t *testing.T) {
	assert.NotNil(t, NewAmericanToadInteractor(newMockAmericanToadGame(), newMockAmericanToadPresenter()))
}

func TestNewAmericanToadInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewAmericanToadInteractor(nil, newMockAmericanToadPresenter()) })
	assert.Panics(t, func() { NewAmericanToadInteractor(newMockAmericanToadGame(), nil) })
}

func TestAmericanToadInteractorReset(t *testing.T) {
	g := newMockAmericanToadGame()
	p := newMockAmericanToadPresenter()
	i := NewAmericanToadInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestAmericanToadInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*AmericanToadInteractor) string
	}{
		{"draw", "Draw", nil, func(i *AmericanToadInteractor) string { return i.Draw() }},
		{"reserve to foundation", "MoveReserveToFoundation", nil, func(i *AmericanToadInteractor) string { return i.MoveReserveToFoundation() }},
		{"reserve to tableau", "MoveReserveToTableau", []any{3}, func(i *AmericanToadInteractor) string { return i.MoveReserveToTableau(3) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *AmericanToadInteractor) string { return i.MoveWasteToFoundation() }},
		{"waste to tableau", "MoveWasteToTableau", []any{2}, func(i *AmericanToadInteractor) string { return i.MoveWasteToTableau(2) }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{1}, func(i *AmericanToadInteractor) string { return i.MoveTableauToFoundation(1) }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 2, 5}, func(i *AmericanToadInteractor) string { return i.MoveTableauToTableau(0, 2, 5) }},
		{"autocomplete", "AutoComplete", nil, func(i *AmericanToadInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *AmericanToadInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *AmericanToadInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockAmericanToadGame()
			p := newMockAmericanToadPresenter()
			i := NewAmericanToadInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockAmericanToadGame()
			p := newMockAmericanToadPresenter()
			i := NewAmericanToadInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestAmericanToadInteractorGiveUp(t *testing.T) {
	g := newMockAmericanToadGame()
	p := newMockAmericanToadPresenter()
	i := NewAmericanToadInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestAmericanToadInteractorHintAndActionLog(t *testing.T) {
	g := newMockAmericanToadGame()
	p := newMockAmericanToadPresenter()
	i := NewAmericanToadInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestAmericanToadInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultAmericanToad()
	d.Reset()
	i := NewAmericanToadInteractor(d, newMockAmericanToadPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreAmericanToadInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultAmericanToad()
		d.Reset()
		p := newMockAmericanToadPresenter()
		data, err := NewAmericanToadInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreAmericanToadInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreAmericanToadInteractor([]byte("not json"), newMockAmericanToadPresenter())
		assert.Error(t, err)
	})
}
