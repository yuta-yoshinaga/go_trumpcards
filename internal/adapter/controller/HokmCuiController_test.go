package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockHokmInteractor() *mockusecase.MockHokmInteractor {
	return new(mockusecase.MockHokmInteractor)
}

func TestHokmCuiControllerSimpleCommands(t *testing.T) {
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
			hi := newMockHokmInteractor()
			c := NewHokmCuiController(hi)
			hi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// 切り札は 4 スートすべてを受け付ける。
func TestHokmCuiControllerTrumpAcceptsEverySuit(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		hi := newMockHokmInteractor()
		c := NewHokmCuiController(hi)
		hi.On("DeclareTrump", suit).Return("trump")
		assert.Equal(t, "trump", c.Exec("trump "+string(rune('0'+suit))))
		hi.AssertCalled(t, "DeclareTrump", suit)
	}
}

// **範囲外のスートは弾く。** 通すと切り札の無いハンドになる。
func TestHokmCuiControllerTrumpRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing suit", "t", msgKey("suitRequired")},
		{"non-numeric", "t x", msgKey("invalidSuit", "val", "x")},
		{"below the range", "t 0", msgKey("invalidSuit", "val", "0")},
		{"above the range", "t 5", msgKey("invalidSuit", "val", "5")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hi := newMockHokmInteractor()
			c := NewHokmCuiController(hi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			hi.AssertNotCalled(t, "DeclareTrump", mock.Anything)
		})
	}
}

func TestHokmCuiControllerResetKeepsConfig(t *testing.T) {
	hi := newMockHokmInteractor()
	c := NewHokmCuiController(hi)
	cfg := domain.HokmConfig{Target: 9}
	hi.On("GetConfig").Return(cfg)
	hi.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	hi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestHokmCuiControllerQuit(t *testing.T) {
	c := NewHokmCuiController(newMockHokmInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestHokmCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			hi := newMockHokmInteractor()
			c := NewHokmCuiController(hi)
			hi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			hi.AssertCalled(t, "Play", 3)
		})
	}
}

func TestHokmCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hi := newMockHokmInteractor()
			c := NewHokmCuiController(hi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			hi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestHokmCuiControllerUnknownCommand(t *testing.T) {
	hi := newMockHokmInteractor()
	c := NewHokmCuiController(hi)
	assert.Contains(t, c.Exec("nex"), "next")
	hi.AssertNotCalled(t, "NextHand")
	hi.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
