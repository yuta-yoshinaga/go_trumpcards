package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockJulepeInteractor() *mockusecase.MockJulepeInteractor {
	return new(mockusecase.MockJulepeInteractor)
}

func TestJulepeCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Play", []string{"in", "play"}},
		{"Pass", []string{"out", "pass"}},
		{"NextRound", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ri := newMockJulepeInteractor()
			c := NewJulepeCuiController(ri)
			ri.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// **参加と降りるを取り違えない。** ラウンドの行方が正反対になる。
func TestJulepeCuiControllerInAndOutAreDistinct(t *testing.T) {
	ri := newMockJulepeInteractor()
	c := NewJulepeCuiController(ri)
	ri.On("Play").Return("in")
	ri.On("Pass").Return("out")

	assert.Equal(t, "in", c.Exec("in"))
	assert.Equal(t, "out", c.Exec("out"))
	ri.AssertNumberOfCalls(t, "Play", 1)
	ri.AssertNumberOfCalls(t, "Pass", 1)
}

func TestJulepeCuiControllerResetKeepsConfig(t *testing.T) {
	ri := newMockJulepeInteractor()
	c := NewJulepeCuiController(ri)
	cfg := domain.JulepeConfig{PlayerCnt: 5, Rounds: 6}
	ri.On("GetConfig").Return(cfg)
	ri.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	ri.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestJulepeCuiControllerQuit(t *testing.T) {
	c := NewJulepeCuiController(newMockJulepeInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestJulepeCuiControllerCard(t *testing.T) {
	for _, alias := range []string{"c", "card"} {
		t.Run(alias, func(t *testing.T) {
			ri := newMockJulepeInteractor()
			c := NewJulepeCuiController(ri)
			ri.On("PlayCard", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			ri.AssertCalled(t, "PlayCard", 3)
		})
	}
}

func TestJulepeCuiControllerCardRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "c", msgCardIndexRequired()},
		{"non-numeric", "c abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ri := newMockJulepeInteractor()
			c := NewJulepeCuiController(ri)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ri.AssertNotCalled(t, "PlayCard", mock.Anything)
		})
	}
}

func TestJulepeCuiControllerUnknownCommand(t *testing.T) {
	ri := newMockJulepeInteractor()
	c := NewJulepeCuiController(ri)
	assert.Contains(t, c.Exec("nex"), "next")
	ri.AssertNotCalled(t, "NextRound")
	ri.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
