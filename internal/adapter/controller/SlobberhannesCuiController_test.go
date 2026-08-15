package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockSlobberhannesInteractor() *mockusecase.MockSlobberhannesInteractor {
	return new(mockusecase.MockSlobberhannesInteractor)
}

func TestSlobberhannesCuiControllerSimpleCommands(t *testing.T) {
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
			si := newMockSlobberhannesInteractor()
			c := NewSlobberhannesCuiController(si)
			si.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// reset は現在の設定を保ったまま配り直す。設定が飛ぶと途中でラウンド数が変わる。
func TestSlobberhannesCuiControllerResetKeepsConfig(t *testing.T) {
	si := newMockSlobberhannesInteractor()
	c := NewSlobberhannesCuiController(si)
	cfg := domain.SlobberhannesConfig{Rounds: 6}
	si.On("GetConfig").Return(cfg)
	si.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	si.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestSlobberhannesCuiControllerQuit(t *testing.T) {
	c := NewSlobberhannesCuiController(newMockSlobberhannesInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSlobberhannesCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			si := newMockSlobberhannesInteractor()
			c := NewSlobberhannesCuiController(si)
			si.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			si.AssertCalled(t, "Play", 3)
		})
	}
}

func TestSlobberhannesCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			si := newMockSlobberhannesInteractor()
			c := NewSlobberhannesCuiController(si)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

// 綴り間違いで盤面が消えないこと。
func TestSlobberhannesCuiControllerUnknownCommand(t *testing.T) {
	si := newMockSlobberhannesInteractor()
	c := NewSlobberhannesCuiController(si)
	assert.Contains(t, c.Exec("pla 1"), "play")
	si.AssertNotCalled(t, "Play", mock.Anything)
	si.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
