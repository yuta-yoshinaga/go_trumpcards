package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newBassetTestInteractor() (*BassetInteractor, *interfaces.MockBassetGame, *presenter.MockBassetPresenter) {
	g := new(interfaces.MockBassetGame)
	p := new(presenter.MockBassetPresenter)
	return NewBassetInteractor(g, p), g, p
}

func TestBassetInteractor_Commands(t *testing.T) {
	tests := []struct {
		name   string
		call   func(*BassetInteractor) string
		method string
	}{
		{"reset", func(i *BassetInteractor) string { return i.Reset() }, "Reset"},
		{"next", func(i *BassetInteractor) string { return i.NextRound() }, "NextRound"},
		{"bet", func(i *BassetInteractor) string { return i.PlaceBet(7, 10) }, "PlayerPlaceBet"},
		{"deal", func(i *BassetInteractor) string { return i.DealTurn() }, "PlayerDealTurn"},
		{"take", func(i *BassetInteractor) string { return i.TakeWinnings() }, "PlayerTakeWinnings"},
		{"paroli", func(i *BassetInteractor) string { return i.PressParoli() }, "PlayerPressParoli"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i, g, p := newBassetTestInteractor()
			g.On(tc.method, mock.Anything, mock.Anything).Maybe().Return(nil)
			if tc.method == "Reset" || tc.method == "NextRound" || tc.method == "PlayerDealTurn" || tc.method == "PlayerTakeWinnings" || tc.method == "PlayerPressParoli" {
				g.On(tc.method).Return(nil)
			}
			p.On("Output", g, nil).Return(tc.name)
			assert.Equal(t, tc.name, tc.call(i))
		})
	}
}

func TestBassetInteractor_PlaceBetError(t *testing.T) {
	i, g, p := newBassetTestInteractor()
	g.On("PlayerPlaceBet", 7, 10).Return(errors.New("bad"))
	p.On("Output", g, mock.MatchedBy(func(err error) bool { return err != nil })).Return("error")
	assert.Equal(t, "error", i.PlaceBet(7, 10))
}
func TestBassetInteractor_ActionLog(t *testing.T) {
	i, g, p := newBassetTestInteractor()
	p.On("ActionLogOutput", g).Return("log")
	assert.Equal(t, "log", i.ActionLog())
}
func TestRestoreBassetInteractor(t *testing.T) {
	_, err := RestoreBassetInteractor([]byte("not json"), nil)
	assert.Error(t, err)
}
