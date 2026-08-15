package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockBalootInteractor() *mockusecase.MockBalootInteractor {
	return new(mockusecase.MockBalootInteractor)
}

func TestBalootCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"DeclareSun", []string{"sun"}},
		{"PassDeclaration", []string{"pass"}},
		{"NextRound", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			bi := newMockBalootInteractor()
			c := NewBalootCuiController(bi)
			bi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// **Sun / Hokom / Pass を取り違えない。** ラウンドの序列そのものが変わる。
func TestBalootCuiControllerDeclareCommandsAreDistinct(t *testing.T) {
	bi := newMockBalootInteractor()
	c := NewBalootCuiController(bi)
	bi.On("DeclareSun").Return("sun")
	bi.On("DeclareHokom", domain.CardDesignHeart).Return("hokom")
	bi.On("PassDeclaration").Return("pass")

	assert.Equal(t, "sun", c.Exec("sun"))
	assert.Equal(t, "hokom", c.Exec("hokom 3"))
	assert.Equal(t, "pass", c.Exec("pass"))
	bi.AssertNumberOfCalls(t, "DeclareSun", 1)
	bi.AssertNumberOfCalls(t, "DeclareHokom", 1)
	bi.AssertNumberOfCalls(t, "PassDeclaration", 1)
}

// hokom は 4 つのスートすべてを受け付ける。
func TestBalootCuiControllerHokomAcceptsEverySuit(t *testing.T) {
	for _, suit := range []int{
		domain.CardDesignSpade, domain.CardDesignClover,
		domain.CardDesignHeart, domain.CardDesignDiamond,
	} {
		bi := newMockBalootInteractor()
		c := NewBalootCuiController(bi)
		bi.On("DeclareHokom", suit).Return("hokom")
		assert.Equal(t, "hokom", c.Exec("hokom "+string(rune('0'+suit))))
		bi.AssertCalled(t, "DeclareHokom", suit)
	}
}

// **範囲外のスートは弾く。** 0 や 5 を通すと切り札の無い Hokom になる。
func TestBalootCuiControllerHokomRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing suit", "hokom", "Suit is required."},
		{"non-numeric", "hokom x", "Invalid suit: x."},
		{"below the range", "hokom 0", "Invalid suit: 0."},
		{"above the range", "hokom 5", "Invalid suit: 5."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bi := newMockBalootInteractor()
			c := NewBalootCuiController(bi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			bi.AssertNotCalled(t, "DeclareHokom", mock.Anything)
		})
	}
}

func TestBalootCuiControllerResetKeepsConfig(t *testing.T) {
	bi := newMockBalootInteractor()
	c := NewBalootCuiController(bi)
	cfg := domain.BalootConfig{Target: 200}
	bi.On("GetConfig").Return(cfg)
	bi.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	bi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestBalootCuiControllerQuit(t *testing.T) {
	c := NewBalootCuiController(newMockBalootInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBalootCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			bi := newMockBalootInteractor()
			c := NewBalootCuiController(bi)
			bi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			bi.AssertCalled(t, "Play", 3)
		})
	}
}

func TestBalootCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bi := newMockBalootInteractor()
			c := NewBalootCuiController(bi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			bi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestBalootCuiControllerUnknownCommand(t *testing.T) {
	bi := newMockBalootInteractor()
	c := NewBalootCuiController(bi)
	assert.Contains(t, c.Exec("pas"), "pass")
	bi.AssertNotCalled(t, "PassDeclaration")
	bi.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
