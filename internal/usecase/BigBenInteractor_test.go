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

func newMockBigBenGame() *interfaces.MockBigBenGame {
	return new(interfaces.MockBigBenGame)
}

func newMockBigBenPresenter() *presenter.MockBigBenPresenter {
	return new(presenter.MockBigBenPresenter)
}

func TestNewBigBenInteractor(t *testing.T) {
	assert.NotNil(t, NewBigBenInteractor(newMockBigBenGame(), newMockBigBenPresenter()))
}

func TestNewBigBenInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewBigBenInteractor(nil, newMockBigBenPresenter()) })
	assert.Panics(t, func() { NewBigBenInteractor(newMockBigBenGame(), nil) })
}

func TestBigBenInteractorReset(t *testing.T) {
	g := newMockBigBenGame()
	p := newMockBigBenPresenter()
	i := NewBigBenInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every move method forwards its arguments and routes the error to the presenter.
func TestBigBenInteractorMoves(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*BigBenInteractor) string
	}{
		{"tableau to foundation", "MoveTableauToFoundation", []any{2, 7}, func(i *BigBenInteractor) string { return i.MoveTableauToFoundation(2, 7) }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 5}, func(i *BigBenInteractor) string { return i.MoveTableauToTableau(0, 5) }},
		{"autocomplete", "AutoComplete", nil, func(i *BigBenInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *BigBenInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *BigBenInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockBigBenGame()
			p := newMockBigBenPresenter()
			i := NewBigBenInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockBigBenGame()
			p := newMockBigBenPresenter()
			i := NewBigBenInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestBigBenInteractorGiveUp(t *testing.T) {
	g := newMockBigBenGame()
	p := newMockBigBenPresenter()
	i := NewBigBenInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestBigBenInteractorHintAndActionLog(t *testing.T) {
	g := newMockBigBenGame()
	p := newMockBigBenPresenter()
	i := NewBigBenInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestBigBenInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultBigBen()
	d.Reset()
	i := NewBigBenInteractor(d, newMockBigBenPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreBigBenInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultBigBen()
		d.Reset()
		p := newMockBigBenPresenter()
		data, err := NewBigBenInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreBigBenInteractor(data, p)
		require.NoError(t, err)
		require.NotNil(t, restored)
		// **枚数まで見る。**NotNil だけでは、盤が空でも通ってしまう。
		assert.Equal(t, d.GetStockCount(), restored.Game.GetStockCount())
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreBigBenInteractor([]byte("not json"), newMockBigBenPresenter())
		assert.Error(t, err)
	})
}
