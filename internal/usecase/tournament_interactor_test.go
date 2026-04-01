package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestTournamentActions() (tournamentActions[interfaces.HoldemGame], *interfaces.MockHoldemGame, *presenter.MockHoldemPresenter) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	ta := newTournamentActions[interfaces.HoldemGame](mg, mp)
	return ta, mg, mp
}

func TestTournamentActions(t *testing.T) {
	tests := []struct {
		name       string
		gameMethod string
		callMethod func(tournamentActions[interfaces.HoldemGame]) string
	}{
		{"Rebuy", "Rebuy", func(ta tournamentActions[interfaces.HoldemGame]) string { return ta.Rebuy() }},
		{"SkipRebuy", "SkipRebuy", func(ta tournamentActions[interfaces.HoldemGame]) string { return ta.SkipRebuy() }},
		{"Addon", "Addon", func(ta tournamentActions[interfaces.HoldemGame]) string { return ta.Addon() }},
		{"SkipAddon", "SkipAddon", func(ta tournamentActions[interfaces.HoldemGame]) string { return ta.SkipAddon() }},
		{"Muck", "Muck", func(ta tournamentActions[interfaces.HoldemGame]) string { return ta.Muck() }},
		{"ShowHand", "ShowHand", func(ta tournamentActions[interfaces.HoldemGame]) string { return ta.ShowHand() }},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			ta, mg, mp := newTestTournamentActions()
			mg.On(tt.gameMethod).Return(nil)
			mp.On("Output", mg, mock.Anything).Return(tt.name + " output")

			result := tt.callMethod(ta)
			assert.Equal(t, tt.name+" output", result)
			mg.AssertCalled(t, tt.gameMethod)
		})
		t.Run(tt.name+"_error", func(t *testing.T) {
			ta, mg, mp := newTestTournamentActions()
			err := errors.New(tt.name + " failed")
			mg.On(tt.gameMethod).Return(err)
			mp.On("Output", mg, err).Return(tt.name + " error output")

			result := tt.callMethod(ta)
			assert.Equal(t, tt.name+" error output", result)
		})
	}
}
