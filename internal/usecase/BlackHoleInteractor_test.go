//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newBhInteractor() (*usecase.BlackHoleInteractor, *domain.BlackHole) {
	g := domain.NewDefaultBlackHole()
	g.Reset()
	li := usecase.NewBlackHoleInteractor(g, new(presenter.BlackHoleWebPresenter))
	return li, g
}

func TestNewBlackHoleInteractor_NilGuards(t *testing.T) {
	op := new(presenter.BlackHoleWebPresenter)
	assert.PanicsWithValue(t, "BlackHoleInteractor: g must not be nil", func() {
		usecase.NewBlackHoleInteractor(nil, op)
	})
	g := domain.NewDefaultBlackHole()
	assert.PanicsWithValue(t, "BlackHoleInteractor: op must not be nil", func() {
		usecase.NewBlackHoleInteractor(g, nil)
	})
}

func TestBlackHoleInteractor_Commands(t *testing.T) {
	li, g := newBhInteractor()
	assert.Contains(t, li.Reset(), `"phase"`)
	assert.Contains(t, li.Hint(), `"phase"`)
	assert.NotEmpty(t, li.ActionLog())

	// Apply one hint-suggested legal move through the interactor.
	if h := g.GetHint(); h != nil {
		assert.Contains(t, li.MoveFanToBlackHole(h.Fan), `"phase"`)
		assert.Contains(t, li.Undo(), `"phase"`)
	}
	assert.NotEmpty(t, li.UndoN(2))
	// Illegal move returns error output.
	assert.NotEmpty(t, li.MoveFanToBlackHole(999))
}

func TestBlackHoleInteractor_Snapshot(t *testing.T) {
	li, _ := newBhInteractor()
	data, err := li.Snapshot()
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)
	restored, err := usecase.RestoreBlackHoleInteractor(data, new(presenter.BlackHoleWebPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	_, err = usecase.RestoreBlackHoleInteractor([]byte("bad"), new(presenter.BlackHoleWebPresenter))
	assert.Error(t, err)
}

func TestBlackHoleInteractor_GiveUpThenBlocked(t *testing.T) {
	li, g := newBhInteractor()
	assert.Contains(t, li.GiveUp(), `"phase"`)
	assert.True(t, g.GetGameEndFlag())
	// Actions after game end still produce output via the presenter.
	assert.NotEmpty(t, li.MoveFanToBlackHole(0))
}
