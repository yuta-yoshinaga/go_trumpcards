//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newDkInteractor() (*usecase.DoubleKlondikeInteractor, *domain.DoubleKlondike) {
	g := domain.NewDefaultDoubleKlondike()
	g.Reset()
	di := usecase.NewDoubleKlondikeInteractor(g, new(presenter.DoubleKlondikeWebPresenter))
	return di, g
}

func TestNewDoubleKlondikeInteractor_NilGuards(t *testing.T) {
	op := new(presenter.DoubleKlondikeWebPresenter)
	assert.PanicsWithValue(t, "DoubleKlondikeInteractor: g must not be nil", func() {
		usecase.NewDoubleKlondikeInteractor(nil, op)
	})
	g := domain.NewDefaultDoubleKlondike()
	assert.PanicsWithValue(t, "DoubleKlondikeInteractor: op must not be nil", func() {
		usecase.NewDoubleKlondikeInteractor(g, nil)
	})
}

func TestDoubleKlondikeInteractor_Commands(t *testing.T) {
	di, g := newDkInteractor()
	assert.Contains(t, di.Reset(), `"phase"`)
	assert.Contains(t, di.Draw(), `"phase"`)
	assert.Contains(t, di.Hint(), `"phase"`)
	assert.NotEmpty(t, di.ActionLog())

	// Apply one hint-suggested legal move through the interactor.
	if h := g.GetHint(); h != nil {
		switch {
		case h.ToZone == "foundation" && h.FromZone == "waste":
			assert.Contains(t, di.MoveWasteToFoundation(), `"phase"`)
		case h.ToZone == "foundation":
			assert.Contains(t, di.MoveTableauToFoundation(h.FromCol), `"phase"`)
		case h.FromZone == "waste":
			assert.Contains(t, di.MoveWasteToTableau(h.ToCol), `"phase"`)
		default:
			assert.Contains(t, di.MoveTableauToTableau(h.FromCol, h.CardIndex, h.ToCol), `"phase"`)
		}
		assert.Contains(t, di.Undo(), `"phase"`)
	}
	assert.NotEmpty(t, di.UndoN(2))
	// Illegal moves return error output.
	assert.NotEmpty(t, di.MoveWasteToTableau(99))
	assert.NotEmpty(t, di.MoveTableauToFoundation(99))
	assert.NotEmpty(t, di.MoveTableauToTableau(0, 0, 99))
}

func TestDoubleKlondikeInteractor_Snapshot(t *testing.T) {
	di, _ := newDkInteractor()
	data, err := di.Snapshot()
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)
	restored, err := usecase.RestoreDoubleKlondikeInteractor(data, new(presenter.DoubleKlondikeWebPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	_, err = usecase.RestoreDoubleKlondikeInteractor([]byte("bad"), new(presenter.DoubleKlondikeWebPresenter))
	assert.Error(t, err)
}

func TestDoubleKlondikeInteractor_GiveUpAutoComplete(t *testing.T) {
	di, g := newDkInteractor()
	assert.NotEmpty(t, di.AutoComplete()) // not all face up after a deal -> error output
	assert.Contains(t, di.GiveUp(), `"phase"`)
	assert.True(t, g.GetGameEndFlag())
	assert.NotEmpty(t, di.Draw())
}
