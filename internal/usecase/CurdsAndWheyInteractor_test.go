//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newSsInteractorCW() (*usecase.CurdsAndWheyInteractor, *domain.CurdsAndWhey) {
	g := domain.NewDefaultCurdsAndWhey()
	g.Reset()
	si := usecase.NewCurdsAndWheyInteractor(g, new(presenter.CurdsAndWheyWebPresenter))
	return si, g
}

func TestNewCurdsAndWheyInteractor_NilGuards(t *testing.T) {
	op := new(presenter.CurdsAndWheyWebPresenter)
	assert.PanicsWithValue(t, "CurdsAndWheyInteractor: g must not be nil", func() {
		usecase.NewCurdsAndWheyInteractor(nil, op)
	})
	g := domain.NewDefaultCurdsAndWhey()
	assert.PanicsWithValue(t, "CurdsAndWheyInteractor: op must not be nil", func() {
		usecase.NewCurdsAndWheyInteractor(g, nil)
	})
}

func TestCurdsAndWheyInteractor_Commands(t *testing.T) {
	si, g := newSsInteractorCW()
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

func TestCurdsAndWheyInteractor_Snapshot(t *testing.T) {
	si, _ := newSsInteractorCW()
	data, err := si.Snapshot()
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)
	restored, err := usecase.RestoreCurdsAndWheyInteractor(data, new(presenter.CurdsAndWheyWebPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	_, err = usecase.RestoreCurdsAndWheyInteractor([]byte("bad"), new(presenter.CurdsAndWheyWebPresenter))
	assert.Error(t, err)
}

func TestCurdsAndWheyInteractor_GiveUpThenBlocked(t *testing.T) {
	si, g := newSsInteractorCW()
	assert.Contains(t, si.GiveUp(), `"phase"`)
	assert.True(t, g.GetGameEndFlag())
	assert.NotEmpty(t, si.MoveSequence(0, 0, 1))
}
