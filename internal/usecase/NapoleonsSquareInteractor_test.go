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

func newMockNapoleonsSquareGame() *interfaces.MockNapoleonsSquareGame {
	return new(interfaces.MockNapoleonsSquareGame)
}

func newMockNapoleonsSquarePresenter() *presenter.MockNapoleonsSquarePresenter {
	return new(presenter.MockNapoleonsSquarePresenter)
}

func TestNewNapoleonsSquareInteractor(t *testing.T) {
	assert.NotNil(t, NewNapoleonsSquareInteractor(newMockNapoleonsSquareGame(), newMockNapoleonsSquarePresenter()))
}

func TestNewNapoleonsSquareInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewNapoleonsSquareInteractor(nil, newMockNapoleonsSquarePresenter()) })
	assert.Panics(t, func() { NewNapoleonsSquareInteractor(newMockNapoleonsSquareGame(), nil) })
}

func TestNapoleonsSquareInteractorReset(t *testing.T) {
	g := newMockNapoleonsSquareGame()
	p := newMockNapoleonsSquarePresenter()
	i := NewNapoleonsSquareInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every move method forwards its arguments and routes the error to the presenter.
func TestNapoleonsSquareInteractorMoves(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*NapoleonsSquareInteractor) string
	}{
		{"draw", "Draw", nil, func(i *NapoleonsSquareInteractor) string { return i.Draw() }},
		{"waste to tableau", "MoveWasteToTableau", []any{3}, func(i *NapoleonsSquareInteractor) string { return i.MoveWasteToTableau(3) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *NapoleonsSquareInteractor) string { return i.MoveWasteToFoundation() }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 2, 5}, func(i *NapoleonsSquareInteractor) string { return i.MoveTableauToTableau(0, 2, 5) }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{7}, func(i *NapoleonsSquareInteractor) string { return i.MoveTableauToFoundation(7) }},
		{"autocomplete", "AutoComplete", nil, func(i *NapoleonsSquareInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *NapoleonsSquareInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *NapoleonsSquareInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockNapoleonsSquareGame()
			p := newMockNapoleonsSquarePresenter()
			i := NewNapoleonsSquareInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockNapoleonsSquareGame()
			p := newMockNapoleonsSquarePresenter()
			i := NewNapoleonsSquareInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestNapoleonsSquareInteractorGiveUp(t *testing.T) {
	g := newMockNapoleonsSquareGame()
	p := newMockNapoleonsSquarePresenter()
	i := NewNapoleonsSquareInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestNapoleonsSquareInteractorHintAndActionLog(t *testing.T) {
	g := newMockNapoleonsSquareGame()
	p := newMockNapoleonsSquarePresenter()
	i := NewNapoleonsSquareInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestNapoleonsSquareInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultNapoleonsSquare()
	d.Reset()
	i := NewNapoleonsSquareInteractor(d, newMockNapoleonsSquarePresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreNapoleonsSquareInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultNapoleonsSquare()
		d.Reset()
		p := newMockNapoleonsSquarePresenter()
		data, err := NewNapoleonsSquareInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreNapoleonsSquareInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreNapoleonsSquareInteractor([]byte("not json"), newMockNapoleonsSquarePresenter())
		assert.Error(t, err)
	})
}
