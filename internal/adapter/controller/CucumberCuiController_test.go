//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newCucumberCui() (*controller.CucumberCuiController, *usecase.MockCucumberInteractor) {
	ci := new(usecase.MockCucumberInteractor)
	return controller.NewCucumberCuiController(ci), ci
}

func TestCucumberCuiController_Commands(t *testing.T) {
	c, ci := newCucumberCui()
	cfg := domain.DefaultCucumberConfig()
	ci.On("GetConfig").Return(cfg)
	ci.On("ResetWithConfig", cfg).Return("reset")
	ci.On("Play", 2).Return("play")
	ci.On("NextRound").Return("next")
	ci.On("GiveUp").Return("giveup")
	ci.On("Hint").Return("hint")
	ci.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"play", "p 2", "play"},
		{"play long", "play 2", "play"},
		{"next", "n", "next"},
		{"next long", "next", "next"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

func TestCucumberCuiController_PlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "p", msgCardIndexRequired()},
		{"index not a number", "p x", msgInvalidCardIndex("x")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, ci := newCucumberCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ci.AssertNotCalled(t, "Play")
		})
	}
}

func TestCucumberCuiController_UnknownCommand(t *testing.T) {
	c, _ := newCucumberCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}
