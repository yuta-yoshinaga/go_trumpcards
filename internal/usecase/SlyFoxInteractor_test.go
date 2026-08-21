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

func newMockSlyFoxGame() *interfaces.MockSlyFoxGame {
	return new(interfaces.MockSlyFoxGame)
}

func newMockSlyFoxPresenter() *presenter.MockSlyFoxPresenter {
	return new(presenter.MockSlyFoxPresenter)
}

func TestNewSlyFoxInteractor(t *testing.T) {
	assert.NotNil(t, NewSlyFoxInteractor(newMockSlyFoxGame(), newMockSlyFoxPresenter()))
}

func TestNewSlyFoxInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewSlyFoxInteractor(nil, newMockSlyFoxPresenter()) })
	assert.Panics(t, func() { NewSlyFoxInteractor(newMockSlyFoxGame(), nil) })
}

func TestSlyFoxInteractorReset(t *testing.T) {
	g := newMockSlyFoxGame()
	p := newMockSlyFoxPresenter()
	i := NewSlyFoxInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestSlyFoxInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*SlyFoxInteractor) string
	}{
		{"deal to a slot", "DealToPile", []any{3}, func(i *SlyFoxInteractor) string { return i.DealToPile(3) }},
		{"deal to a foundation", "DealToFoundation", []any{1}, func(i *SlyFoxInteractor) string { return i.DealToFoundation(1) }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{3}, func(i *SlyFoxInteractor) string { return i.MoveTableauToFoundation(3) }},
		{"autocomplete", "AutoComplete", nil, func(i *SlyFoxInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *SlyFoxInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *SlyFoxInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockSlyFoxGame()
			p := newMockSlyFoxPresenter()
			i := NewSlyFoxInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockSlyFoxGame()
			p := newMockSlyFoxPresenter()
			i := NewSlyFoxInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestSlyFoxInteractorGiveUp(t *testing.T) {
	g := newMockSlyFoxGame()
	p := newMockSlyFoxPresenter()
	i := NewSlyFoxInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestSlyFoxInteractorHintAndActionLog(t *testing.T) {
	g := newMockSlyFoxGame()
	p := newMockSlyFoxPresenter()
	i := NewSlyFoxInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestSlyFoxInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultSlyFox()
	d.Reset()
	i := NewSlyFoxInteractor(d, newMockSlyFoxPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreSlyFoxInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultSlyFox()
		d.Reset()
		p := newMockSlyFoxPresenter()
		data, err := NewSlyFoxInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreSlyFoxInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreSlyFoxInteractor([]byte("not json"), newMockSlyFoxPresenter())
		assert.Error(t, err)
	})
}
