package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockShelemInteractor() *mockusecase.MockShelemInteractor {
	return new(mockusecase.MockShelemInteractor)
}

func TestShelemCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"BidShelem", []string{"shelem"}},
		{"Pass", []string{"pass"}},
		{"NextRound", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			si := newMockShelemInteractor()
			c := NewShelemCuiController(si)
			si.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// **入札・Shelem・降りは別のコマンド。** 取り違えると契約が変わる。
func TestShelemCuiControllerBiddingCommandsAreDistinct(t *testing.T) {
	si := newMockShelemInteractor()
	c := NewShelemCuiController(si)
	si.On("Bid", 120).Return("bid")
	si.On("BidShelem").Return("shelem")
	si.On("Pass").Return("pass")

	assert.Equal(t, "bid", c.Exec("b 120"))
	assert.Equal(t, "shelem", c.Exec("shelem"))
	assert.Equal(t, "pass", c.Exec("pass"))
	si.AssertNumberOfCalls(t, "Bid", 1)
	si.AssertNumberOfCalls(t, "BidShelem", 1)
	si.AssertNumberOfCalls(t, "Pass", 1)
}

// 入札は下限と上限を受け付ける。
func TestShelemCuiControllerBidAcceptsTheRange(t *testing.T) {
	for _, bid := range []int{domain.ShelemMinBid, domain.ShelemMaxBid} {
		si := newMockShelemInteractor()
		c := NewShelemCuiController(si)
		si.On("Bid", bid).Return("bid")
		assert.Equal(t, "bid", c.Exec("b "+itoa(bid)))
		si.AssertCalled(t, "Bid", bid)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

// **範囲外の入札は弾く。** 100 未満も 165 超も通さない。
func TestShelemCuiControllerBidRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing bid", "b", "Bid is required."},
		{"non-numeric", "b x", "Invalid bid: x."},
		{"below the minimum", "b 95", "Invalid bid: 95."},
		{"above the maximum", "b 200", "Invalid bid: 200."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			si := newMockShelemInteractor()
			c := NewShelemCuiController(si)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Bid", mock.Anything)
		})
	}
}

// **捨て札は 4 つのインデックス + スートの 5 引数。** どれも既定値で埋めない。
func TestShelemCuiControllerDiscardTakesFiveArguments(t *testing.T) {
	si := newMockShelemInteractor()
	c := NewShelemCuiController(si)
	si.On("Discard", []int{0, 3, 7, 11}, domain.CardDesignHeart).Return("discarded")

	assert.Equal(t, "discarded", c.Exec("d 0 3 7 11 3"))
	si.AssertCalled(t, "Discard", []int{0, 3, 7, 11}, domain.CardDesignHeart)
}

func TestShelemCuiControllerDiscardRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"no args", "d", "Four card indices and a suit are required."},
		{"too few", "d 0 1 2 3", "Four card indices and a suit are required."},
		{"non-numeric index", "d 0 x 2 3 1", "Invalid card index: x."},
		{"suit out of range", "d 0 1 2 3 5", "Invalid suit: 5."},
		{"non-numeric suit", "d 0 1 2 3 x", "Invalid suit: x."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			si := newMockShelemInteractor()
			c := NewShelemCuiController(si)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Discard", mock.Anything, mock.Anything)
		})
	}
}

func TestShelemCuiControllerResetKeepsConfig(t *testing.T) {
	si := newMockShelemInteractor()
	c := NewShelemCuiController(si)
	cfg := domain.ShelemConfig{Target: 700}
	si.On("GetConfig").Return(cfg)
	si.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	si.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestShelemCuiControllerQuit(t *testing.T) {
	c := NewShelemCuiController(newMockShelemInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestShelemCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			si := newMockShelemInteractor()
			c := NewShelemCuiController(si)
			si.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			si.AssertCalled(t, "Play", 3)
		})
	}
}

func TestShelemCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", "Card index is required."},
		{"non-numeric", "p abc", "Invalid card index: abc."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			si := newMockShelemInteractor()
			c := NewShelemCuiController(si)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestShelemCuiControllerUnknownCommand(t *testing.T) {
	si := newMockShelemInteractor()
	c := NewShelemCuiController(si)
	assert.Contains(t, c.Exec("pas"), "pass")
	si.AssertNotCalled(t, "Pass")
	si.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
