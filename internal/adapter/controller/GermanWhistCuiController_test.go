package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockGermanWhistInteractor() *mockusecase.MockGermanWhistInteractor {
	return new(mockusecase.MockGermanWhistInteractor)
}

func TestGermanWhistCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"Reset", []string{"r", "reset"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			gi := newMockGermanWhistInteractor()
			c := NewGermanWhistCuiController(gi)
			gi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestGermanWhistCuiControllerQuit(t *testing.T) {
	c := NewGermanWhistCuiController(newMockGermanWhistInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestGermanWhistCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			gi := newMockGermanWhistInteractor()
			c := NewGermanWhistCuiController(gi)
			gi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			gi.AssertCalled(t, "Play", 3)
		})
	}
}

// 引数の欠落と不正はそれぞれ固有の文言で断り、インタラクターには届かない。
func TestGermanWhistCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gi := newMockGermanWhistInteractor()
			c := NewGermanWhistCuiController(gi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			gi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

// 綴り間違いは候補を出す。ここで Play や Reset が呼ばれると盤面が壊れる。
func TestGermanWhistCuiControllerUnknownCommand(t *testing.T) {
	gi := newMockGermanWhistInteractor()
	c := NewGermanWhistCuiController(gi)
	assert.Contains(t, c.Exec("pla 1"), "play", "近いコマンドを提案する")
	gi.AssertNotCalled(t, "Play", mock.Anything)
	gi.AssertNotCalled(t, "Reset")
}
