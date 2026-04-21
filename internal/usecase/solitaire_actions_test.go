//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestSolitaireActions() (solitaireActions[interfaces.KlondikeGame], *interfaces.MockKlondikeGame, *presenter.MockKlondikePresenter) {
	mg := new(interfaces.MockKlondikeGame)
	mp := new(presenter.MockKlondikePresenter)
	sa := newSolitaireActions[interfaces.KlondikeGame](mg, mp)
	return sa, mg, mp
}

func TestSolitaireActions(t *testing.T) {
	tests := []struct {
		name       string
		gameMethod string
		returnsErr bool
		callMethod func(solitaireActions[interfaces.KlondikeGame]) string
	}{
		{"GiveUp", "GiveUp", false, func(sa solitaireActions[interfaces.KlondikeGame]) string { return sa.GiveUp() }},
		{"AutoComplete", "AutoComplete", true, func(sa solitaireActions[interfaces.KlondikeGame]) string { return sa.AutoComplete() }},
		{"Undo", "Undo", true, func(sa solitaireActions[interfaces.KlondikeGame]) string { return sa.Undo() }},
		{"UndoN", "UndoN", true, func(sa solitaireActions[interfaces.KlondikeGame]) string { return sa.UndoN(3) }},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			sa, mg, mp := newTestSolitaireActions()
			if tt.returnsErr {
				if tt.gameMethod == "UndoN" {
					mg.On(tt.gameMethod, 3).Return(nil)
				} else {
					mg.On(tt.gameMethod).Return(nil)
				}
			} else {
				mg.On(tt.gameMethod).Return()
			}
			mp.On("Output", mg, mock.Anything).Return(tt.name + " output")

			result := tt.callMethod(sa)
			assert.Equal(t, tt.name+" output", result)
			mg.AssertExpectations(t)
		})

		if tt.returnsErr {
			t.Run(tt.name+"_error", func(t *testing.T) {
				sa, mg, mp := newTestSolitaireActions()
				boom := errors.New(tt.name + " failed")
				if tt.gameMethod == "UndoN" {
					mg.On(tt.gameMethod, 3).Return(boom)
				} else {
					mg.On(tt.gameMethod).Return(boom)
				}
				mp.On("Output", mg, boom).Return(tt.name + " error output")

				result := tt.callMethod(sa)
				assert.Equal(t, tt.name+" error output", result)
			})
		}
	}
}
