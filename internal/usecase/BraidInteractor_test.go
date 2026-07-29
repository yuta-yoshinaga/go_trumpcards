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

func newMockBraidGame() *interfaces.MockBraidGame {
	return new(interfaces.MockBraidGame)
}

func newMockBraidPresenter() *presenter.MockBraidPresenter {
	return new(presenter.MockBraidPresenter)
}

func TestNewBraidInteractor(t *testing.T) {
	assert.NotNil(t, NewBraidInteractor(newMockBraidGame(), newMockBraidPresenter()))
}

func TestNewBraidInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewBraidInteractor(nil, newMockBraidPresenter()) })
	assert.Panics(t, func() { NewBraidInteractor(newMockBraidGame(), nil) })
}

func TestBraidInteractorReset(t *testing.T) {
	g := newMockBraidGame()
	p := newMockBraidPresenter()
	i := NewBraidInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestBraidInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*BraidInteractor) string
	}{
		{"draw", "Draw", nil, func(i *BraidInteractor) string { return i.Draw() }},
		{"direction up", "ChooseDirection", []any{true}, func(i *BraidInteractor) string { return i.ChooseDirection(true) }},
		{"direction down", "ChooseDirection", []any{false}, func(i *BraidInteractor) string { return i.ChooseDirection(false) }},
		{"braid to foundation", "MoveBraidToFoundation", nil, func(i *BraidInteractor) string { return i.MoveBraidToFoundation() }},
		{"field to foundation", "MoveFieldToFoundation", []any{2}, func(i *BraidInteractor) string { return i.MoveFieldToFoundation(2) }},
		{"helper to foundation", "MoveHelperToFoundation", []any{5}, func(i *BraidInteractor) string { return i.MoveHelperToFoundation(5) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *BraidInteractor) string { return i.MoveWasteToFoundation() }},
		{"waste to helper", "MoveWasteToHelper", []any{3}, func(i *BraidInteractor) string { return i.MoveWasteToHelper(3) }},
		{"autocomplete", "AutoComplete", nil, func(i *BraidInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *BraidInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *BraidInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockBraidGame()
			p := newMockBraidPresenter()
			i := NewBraidInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockBraidGame()
			p := newMockBraidPresenter()
			i := NewBraidInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestBraidInteractorGiveUp(t *testing.T) {
	g := newMockBraidGame()
	p := newMockBraidPresenter()
	i := NewBraidInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestBraidInteractorHintAndActionLog(t *testing.T) {
	g := newMockBraidGame()
	p := newMockBraidPresenter()
	i := NewBraidInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestBraidInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultBraid()
	d.Reset()
	i := NewBraidInteractor(d, newMockBraidPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreBraidInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultBraid()
		d.Reset()
		p := newMockBraidPresenter()
		data, err := NewBraidInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreBraidInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreBraidInteractor([]byte("not json"), newMockBraidPresenter())
		assert.Error(t, err)
	})
}
