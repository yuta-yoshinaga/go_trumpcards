package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockTeenDoPaanchInteractor() *mockusecase.MockTeenDoPaanchInteractor {
	return new(mockusecase.MockTeenDoPaanchInteractor)
}

func TestTeenDoPaanchCuiControllerSimpleCommands(t *testing.T) {
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
			ti := newMockTeenDoPaanchInteractor()
			c := NewTeenDoPaanchCuiController(ti)
			ti.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// 切り札は 4 スートすべてを受け付ける。
func TestTeenDoPaanchCuiControllerTrumpAcceptsEverySuit(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		ti := newMockTeenDoPaanchInteractor()
		c := NewTeenDoPaanchCuiController(ti)
		ti.On("DeclareTrump", suit).Return("trump")
		assert.Equal(t, "trump", c.Exec("trump "+string(rune('0'+suit))))
		ti.AssertCalled(t, "DeclareTrump", suit)
	}
}

// **範囲外のスートは弾く。** 通すと選んでいないスートが切り札になる。
func TestTeenDoPaanchCuiControllerTrumpRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing suit", "t", msgKey("suitRequired")},
		{"non-numeric", "t x", msgKey("invalidSuit", "val", "x")},
		{"below the range", "t 0", msgKey("invalidSuit", "val", "0")},
		{"above the range", "t 5", msgKey("invalidSuit", "val", "5")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ti := newMockTeenDoPaanchInteractor()
			c := NewTeenDoPaanchCuiController(ti)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ti.AssertNotCalled(t, "DeclareTrump", mock.Anything)
		})
	}
}

func TestTeenDoPaanchCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			ti := newMockTeenDoPaanchInteractor()
			c := NewTeenDoPaanchCuiController(ti)
			ti.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			ti.AssertCalled(t, "Play", 3)
		})
	}
}

func TestTeenDoPaanchCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ti := newMockTeenDoPaanchInteractor()
			c := NewTeenDoPaanchCuiController(ti)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ti.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

// **ノルマを宣言するコマンドは無い。** 3・2・5 は割り当てで、選ぶ余地が無い。
func TestTeenDoPaanchCuiControllerHasNoBidCommand(t *testing.T) {
	for _, cmd := range []string{"b 3", "bid 5", "b", "bid"} {
		ti := newMockTeenDoPaanchInteractor()
		c := NewTeenDoPaanchCuiController(ti)
		assert.NotEmpty(t, c.Exec(cmd), "未知コマンドとして案内を返す: "+cmd)
		ti.AssertNotCalled(t, "Play", mock.Anything)
		ti.AssertNotCalled(t, "DeclareTrump", mock.Anything)
	}
}

func TestTeenDoPaanchCuiControllerResetKeepsConfig(t *testing.T) {
	ti := newMockTeenDoPaanchInteractor()
	c := NewTeenDoPaanchCuiController(ti)
	cfg := domain.TeenDoPaanchConfig{Rounds: 6}
	ti.On("GetConfig").Return(cfg)
	ti.On("ResetWithConfig", cfg).Return("reset")
	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	ti.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestTeenDoPaanchCuiControllerQuit(t *testing.T) {
	c := NewTeenDoPaanchCuiController(newMockTeenDoPaanchInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestTeenDoPaanchCuiControllerUnknownCommand(t *testing.T) {
	ti := newMockTeenDoPaanchInteractor()
	c := NewTeenDoPaanchCuiController(ti)
	assert.Contains(t, c.Exec("nex"), "next")
	ti.AssertNotCalled(t, "NextRound")
	ti.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
