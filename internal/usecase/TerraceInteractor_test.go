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

func newMockTerraceGame() *interfaces.MockTerraceGame {
	return new(interfaces.MockTerraceGame)
}

func newMockTerracePresenter() *presenter.MockTerracePresenter {
	return new(presenter.MockTerracePresenter)
}

func TestNewTerraceInteractor(t *testing.T) {
	assert.NotNil(t, NewTerraceInteractor(newMockTerraceGame(), newMockTerracePresenter()))
}

func TestNewTerraceInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewTerraceInteractor(nil, newMockTerracePresenter()) })
	assert.Panics(t, func() { NewTerraceInteractor(newMockTerraceGame(), nil) })
}

func TestTerraceInteractorReset(t *testing.T) {
	g := newMockTerraceGame()
	p := newMockTerracePresenter()
	i := NewTerraceInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestTerraceInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*TerraceInteractor) string
	}{
		{"draw", "Draw", nil, func(i *TerraceInteractor) string { return i.Draw() }},
		{"reserve to foundation", "MoveReserveToFoundation", nil, func(i *TerraceInteractor) string { return i.MoveReserveToFoundation() }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *TerraceInteractor) string { return i.MoveWasteToFoundation() }},
		{"waste to tableau", "MoveWasteToTableau", []any{2}, func(i *TerraceInteractor) string { return i.MoveWasteToTableau(2) }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{3}, func(i *TerraceInteractor) string { return i.MoveTableauToFoundation(3) }},
		{"tableau to tableau", "MoveTableauToTableau", []any{0, 5}, func(i *TerraceInteractor) string { return i.MoveTableauToTableau(0, 5) }},
		{"autocomplete", "AutoComplete", nil, func(i *TerraceInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *TerraceInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *TerraceInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockTerraceGame()
			p := newMockTerracePresenter()
			i := NewTerraceInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockTerraceGame()
			p := newMockTerracePresenter()
			i := NewTerraceInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestTerraceInteractorGiveUp(t *testing.T) {
	g := newMockTerraceGame()
	p := newMockTerracePresenter()
	i := NewTerraceInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestTerraceInteractorHintAndActionLog(t *testing.T) {
	g := newMockTerraceGame()
	p := newMockTerracePresenter()
	i := NewTerraceInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestTerraceInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultTerrace()
	d.Reset()
	i := NewTerraceInteractor(d, newMockTerracePresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreTerraceInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultTerrace()
		d.Reset()
		p := newMockTerracePresenter()
		data, err := NewTerraceInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreTerraceInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreTerraceInteractor([]byte("not json"), newMockTerracePresenter())
		assert.Error(t, err)
	})
}
