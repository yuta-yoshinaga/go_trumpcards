//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newLlInteractorSH() (*usecase.ShamrocksInteractor, *domain.Shamrocks) {
	g := domain.NewDefaultShamrocks()
	g.Reset()
	li := usecase.NewShamrocksInteractor(g, new(presenter.ShamrocksWebPresenter))
	return li, g
}

func TestNewShamrocksInteractor_NilGuards(t *testing.T) {
	op := new(presenter.ShamrocksWebPresenter)
	assert.PanicsWithValue(t, "ShamrocksInteractor: g must not be nil", func() {
		usecase.NewShamrocksInteractor(nil, op)
	})
	g := domain.NewDefaultShamrocks()
	assert.PanicsWithValue(t, "ShamrocksInteractor: op must not be nil", func() {
		usecase.NewShamrocksInteractor(g, nil)
	})
}

func TestShamrocksInteractor_Commands(t *testing.T) {
	li, g := newLlInteractorSH()
	assert.Contains(t, li.Reset(), `"phase"`)
	assert.Contains(t, li.Hint(), `"phase"`)
	assert.NotEmpty(t, li.ActionLog())
	assert.Contains(t, li.AutoComplete(), `"phase"`)

	// Apply one hint-suggested legal move through the interactor.
	if h := g.GetHint(); h != nil {
		if h.ToFoundation {
			assert.Contains(t, li.MoveFanToFoundation(h.FromFan), `"phase"`)
		} else {
			assert.Contains(t, li.MoveFanToFan(h.FromFan, h.ToFan), `"phase"`)
		}
		assert.Contains(t, li.Undo(), `"phase"`)
	}
	assert.NotEmpty(t, li.UndoN(2))
	// Illegal move returns error output.
	assert.NotEmpty(t, li.MoveFanToFoundation(999))
}

func TestShamrocksInteractor_Snapshot(t *testing.T) {
	li, _ := newLlInteractorSH()
	data, err := li.Snapshot()
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)
	restored, err := usecase.RestoreShamrocksInteractor(data, new(presenter.ShamrocksWebPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	_, err = usecase.RestoreShamrocksInteractor([]byte("bad"), new(presenter.ShamrocksWebPresenter))
	assert.Error(t, err)
}

func TestShamrocksInteractor_GiveUpThenBlocked(t *testing.T) {
	li, g := newLlInteractorSH()
	assert.Contains(t, li.GiveUp(), `"phase"`)
	assert.True(t, g.GetGameEndFlag())
	// Actions after game end still produce output via the presenter.
	assert.NotEmpty(t, li.MoveFanToFan(0, 1))
}
