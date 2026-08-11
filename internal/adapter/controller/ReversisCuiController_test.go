package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockReversisInteractor() *mockusecase.MockReversisInteractor {
	return new(mockusecase.MockReversisInteractor)
}

func TestReversisCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"NextRound", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ri := newMockReversisInteractor()
			c := NewReversisCuiController(ri)
			ri.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestReversisCuiControllerResetKeepsConfig(t *testing.T) {
	ri := newMockReversisInteractor()
	c := NewReversisCuiController(ri)
	cfg := domain.ReversisConfig{Rounds: 6}
	ri.On("GetConfig").Return(cfg)
	ri.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	ri.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestReversisCuiControllerQuit(t *testing.T) {
	c := NewReversisCuiController(newMockReversisInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestReversisCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			ri := newMockReversisInteractor()
			c := NewReversisCuiController(ri)
			ri.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			ri.AssertCalled(t, "Play", 3)
		})
	}
}

func TestReversisCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", "Card index is required."},
		{"non-numeric", "p abc", "Invalid card index: abc."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ri := newMockReversisInteractor()
			c := NewReversisCuiController(ri)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ri.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestReversisCuiControllerUnknownCommand(t *testing.T) {
	ri := newMockReversisInteractor()
	c := NewReversisCuiController(ri)
	assert.Contains(t, c.Exec("nex"), "next")
	ri.AssertNotCalled(t, "NextRound")
	ri.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
