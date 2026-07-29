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

func newMockMissMilliganGame() *interfaces.MockMissMilliganGame {
	return new(interfaces.MockMissMilliganGame)
}

func newMockMissMilliganPresenter() *presenter.MockMissMilliganPresenter {
	return new(presenter.MockMissMilliganPresenter)
}

func TestNewMissMilliganInteractor(t *testing.T) {
	assert.NotNil(t, NewMissMilliganInteractor(newMockMissMilliganGame(), newMockMissMilliganPresenter()))
}

func TestNewMissMilliganInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewMissMilliganInteractor(nil, newMockMissMilliganPresenter()) })
	assert.Panics(t, func() { NewMissMilliganInteractor(newMockMissMilliganGame(), nil) })
}

func TestMissMilliganInteractorReset(t *testing.T) {
	g := newMockMissMilliganGame()
	p := newMockMissMilliganPresenter()
	i := NewMissMilliganInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestMissMilliganInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*MissMilliganInteractor) string
	}{
		{"deal", "Deal", nil, func(i *MissMilliganInteractor) string { return i.Deal() }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 2, 5}, func(i *MissMilliganInteractor) string { return i.MoveTableauToTableau(0, 2, 5) }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{7}, func(i *MissMilliganInteractor) string { return i.MoveTableauToFoundation(7) }},
		{"waive", "Waive", []any{3, -1}, func(i *MissMilliganInteractor) string { return i.Waive(3, -1) }},
		{"place waived", "PlaceWaived", []any{4}, func(i *MissMilliganInteractor) string { return i.PlaceWaived(4) }},
		{"waived to foundation", "MoveWaivedToFoundation", nil, func(i *MissMilliganInteractor) string { return i.MoveWaivedToFoundation() }},
		{"autocomplete", "AutoComplete", nil, func(i *MissMilliganInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *MissMilliganInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *MissMilliganInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockMissMilliganGame()
			p := newMockMissMilliganPresenter()
			i := NewMissMilliganInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockMissMilliganGame()
			p := newMockMissMilliganPresenter()
			i := NewMissMilliganInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestMissMilliganInteractorGiveUp(t *testing.T) {
	g := newMockMissMilliganGame()
	p := newMockMissMilliganPresenter()
	i := NewMissMilliganInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestMissMilliganInteractorHintAndActionLog(t *testing.T) {
	g := newMockMissMilliganGame()
	p := newMockMissMilliganPresenter()
	i := NewMissMilliganInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestMissMilliganInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultMissMilligan()
	d.Reset()
	i := NewMissMilliganInteractor(d, newMockMissMilliganPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreMissMilliganInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultMissMilligan()
		d.Reset()
		p := newMockMissMilliganPresenter()
		data, err := NewMissMilliganInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreMissMilliganInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreMissMilliganInteractor([]byte("not json"), newMockMissMilliganPresenter())
		assert.Error(t, err)
	})
}
