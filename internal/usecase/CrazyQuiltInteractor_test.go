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

func newMockCrazyQuiltGame() *interfaces.MockCrazyQuiltGame {
	return new(interfaces.MockCrazyQuiltGame)
}

func newMockCrazyQuiltPresenter() *presenter.MockCrazyQuiltPresenter {
	return new(presenter.MockCrazyQuiltPresenter)
}

func TestNewCrazyQuiltInteractor(t *testing.T) {
	assert.NotNil(t, NewCrazyQuiltInteractor(newMockCrazyQuiltGame(), newMockCrazyQuiltPresenter()))
}

func TestNewCrazyQuiltInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewCrazyQuiltInteractor(nil, newMockCrazyQuiltPresenter()) })
	assert.Panics(t, func() { NewCrazyQuiltInteractor(newMockCrazyQuiltGame(), nil) })
}

func TestCrazyQuiltInteractorReset(t *testing.T) {
	g := newMockCrazyQuiltGame()
	p := newMockCrazyQuiltPresenter()
	i := NewCrazyQuiltInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestCrazyQuiltInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*CrazyQuiltInteractor) string
	}{
		{"draw", "Draw", nil, func(i *CrazyQuiltInteractor) string { return i.Draw() }},
		{"quilt to foundation", "MoveQuiltToFoundation", []any{3}, func(i *CrazyQuiltInteractor) string { return i.MoveQuiltToFoundation(3) }},
		{"quilt to waste", "MoveQuiltToWaste", []any{5}, func(i *CrazyQuiltInteractor) string { return i.MoveQuiltToWaste(5) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *CrazyQuiltInteractor) string { return i.MoveWasteToFoundation() }},
		{"autocomplete", "AutoComplete", nil, func(i *CrazyQuiltInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *CrazyQuiltInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *CrazyQuiltInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockCrazyQuiltGame()
			p := newMockCrazyQuiltPresenter()
			i := NewCrazyQuiltInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockCrazyQuiltGame()
			p := newMockCrazyQuiltPresenter()
			i := NewCrazyQuiltInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestCrazyQuiltInteractorGiveUp(t *testing.T) {
	g := newMockCrazyQuiltGame()
	p := newMockCrazyQuiltPresenter()
	i := NewCrazyQuiltInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestCrazyQuiltInteractorHintAndActionLog(t *testing.T) {
	g := newMockCrazyQuiltGame()
	p := newMockCrazyQuiltPresenter()
	i := NewCrazyQuiltInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestCrazyQuiltInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultCrazyQuilt()
	d.Reset()
	i := NewCrazyQuiltInteractor(d, newMockCrazyQuiltPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreCrazyQuiltInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultCrazyQuilt()
		d.Reset()
		p := newMockCrazyQuiltPresenter()
		data, err := NewCrazyQuiltInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreCrazyQuiltInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreCrazyQuiltInteractor([]byte("not json"), newMockCrazyQuiltPresenter())
		assert.Error(t, err)
	})
}
