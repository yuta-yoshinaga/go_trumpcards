package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockIsraeliWhistInteractor() *mockusecase.MockIsraeliWhistInteractor {
	return new(mockusecase.MockIsraeliWhistInteractor)
}

func TestIsraeliWhistCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"AuctionPass", []string{"pass"}},
		{"NextRound", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			wi := newMockIsraeliWhistInteractor()
			c := NewIsraeliWhistCuiController(wi)
			wi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// **入札は数とスートの 2 引数。** 順番を取り違えると別の入札になる。
func TestIsraeliWhistCuiControllerAuctionTakesBothArguments(t *testing.T) {
	for _, tc := range []struct {
		cmd       string
		bid, suit int
	}{
		{"a 5 1", 5, domain.CardDesignSpade},
		{"auction 9 2", 9, domain.CardDesignClover},
		{"a 13 4", 13, domain.CardDesignDiamond},
	} {
		wi := newMockIsraeliWhistInteractor()
		c := NewIsraeliWhistCuiController(wi)
		wi.On("AuctionBid", tc.bid, tc.suit).Return("auction")
		assert.Equal(t, "auction", c.Exec(tc.cmd))
		wi.AssertCalled(t, "AuctionBid", tc.bid, tc.suit)
	}
}

// **最低入札を下回る／スート欠けは弾く。** 既定値で埋めない。
func TestIsraeliWhistCuiControllerAuctionRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing both", "a", "Bid is required."},
		{"missing suit", "a 7", "Suit is required."},
		{"non-numeric bid", "a x 1", "Invalid bid: x."},
		{"non-numeric suit", "a 7 x", "Invalid suit: x."},
		{"below the minimum", "a 4 1", "Invalid bid: 4."},
		{"suit out of range", "a 7 5", "Invalid suit: 5."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wi := newMockIsraeliWhistInteractor()
			c := NewIsraeliWhistCuiController(wi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			wi.AssertNotCalled(t, "AuctionBid", mock.Anything, mock.Anything)
		})
	}
}

// **入札と宣言を取り違えない。** どちらも数値引数で、綴りだけが違う。
func TestIsraeliWhistCuiControllerAuctionAndBidAreDistinct(t *testing.T) {
	wi := newMockIsraeliWhistInteractor()
	c := NewIsraeliWhistCuiController(wi)
	wi.On("AuctionBid", 7, domain.CardDesignHeart).Return("auction")
	wi.On("Bid", 7).Return("bid")

	assert.Equal(t, "auction", c.Exec("a 7 3"))
	assert.Equal(t, "bid", c.Exec("b 7"))
	wi.AssertNumberOfCalls(t, "AuctionBid", 1)
	wi.AssertNumberOfCalls(t, "Bid", 1)
}

// **0 は合法な宣言。** オークションと違って下限は 0。
func TestIsraeliWhistCuiControllerBidAcceptsZeroAndThirteen(t *testing.T) {
	for _, tc := range []struct {
		cmd string
		bid int
	}{{"b 0", 0}, {"bid 13", domain.IsraeliWhistHandSize}} {
		wi := newMockIsraeliWhistInteractor()
		c := NewIsraeliWhistCuiController(wi)
		wi.On("Bid", tc.bid).Return("bid")
		assert.Equal(t, "bid", c.Exec(tc.cmd))
		wi.AssertCalled(t, "Bid", tc.bid)
	}
}

func TestIsraeliWhistCuiControllerBidRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing bid", "b", "Bid is required."},
		{"non-numeric", "b x", "Invalid bid: x."},
		{"negative", "b -1", "Invalid bid: -1."},
		{"above the hand size", "b 14", "Invalid bid: 14."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wi := newMockIsraeliWhistInteractor()
			c := NewIsraeliWhistCuiController(wi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			wi.AssertNotCalled(t, "Bid", mock.Anything)
		})
	}
}

func TestIsraeliWhistCuiControllerResetKeepsConfig(t *testing.T) {
	wi := newMockIsraeliWhistInteractor()
	c := NewIsraeliWhistCuiController(wi)
	cfg := domain.IsraeliWhistConfig{Rounds: 6}
	wi.On("GetConfig").Return(cfg)
	wi.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	wi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestIsraeliWhistCuiControllerQuit(t *testing.T) {
	c := NewIsraeliWhistCuiController(newMockIsraeliWhistInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestIsraeliWhistCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			wi := newMockIsraeliWhistInteractor()
			c := NewIsraeliWhistCuiController(wi)
			wi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			wi.AssertCalled(t, "Play", 3)
		})
	}
}

func TestIsraeliWhistCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wi := newMockIsraeliWhistInteractor()
			c := NewIsraeliWhistCuiController(wi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			wi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestIsraeliWhistCuiControllerUnknownCommand(t *testing.T) {
	wi := newMockIsraeliWhistInteractor()
	c := NewIsraeliWhistCuiController(wi)
	assert.Contains(t, c.Exec("pas"), "pass")
	wi.AssertNotCalled(t, "AuctionPass")
	wi.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
