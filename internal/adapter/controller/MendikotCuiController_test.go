package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockMendikotInteractor() *mockusecase.MockMendikotInteractor {
	return new(mockusecase.MockMendikotInteractor)
}

func TestMendikotCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"NextHand", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			mi := newMockMendikotInteractor()
			c := NewMendikotCuiController(mi)
			mi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestMendikotCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			mi := newMockMendikotInteractor()
			c := NewMendikotCuiController(mi)
			mi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			mi.AssertCalled(t, "Play", 3)
		})
	}
}

func TestMendikotCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", "Card index is required."},
		{"non-numeric", "p abc", "Invalid card index: abc."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mi := newMockMendikotInteractor()
			c := NewMendikotCuiController(mi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			mi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

// **切り札を選ぶコマンドは存在しない。** フォローできずに出した札のスートが
// そのまま切り札になるので、`t`/`trump` を受け付けたらルールが二重になる。
func TestMendikotCuiControllerHasNoTrumpCommand(t *testing.T) {
	for _, cmd := range []string{"t 3", "trump 3", "t", "trump"} {
		mi := newMockMendikotInteractor()
		c := NewMendikotCuiController(mi)
		out := c.Exec(cmd)
		assert.NotEmpty(t, out, "未知コマンドとして案内を返す: "+cmd)
		mi.AssertNotCalled(t, "Play", mock.Anything)
		mi.AssertNotCalled(t, "NextHand")
	}
}

func TestMendikotCuiControllerResetKeepsConfig(t *testing.T) {
	mi := newMockMendikotInteractor()
	c := NewMendikotCuiController(mi)
	cfg := domain.MendikotConfig{Target: 7}
	mi.On("GetConfig").Return(cfg)
	mi.On("ResetWithConfig", cfg).Return("reset")
	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	mi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestMendikotCuiControllerQuit(t *testing.T) {
	c := NewMendikotCuiController(newMockMendikotInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestMendikotCuiControllerUnknownCommand(t *testing.T) {
	mi := newMockMendikotInteractor()
	c := NewMendikotCuiController(mi)
	assert.Contains(t, c.Exec("nex"), "next")
	mi.AssertNotCalled(t, "NextHand")
	mi.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
