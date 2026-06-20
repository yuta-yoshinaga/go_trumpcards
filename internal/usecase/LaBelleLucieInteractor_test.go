//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newLlInteractor() (*usecase.LaBelleLucieInteractor, *domain.LaBelleLucie) {
	g := domain.NewDefaultLaBelleLucie()
	g.Reset()
	li := usecase.NewLaBelleLucieInteractor(g, new(presenter.LaBelleLucieWebPresenter))
	return li, g
}

func TestNewLaBelleLucieInteractor_NilGuards(t *testing.T) {
	op := new(presenter.LaBelleLucieWebPresenter)
	assert.PanicsWithValue(t, "LaBelleLucieInteractor: g must not be nil", func() {
		usecase.NewLaBelleLucieInteractor(nil, op)
	})
	g := domain.NewDefaultLaBelleLucie()
	assert.PanicsWithValue(t, "LaBelleLucieInteractor: op must not be nil", func() {
		usecase.NewLaBelleLucieInteractor(g, nil)
	})
}

func TestLaBelleLucieInteractor_Commands(t *testing.T) {
	li, g := newLlInteractor()
	assert.Contains(t, li.Reset(), `"phase"`)
	assert.Contains(t, li.Hint(), `"phase"`)
	assert.NotEmpty(t, li.ActionLog())
	assert.Contains(t, li.Redeal(), `"redealsLeft"`)
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

func TestLaBelleLucieInteractor_Snapshot(t *testing.T) {
	li, _ := newLlInteractor()
	data, err := li.Snapshot()
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)
	restored, err := usecase.RestoreLaBelleLucieInteractor(data, new(presenter.LaBelleLucieWebPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	_, err = usecase.RestoreLaBelleLucieInteractor([]byte("bad"), new(presenter.LaBelleLucieWebPresenter))
	assert.Error(t, err)
}

func TestLaBelleLucieInteractor_GiveUpThenBlocked(t *testing.T) {
	li, g := newLlInteractor()
	assert.Contains(t, li.GiveUp(), `"phase"`)
	assert.True(t, g.GetGameEndFlag())
	// Actions after game end still produce output via the presenter.
	assert.NotEmpty(t, li.Redeal())
	assert.NotEmpty(t, li.MoveFanToFan(0, 1))
}
