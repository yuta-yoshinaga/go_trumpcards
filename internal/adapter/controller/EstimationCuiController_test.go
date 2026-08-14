package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockEstimationInteractor() *mockusecase.MockEstimationInteractor {
	return new(mockusecase.MockEstimationInteractor)
}

func TestEstimationCuiControllerSimpleCommands(t *testing.T) {
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
			ei := newMockEstimationInteractor()
			c := NewEstimationCuiController(ei)
			ei.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// **切り札と宣言を取り違えない。** どちらも数値引数なので、綴りだけが違う。
func TestEstimationCuiControllerTrumpAndBidAreDistinct(t *testing.T) {
	ei := newMockEstimationInteractor()
	c := NewEstimationCuiController(ei)
	ei.On("SelectTrump", 3).Return("trump")
	ei.On("Bid", 3).Return("bid")

	assert.Equal(t, "trump", c.Exec("t 3"))
	assert.Equal(t, "bid", c.Exec("b 3"))
	ei.AssertNumberOfCalls(t, "SelectTrump", 1)
	ei.AssertNumberOfCalls(t, "Bid", 1)
}

// 切り札は 4 スートすべてを受け付ける。
func TestEstimationCuiControllerTrumpAcceptsEverySuit(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		ei := newMockEstimationInteractor()
		c := NewEstimationCuiController(ei)
		ei.On("SelectTrump", suit).Return("trump")
		assert.Equal(t, "trump", c.Exec("trump "+string(rune('0'+suit))))
		ei.AssertCalled(t, "SelectTrump", suit)
	}
}

// **範囲外のスートは弾く。** 0 や 5 を通すと切り札の無いラウンドになる。
func TestEstimationCuiControllerTrumpRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing suit", "t", "Suit is required."},
		{"non-numeric", "t x", "Invalid suit: x."},
		{"below the range", "t 0", "Invalid suit: 0."},
		{"above the range", "t 5", "Invalid suit: 5."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ei := newMockEstimationInteractor()
			c := NewEstimationCuiController(ei)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ei.AssertNotCalled(t, "SelectTrump", mock.Anything)
		})
	}
}

// **0 は合法な宣言（Dash Call）。** 下限で弾いてはいけない。
func TestEstimationCuiControllerBidAcceptsZeroAndThirteen(t *testing.T) {
	for _, bid := range []int{0, domain.EstimationHandSize} {
		ei := newMockEstimationInteractor()
		c := NewEstimationCuiController(ei)
		ei.On("Bid", bid).Return("bid")
		cmd := "b 0"
		if bid == domain.EstimationHandSize {
			cmd = "b 13"
		}
		assert.Equal(t, "bid", c.Exec(cmd))
		ei.AssertCalled(t, "Bid", bid)
	}
}

func TestEstimationCuiControllerBidRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing bid", "b", "Bid is required."},
		{"non-numeric", "b x", "Invalid bid: x."},
		{"negative", "b -1", "Invalid bid: -1."},
		{"above the hand size", "b 14", "Invalid bid: 14."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ei := newMockEstimationInteractor()
			c := NewEstimationCuiController(ei)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ei.AssertNotCalled(t, "Bid", mock.Anything)
		})
	}
}

func TestEstimationCuiControllerResetKeepsConfig(t *testing.T) {
	ei := newMockEstimationInteractor()
	c := NewEstimationCuiController(ei)
	cfg := domain.EstimationConfig{Rounds: 7}
	ei.On("GetConfig").Return(cfg)
	ei.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	ei.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestEstimationCuiControllerQuit(t *testing.T) {
	c := NewEstimationCuiController(newMockEstimationInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestEstimationCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			ei := newMockEstimationInteractor()
			c := NewEstimationCuiController(ei)
			ei.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			ei.AssertCalled(t, "Play", 3)
		})
	}
}

func TestEstimationCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", "Card index is required."},
		{"non-numeric", "p abc", "Invalid card index: abc."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ei := newMockEstimationInteractor()
			c := NewEstimationCuiController(ei)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ei.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestEstimationCuiControllerUnknownCommand(t *testing.T) {
	ei := newMockEstimationInteractor()
	c := NewEstimationCuiController(ei)
	assert.Contains(t, c.Exec("nex"), "next")
	ei.AssertNotCalled(t, "NextRound")
	ei.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
