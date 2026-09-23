//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

type mockMatrimonyGame struct{ mock.Mock }

func (g *mockMatrimonyGame) Reset()                                 { g.Called() }
func (g *mockMatrimonyGame) Draw() error                            { return g.Called().Error(0) }
func (g *mockMatrimonyGame) MoveTableauToFoundation(pile int) error { return g.Called(pile).Error(0) }
func (g *mockMatrimonyGame) MoveWasteToFoundation() error           { return g.Called().Error(0) }
func (g *mockMatrimonyGame) MoveWasteToTableau(pile int) error      { return g.Called(pile).Error(0) }
func (g *mockMatrimonyGame) MoveStockToTableau(pile int) error      { return g.Called(pile).Error(0) }
func (g *mockMatrimonyGame) GiveUp()                                { g.Called() }
func (g *mockMatrimonyGame) AutoComplete() error                    { return g.Called().Error(0) }
func (g *mockMatrimonyGame) Undo() error                            { return g.Called().Error(0) }
func (g *mockMatrimonyGame) UndoN(n int) error                      { return g.Called(n).Error(0) }
func (g *mockMatrimonyGame) CanUndo() bool                          { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) UndoToEscape() int                      { return g.Called().Int(0) }
func (g *mockMatrimonyGame) AllFaceUp() bool                        { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) GetGameEndFlag() bool                   { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) GetHint() *domain.MatrimonyHint {
	v := g.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.MatrimonyHint)
}
func (g *mockMatrimonyGame) GetPhase() domain.MatrimonyPhase {
	return g.Called().Get(0).(domain.MatrimonyPhase)
}
func (g *mockMatrimonyGame) GetMoveCount() int   { return g.Called().Int(0) }
func (g *mockMatrimonyGame) GetStockCount() int  { return g.Called().Int(0) }
func (g *mockMatrimonyGame) GetRedealCount() int { return g.Called().Int(0) }
func (g *mockMatrimonyGame) GetWaste() []*domain.Card {
	v := g.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}
func (g *mockMatrimonyGame) GetTableau() [domain.MatrimonyTableauCnt]*domain.Card {
	return g.Called().Get(0).([domain.MatrimonyTableauCnt]*domain.Card)
}
func (g *mockMatrimonyGame) GetFoundation() [domain.MatrimonyFoundationCnt][]*domain.Card {
	return g.Called().Get(0).([domain.MatrimonyFoundationCnt][]*domain.Card)
}
func (g *mockMatrimonyGame) IsStalemate() bool { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) GetActionLog() []*domain.ActionLogEntry {
	v := g.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func newMockMatrimonyGame() *mockMatrimonyGame {
	return new(mockMatrimonyGame)
}

func newMockMatrimonyPresenter() *presenter.MockMatrimonyPresenter {
	return new(presenter.MockMatrimonyPresenter)
}

func TestNewMatrimonyInteractor(t *testing.T) {
	assert.NotNil(t, NewMatrimonyInteractor(newMockMatrimonyGame(), newMockMatrimonyPresenter()))
}

func TestNewMatrimonyInteractorPanicsOnNil(t *testing.T) {
	assert.Panics(t, func() { NewMatrimonyInteractor(nil, newMockMatrimonyPresenter()) })
	assert.Panics(t, func() { NewMatrimonyInteractor(newMockMatrimonyGame(), nil) })
}

func TestMatrimonyInteractorReset(t *testing.T) {
	g := newMockMatrimonyGame()
	p := newMockMatrimonyPresenter()
	i := NewMatrimonyInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", i.Reset())
	g.AssertCalled(t, "Reset")
}

// Every action forwards its arguments and routes the error to the presenter.
func TestMatrimonyInteractorActions(t *testing.T) {
	cases := []struct {
		name   string
		method string
		args   []any
		call   func(*MatrimonyInteractor) string
	}{
		{"draw", "Draw", nil, func(i *MatrimonyInteractor) string { return i.Draw() }},
		{"tableau to foundation", "MoveTableauToFoundation", []any{3}, func(i *MatrimonyInteractor) string { return i.MoveTableauToFoundation(3) }},
		{"waste to foundation", "MoveWasteToFoundation", nil, func(i *MatrimonyInteractor) string { return i.MoveWasteToFoundation() }},
		{"waste to tableau", "MoveWasteToTableau", []any{2}, func(i *MatrimonyInteractor) string { return i.MoveWasteToTableau(2) }},
		{"stock to tableau", "MoveStockToTableau", []any{4}, func(i *MatrimonyInteractor) string { return i.MoveStockToTableau(4) }},
		{"autocomplete", "AutoComplete", nil, func(i *MatrimonyInteractor) string { return i.AutoComplete() }},
		{"undo", "Undo", nil, func(i *MatrimonyInteractor) string { return i.Undo() }},
		{"undo n", "UndoN", []any{3}, func(i *MatrimonyInteractor) string { return i.UndoN(3) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+" success", func(t *testing.T) {
			g := newMockMatrimonyGame()
			p := newMockMatrimonyPresenter()
			i := NewMatrimonyInteractor(g, p)
			g.On(tc.method, tc.args...).Return(nil)
			p.On("Output", g, nil).Return("ok_output")
			assert.Equal(t, "ok_output", tc.call(i))
			g.AssertCalled(t, tc.method, tc.args...)
		})
		t.Run(tc.name+" error", func(t *testing.T) {
			g := newMockMatrimonyGame()
			p := newMockMatrimonyPresenter()
			i := NewMatrimonyInteractor(g, p)
			err := errors.New("invalid")
			g.On(tc.method, tc.args...).Return(err)
			p.On("Output", g, err).Return("error_output")
			assert.Equal(t, "error_output", tc.call(i))
		})
	}
}

func TestMatrimonyInteractorGiveUp(t *testing.T) {
	g := newMockMatrimonyGame()
	p := newMockMatrimonyPresenter()
	i := NewMatrimonyInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", i.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestMatrimonyInteractorHintAndActionLog(t *testing.T) {
	g := newMockMatrimonyGame()
	p := newMockMatrimonyPresenter()
	i := NewMatrimonyInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")
	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "hint_output", i.Hint())
	assert.Equal(t, "log_output", i.ActionLog())
}

func TestMatrimonyInteractorSnapshot(t *testing.T) {
	d := domain.NewDefaultMatrimony()
	d.Reset()
	i := NewMatrimonyInteractor(d, newMockMatrimonyPresenter())

	data, err := i.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreMatrimonyInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		d := domain.NewDefaultMatrimony()
		d.Reset()
		p := newMockMatrimonyPresenter()
		data, err := NewMatrimonyInteractor(d, p).Snapshot()
		require.NoError(t, err)

		restored, err := RestoreMatrimonyInteractor(data, p)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := RestoreMatrimonyInteractor([]byte("not json"), newMockMatrimonyPresenter())
		assert.Error(t, err)
	})
}
