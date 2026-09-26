//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestHorseTableRound(t *testing.T) {
	for _, tt := range []struct {
		discipline domain.HorseDiscipline
		phase      int
		want       string
	}{
		{domain.HorseHoldem, domain.HoldemPhaseFlop, "flop"},
		{domain.HorseRazz, domain.SevenCardStudPhaseFourthStreet, "fourth"},
		{domain.HorseTripleDraw, domain.DeuceToSevenPhaseDraw, "drawing"},
		{domain.HorseTripleDraw, domain.DeuceToSevenPhaseBet, "betting"},
		{domain.HorseTripleDraw, domain.DeuceToSevenPhaseShowdown, ""},
	} {
		assert.Equal(t, tt.want, horseTableRound(tt.discipline, tt.phase))
	}
}

func TestHorseActionLabelDistinguishesDrawTurn(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("ja") })
	i18n.SetLang("ja")

	g := new(interfaces.MockHorseGame)
	g.On("IsDrawPhase").Return(true)
	g.On("GetDrawIndex").Return(2)
	assert.Equal(t, "ドロー手番（2回目）", horseActionLabel(g))
	g.AssertExpectations(t)
}
