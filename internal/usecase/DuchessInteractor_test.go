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

func newMockDuchessGame() *interfaces.MockDuchessGame {
	return new(interfaces.MockDuchessGame)
}

func newMockDuchessPresenter() *presenter.MockDuchessPresenter {
	return new(presenter.MockDuchessPresenter)
}

func TestNewDuchessInteractor(t *testing.T) {
	assert.NotNil(t, NewDuchessInteractor(newMockDuchessGame(), newMockDuchessPresenter()))
}

func TestNewDuchessInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewDuchessInteractor(nil, newMockDuchessPresenter()) })
	assert.Panics(t, func() { NewDuchessInteractor(newMockDuchessGame(), nil) })
}

func TestDuchessInteractorReset(t *testing.T) {
	g := newMockDuchessGame()
	p := newMockDuchessPresenter()
	i := NewDuchessInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestDuchessInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*DuchessInteractor) string
	}{
		{"choose base rank", "ChooseBaseRank", []any{2}, func(i *DuchessInteractor) string { return i.ChooseBaseRank(2) }},
		{"draw", "Draw", nil, func(i *DuchessInteractor) string { return i.Draw() }},
		{"reserve to foundation", "MoveReserveToFoundation", []any{1}, func(i *DuchessInteractor) string { return i.MoveReserveToFoundation(1) }},
		{"reserve to tableau", "MoveReserveToTableau", []any{3, 0}, func(i *DuchessInteractor) string { return i.MoveReserveToTableau(3, 0) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *DuchessInteractor) string { return i.MoveWasteToFoundation() }},
		{"waste to tableau", "MoveWasteToTableau", []any{2}, func(i *DuchessInteractor) string { return i.MoveWasteToTableau(2) }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{1}, func(i *DuchessInteractor) string { return i.MoveTableauToFoundation(1) }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 2, 3}, func(i *DuchessInteractor) string { return i.MoveTableauToTableau(0, 2, 3) }},
		{"autocomplete", "AutoComplete", nil, func(i *DuchessInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *DuchessInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *DuchessInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockDuchessGame()
			p := newMockDuchessPresenter()
			i := NewDuchessInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockDuchessGame()
			p := newMockDuchessPresenter()
			i := NewDuchessInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestDuchessInteractorGiveUp(t *testing.T) {
	g := newMockDuchessGame()
	p := newMockDuchessPresenter()
	i := NewDuchessInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestDuchessInteractorHintAndActionLog(t *testing.T) {
	g := newMockDuchessGame()
	p := newMockDuchessPresenter()
	i := NewDuchessInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestDuchessInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultDuchess()
	d.Reset()
	i := NewDuchessInteractor(d, newMockDuchessPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreDuchessInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultDuchess()
		d.Reset()
		p := newMockDuchessPresenter()
		data, err := NewDuchessInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreDuchessInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreDuchessInteractor([]byte("not json"), newMockDuchessPresenter())
		assert.Error(t, err)
	})
}
