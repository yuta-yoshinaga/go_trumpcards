package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockSetteEMezzoInteractor() *mockusecase.MockSetteEMezzoInteractor {
	return new(mockusecase.MockSetteEMezzoInteractor)
}

func TestSetteEMezzoCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"Deal", []string{"deal"}},
		{"Hit", []string{"h", "hit"}},
		{"Stand", []string{"s", "stand"}},
		{"BankerHit", []string{"bh"}},
		{"BankerStand", []string{"bs"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			si := newMockSetteEMezzoInteractor()
			c := NewSetteEMezzoCuiController(si)
			si.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestSetteEMezzoCuiControllerQuit(t *testing.T) {
	c := NewSetteEMezzoCuiController(newMockSetteEMezzoInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSetteEMezzoCuiControllerBet(t *testing.T) {
	si := newMockSetteEMezzoInteractor()
	c := NewSetteEMezzoCuiController(si)
	si.On("Bet", 100).Return("bet")

	assert.Equal(t, "bet", c.Exec("b 100"))
	assert.Equal(t, "bet", c.Exec("bet 100"))
}

// The player types POINTS; the interactor takes HALVES. Getting this conversion
// wrong would silently halve or double every matta.
func TestSetteEMezzoCuiControllerMattaConvertsPointsToHalves(t *testing.T) {
	for _, tc := range []struct {
		input  string
		halves int
	}{
		{"matta 0.5", 1},
		{"matta 1", 2},
		{"matta 3", 6},
		{"matta 7", 14},
	} {
		t.Run(tc.input, func(t *testing.T) {
			si := newMockSetteEMezzoInteractor()
			c := NewSetteEMezzoCuiController(si)
			si.On("Matta", tc.halves).Return("matta")
			assert.Equal(t, "matta", c.Exec(tc.input))
			si.AssertCalled(t, "Matta", tc.halves)
		})
	}
}

func TestSetteEMezzoCuiControllerRejectsBadInput(t *testing.T) {
	for _, tc := range []struct{ cmd, contains string }{
		{"b", msgBetAmountRequired()},
		{"b abc", msgInvalidBetAmountPrefix()},
		{"matta", msgStem("mattaValueRequired05Or17")},
		{"matta abc", msgStem("invalidMattaValueEnter05OrAWholeNumberFrom1To7")},
		// 0.5 と 1〜7 以外は取れない。
		{"matta 8", msgStem("invalidMattaValueEnter05OrAWholeNumberFrom1To7")},
		{"matta 0", msgStem("invalidMattaValueEnter05OrAWholeNumberFrom1To7")},
	} {
		t.Run(tc.cmd, func(t *testing.T) {
			c := NewSetteEMezzoCuiController(newMockSetteEMezzoInteractor())
			assert.Contains(t, c.Exec(tc.cmd), tc.contains)
		})
	}
}

func TestSetteEMezzoCuiControllerUnknownCommand(t *testing.T) {
	c := NewSetteEMezzoCuiController(newMockSetteEMezzoInteractor())
	assert.Contains(t, c.Exec("zzzzz"), "zzzzz")
}
