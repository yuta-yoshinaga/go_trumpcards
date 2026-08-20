package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockHasenpfefferInteractor() *mockusecase.MockHasenpfefferInteractor {
	return new(mockusecase.MockHasenpfefferInteractor)
}

func TestHasenpfefferCuiControllerSimpleCommands(t *testing.T) {
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
			hi := newMockHasenpfefferInteractor()
			c := NewHasenpfefferCuiController(hi)
			hi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestHasenpfefferCuiControllerBid(t *testing.T) {
	for _, alias := range []string{"b", "bid"} {
		t.Run(alias, func(t *testing.T) {
			hi := newMockHasenpfefferInteractor()
			c := NewHasenpfefferCuiController(hi)
			hi.On("Bid", 4).Return("bid")
			assert.Equal(t, "bid", c.Exec(alias+" 4"))
			hi.AssertCalled(t, "Bid", 4)
		})
	}
}

// **降りるのは専用コマンド。** `bid 0` を通すと下限の検査がすり抜ける。
func TestHasenpfefferCuiControllerPassIsItsOwnCommand(t *testing.T) {
	hi := newMockHasenpfefferInteractor()
	c := NewHasenpfefferCuiController(hi)
	hi.On("Bid", 0).Return("passed")
	assert.Equal(t, "passed", c.Exec("pass"))
	hi.AssertCalled(t, "Bid", 0)

	// bid 0 は下限未満として弾かれる。
	other := newMockHasenpfefferInteractor()
	c2 := NewHasenpfefferCuiController(other)
	assert.Equal(t, msgKey("invalidBid", "val", "0"), c2.Exec("b 0"))
	other.AssertNotCalled(t, "Bid", mock.Anything)
}

// **範囲外の宣言は弾く。**
func TestHasenpfefferCuiControllerBidRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing bid", "b", msgKey("bidRequired")},
		{"non-numeric", "b x", msgKey("invalidBid", "val", "x")},
		{"below the minimum", "b 2", msgKey("invalidBid", "val", "2")},
		{"above the maximum", "b 7", msgKey("invalidBid", "val", "7")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hi := newMockHasenpfefferInteractor()
			c := NewHasenpfefferCuiController(hi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			hi.AssertNotCalled(t, "Bid", mock.Anything)
		})
	}
}

func TestHasenpfefferCuiControllerDiscard(t *testing.T) {
	for _, alias := range []string{"d", "discard"} {
		t.Run(alias, func(t *testing.T) {
			hi := newMockHasenpfefferInteractor()
			c := NewHasenpfefferCuiController(hi)
			hi.On("Discard", 2, domain.CardDesignHeart).Return("discarded")
			assert.Equal(t, "discarded", c.Exec(alias+" 2 3"))
			hi.AssertCalled(t, "Discard", 2, domain.CardDesignHeart)
		})
	}
}

// **どちらの引数も既定値で埋めない。** 埋めると選んでいないスートが切り札になる。
func TestHasenpfefferCuiControllerDiscardRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"no args", "d", msgCardIndexRequired()},
		{"only the index", "d 2", msgKey("suitRequired")},
		{"non-numeric index", "d x 3", msgInvalidCardIndex("x")},
		{"non-numeric suit", "d 2 x", msgKey("invalidSuit", "val", "x")},
		{"suit below the range", "d 2 0", msgKey("invalidSuit", "val", "0")},
		{"suit above the range", "d 2 5", msgKey("invalidSuit", "val", "5")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hi := newMockHasenpfefferInteractor()
			c := NewHasenpfefferCuiController(hi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			hi.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
		})
	}
}

func TestHasenpfefferCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			hi := newMockHasenpfefferInteractor()
			c := NewHasenpfefferCuiController(hi)
			hi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			hi.AssertCalled(t, "Play", 3)
		})
	}
}

func TestHasenpfefferCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hi := newMockHasenpfefferInteractor()
			c := NewHasenpfefferCuiController(hi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			hi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestHasenpfefferCuiControllerResetKeepsConfig(t *testing.T) {
	hi := newMockHasenpfefferInteractor()
	c := NewHasenpfefferCuiController(hi)
	cfg := domain.HasenpfefferConfig{Target: 15}
	hi.On("GetConfig").Return(cfg)
	hi.On("ResetWithConfig", cfg).Return("reset")
	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	hi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestHasenpfefferCuiControllerQuit(t *testing.T) {
	c := NewHasenpfefferCuiController(newMockHasenpfefferInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestHasenpfefferCuiControllerUnknownCommand(t *testing.T) {
	hi := newMockHasenpfefferInteractor()
	c := NewHasenpfefferCuiController(hi)
	assert.Contains(t, c.Exec("nex"), "next")
	hi.AssertNotCalled(t, "NextHand")
	hi.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
