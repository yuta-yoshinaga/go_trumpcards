//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newStealingBundlesCui() (*controller.StealingBundlesCuiController, *usecase.MockStealingBundlesInteractor) {
	si := new(usecase.MockStealingBundlesInteractor)
	return controller.NewStealingBundlesCuiController(si), si
}

func TestStealingBundlesCuiController_Commands(t *testing.T) {
	c, si := newStealingBundlesCui()
	cfg := domain.DefaultStealingBundlesConfig()
	si.On("GetConfig").Return(cfg)
	si.On("ResetWithConfig", cfg).Return("reset")
	si.On("Take", 2).Return("take")
	si.On("Steal", 1, 3).Return("steal")
	si.On("Trail", 0).Return("trail")
	si.On("GiveUp").Return("giveup")
	si.On("Hint").Return("hint")
	si.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"take", "t 2", "take"},
		{"take long", "take 2", "take"},
		// **略奪は札と相手の 2 つを取ります。**
		{"steal", "s 1 3", "steal"},
		{"steal long", "steal 1 3", "steal"},
		{"trail", "d 0", "trail"},
		{"trail long", "trail 0", "trail"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

// **どちらの引数が欠けているかを言い分けます。**
func TestStealingBundlesCuiController_StealRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"no args", "s", msgCardIndexRequired()},
		{"card not a number", "s x 1", msgInvalidCardIndex("x")},
		{"victim missing", "s 1", msgKey("victimIndexRequired")},
		{"victim not a number", "s 1 y", "Invalid victim index: y."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, si := newStealingBundlesCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Steal")
		})
	}
}

func TestStealingBundlesCuiController_TakeAndTrailRejectBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"take missing", "t", msgCardIndexRequired()},
		{"take not a number", "t x", msgInvalidCardIndex("x")},
		{"trail missing", "d", msgCardIndexRequired()},
		{"trail not a number", "d x", msgInvalidCardIndex("x")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, si := newStealingBundlesCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Take")
			si.AssertNotCalled(t, "Trail")
		})
	}
}

func TestStealingBundlesCuiController_UnknownCommand(t *testing.T) {
	c, _ := newStealingBundlesCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}
