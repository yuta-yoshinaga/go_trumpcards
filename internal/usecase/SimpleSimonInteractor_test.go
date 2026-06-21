//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newSsInteractor() (*usecase.SimpleSimonInteractor, *domain.SimpleSimon) {
	g := domain.NewDefaultSimpleSimon()
	g.Reset()
	si := usecase.NewSimpleSimonInteractor(g, new(presenter.SimpleSimonWebPresenter))
	return si, g
}

func TestNewSimpleSimonInteractor_NilGuards(t *testing.T) {
	op := new(presenter.SimpleSimonWebPresenter)
	assert.PanicsWithValue(t, "SimpleSimonInteractor: g must not be nil", func() {
		usecase.NewSimpleSimonInteractor(nil, op)
	})
	g := domain.NewDefaultSimpleSimon()
	assert.PanicsWithValue(t, "SimpleSimonInteractor: op must not be nil", func() {
		usecase.NewSimpleSimonInteractor(g, nil)
	})
}

func TestSimpleSimonInteractor_Commands(t *testing.T) {
	si, g := newSsInteractor()
	assert.Contains(t, si.Reset(), `"phase"`)
	assert.Contains(t, si.Hint(), `"phase"`)
	assert.NotEmpty(t, si.ActionLog())

	// Apply one hint-suggested legal move.
	if h := g.GetHint(); h != nil {
		assert.Contains(t, si.MoveSequence(h.FromCol, h.CardIndex, h.ToCol), `"phase"`)
		assert.Contains(t, si.Undo(), `"phase"`)
	}
	assert.NotEmpty(t, si.UndoN(2))
	// Illegal move returns error output.
	assert.NotEmpty(t, si.MoveSequence(0, 0, 99))
}

func TestSimpleSimonInteractor_Snapshot(t *testing.T) {
	si, _ := newSsInteractor()
	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)
	restored, err := usecase.RestoreSimpleSimonInteractor(data, new(presenter.SimpleSimonWebPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	_, err = usecase.RestoreSimpleSimonInteractor([]byte("bad"), new(presenter.SimpleSimonWebPresenter))
	assert.Error(t, err)
}

func TestSimpleSimonInteractor_GiveUpThenBlocked(t *testing.T) {
	si, g := newSsInteractor()
	assert.Contains(t, si.GiveUp(), `"phase"`)
	assert.True(t, g.GetGameEndFlag())
	assert.NotEmpty(t, si.MoveSequence(0, 0, 1))
}
