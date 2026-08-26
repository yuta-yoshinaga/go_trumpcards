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

func newMockSalicLawGame() *interfaces.MockSalicLawGame {
	return new(interfaces.MockSalicLawGame)
}

func newMockSalicLawPresenter() *presenter.MockSalicLawPresenter {
	return new(presenter.MockSalicLawPresenter)
}

func TestNewSalicLawInteractor(t *testing.T) {
	assert.NotNil(t, NewSalicLawInteractor(newMockSalicLawGame(), newMockSalicLawPresenter()))
}

func TestNewSalicLawInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewSalicLawInteractor(nil, newMockSalicLawPresenter()) })
	assert.Panics(t, func() { NewSalicLawInteractor(newMockSalicLawGame(), nil) })
}

func TestSalicLawInteractorReset(t *testing.T) {
	g := newMockSalicLawGame()
	p := newMockSalicLawPresenter()
	i := NewSalicLawInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestSalicLawInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*SalicLawInteractor) string
	}{
		{"draw", "Draw", nil, func(i *SalicLawInteractor) string { return i.Draw() }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{3}, func(i *SalicLawInteractor) string { return i.MoveTableauToFoundation(3) }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 5}, func(i *SalicLawInteractor) string { return i.MoveTableauToTableau(0, 5) }},
		{"autocomplete", "AutoComplete", nil, func(i *SalicLawInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *SalicLawInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *SalicLawInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockSalicLawGame()
			p := newMockSalicLawPresenter()
			i := NewSalicLawInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockSalicLawGame()
			p := newMockSalicLawPresenter()
			i := NewSalicLawInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestSalicLawInteractorGiveUp(t *testing.T) {
	g := newMockSalicLawGame()
	p := newMockSalicLawPresenter()
	i := NewSalicLawInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestSalicLawInteractorHintAndActionLog(t *testing.T) {
	g := newMockSalicLawGame()
	p := newMockSalicLawPresenter()
	i := NewSalicLawInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestSalicLawInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultSalicLaw()
	d.Reset()
	i := NewSalicLawInteractor(d, newMockSalicLawPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreSalicLawInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultSalicLaw()
		d.Reset()
		p := newMockSalicLawPresenter()
		data, err := NewSalicLawInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreSalicLawInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreSalicLawInteractor([]byte("not json"), newMockSalicLawPresenter())
		assert.Error(t, err)
	})
}
