package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockSergeantMajorInteractor() *mockusecase.MockSergeantMajorInteractor {
	return new(mockusecase.MockSergeantMajorInteractor)
}

func TestSergeantMajorCuiControllerSimpleCommands(t *testing.T) {
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
			si := newMockSergeantMajorInteractor()
			c := NewSergeantMajorCuiController(si)
			si.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestSergeantMajorCuiControllerTrumpAcceptsEverySuit(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		si := newMockSergeantMajorInteractor()
		c := NewSergeantMajorCuiController(si)
		si.On("DeclareTrump", suit).Return("trump")
		assert.Equal(t, "trump", c.Exec("trump "+string(rune('0'+suit))))
		si.AssertCalled(t, "DeclareTrump", suit)
	}
}

// **範囲外のスートは弾く。** 通すと選んでいないスートが切り札になる。
func TestSergeantMajorCuiControllerTrumpRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing suit", "t", msgKey("suitRequired")},
		{"non-numeric", "t x", msgKey("invalidSuit", "val", "x")},
		{"below the range", "t 0", msgKey("invalidSuit", "val", "0")},
		{"above the range", "t 5", msgKey("invalidSuit", "val", "5")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			si := newMockSergeantMajorInteractor()
			c := NewSergeantMajorCuiController(si)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "DeclareTrump", mock.Anything)
		})
	}
}

// **捨て札は 4 つの引数をすべて取る。**
func TestSergeantMajorCuiControllerDiscard(t *testing.T) {
	for _, alias := range []string{"d", "discard"} {
		t.Run(alias, func(t *testing.T) {
			si := newMockSergeantMajorInteractor()
			c := NewSergeantMajorCuiController(si)
			si.On("Discard", []int{0, 2, 5, 7}).Return("discarded")
			assert.Equal(t, "discarded", c.Exec(alias+" 0 2 5 7"))
			si.AssertCalled(t, "Discard", []int{0, 2, 5, 7})
		})
	}
}

// **既定値で埋めない。** 埋めると捨てていない札が捨てられる。
func TestSergeantMajorCuiControllerDiscardRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"no args", "d", msgKey("fourIndicesRequired")},
		{"too few", "d 0 1 2", msgKey("fourIndicesRequired")},
		{"non-numeric", "d 0 x 2 3", msgInvalidCardIndex("x")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			si := newMockSergeantMajorInteractor()
			c := NewSergeantMajorCuiController(si)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Discard", mock.Anything)
		})
	}
}

func TestSergeantMajorCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			si := newMockSergeantMajorInteractor()
			c := NewSergeantMajorCuiController(si)
			si.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			si.AssertCalled(t, "Play", 3)
		})
	}
}

func TestSergeantMajorCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			si := newMockSergeantMajorInteractor()
			c := NewSergeantMajorCuiController(si)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			si.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

// **ノルマを宣言するコマンドは無い。** 8・5・3 は席順で決まる。
func TestSergeantMajorCuiControllerHasNoBidCommand(t *testing.T) {
	for _, cmd := range []string{"b 8", "bid 5", "b", "bid"} {
		si := newMockSergeantMajorInteractor()
		c := NewSergeantMajorCuiController(si)
		assert.NotEmpty(t, c.Exec(cmd), "未知コマンドとして案内を返す: "+cmd)
		si.AssertNotCalled(t, "Play", mock.Anything)
		si.AssertNotCalled(t, "DeclareTrump", mock.Anything)
	}
}

func TestSergeantMajorCuiControllerResetKeepsConfig(t *testing.T) {
	si := newMockSergeantMajorInteractor()
	c := NewSergeantMajorCuiController(si)
	cfg := domain.SergeantMajorConfig{Rounds: 6}
	si.On("GetConfig").Return(cfg)
	si.On("ResetWithConfig", cfg).Return("reset")
	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	si.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestSergeantMajorCuiControllerQuit(t *testing.T) {
	c := NewSergeantMajorCuiController(newMockSergeantMajorInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSergeantMajorCuiControllerUnknownCommand(t *testing.T) {
	si := newMockSergeantMajorInteractor()
	c := NewSergeantMajorCuiController(si)
	assert.Contains(t, c.Exec("nex"), "next")
	si.AssertNotCalled(t, "NextRound")
	si.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
