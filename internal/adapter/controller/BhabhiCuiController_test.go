package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockBhabhiInteractor() *mockusecase.MockBhabhiInteractor {
	return new(mockusecase.MockBhabhiInteractor)
}

func TestBhabhiCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			bi := newMockBhabhiInteractor()
			c := NewBhabhiCuiController(bi)
			bi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestBhabhiCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			bi := newMockBhabhiInteractor()
			c := NewBhabhiCuiController(bi)
			bi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			bi.AssertCalled(t, "Play", 3)
		})
	}
}

func TestBhabhiCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", "Card index is required."},
		{"non-numeric", "p abc", "Invalid card index: abc."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bi := newMockBhabhiInteractor()
			c := NewBhabhiCuiController(bi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			bi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

// **次のハンドへ進むコマンドは無い。** 配り切りの 1 ゲームで終わるので、
// 受け付けるとありもしない区切りを案内することになる。
func TestBhabhiCuiControllerHasNoNextCommand(t *testing.T) {
	for _, cmd := range []string{"n", "next"} {
		bi := newMockBhabhiInteractor()
		c := NewBhabhiCuiController(bi)
		assert.NotEmpty(t, c.Exec(cmd), "未知コマンドとして案内を返す: "+cmd)
		bi.AssertNotCalled(t, "Play", mock.Anything)
		bi.AssertNotCalled(t, "GiveUp")
	}
}

func TestBhabhiCuiControllerResetKeepsConfig(t *testing.T) {
	bi := newMockBhabhiInteractor()
	c := NewBhabhiCuiController(bi)
	cfg := domain.BhabhiConfig{PlayerCnt: 5}
	bi.On("GetConfig").Return(cfg)
	bi.On("ResetWithConfig", cfg).Return("reset")
	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	bi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestBhabhiCuiControllerQuit(t *testing.T) {
	c := NewBhabhiCuiController(newMockBhabhiInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBhabhiCuiControllerUnknownCommand(t *testing.T) {
	bi := newMockBhabhiInteractor()
	c := NewBhabhiCuiController(bi)
	assert.Contains(t, c.Exec("pla"), "play")
	bi.AssertNotCalled(t, "Play", mock.Anything)
	bi.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
