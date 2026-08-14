package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockPolignacInteractor() *mockusecase.MockPolignacInteractor {
	return new(mockusecase.MockPolignacInteractor)
}

func TestPolignacCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"DeclareCapot", []string{"c", "capot"}},
		{"Pass", []string{"pass"}},
		{"NextRound", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			pi := newMockPolignacInteractor()
			c := NewPolignacCuiController(pi)
			pi.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

func TestPolignacCuiControllerResetKeepsConfig(t *testing.T) {
	pi := newMockPolignacInteractor()
	c := NewPolignacCuiController(pi)
	cfg := domain.PolignacConfig{Rounds: 6}
	pi.On("GetConfig").Return(cfg)
	pi.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	pi.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestPolignacCuiControllerQuit(t *testing.T) {
	c := NewPolignacCuiController(newMockPolignacInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestPolignacCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			pi := newMockPolignacInteractor()
			c := NewPolignacCuiController(pi)
			pi.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			pi.AssertCalled(t, "Play", 3)
		})
	}
}

func TestPolignacCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", "Card index is required."},
		{"non-numeric", "p abc", "Invalid card index: abc."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pi := newMockPolignacInteractor()
			c := NewPolignacCuiController(pi)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			pi.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

// **capot と pass を取り違えない。** どちらもラウンドの行方を決める。
func TestPolignacCuiControllerCapotAndPassAreDistinct(t *testing.T) {
	pi := newMockPolignacInteractor()
	c := NewPolignacCuiController(pi)
	pi.On("DeclareCapot").Return("capot")
	pi.On("Pass").Return("pass")

	assert.Equal(t, "capot", c.Exec("c"))
	assert.Equal(t, "pass", c.Exec("pass"))
	pi.AssertNumberOfCalls(t, "DeclareCapot", 1)
	pi.AssertNumberOfCalls(t, "Pass", 1)
}

func TestPolignacCuiControllerUnknownCommand(t *testing.T) {
	pi := newMockPolignacInteractor()
	c := NewPolignacCuiController(pi)
	assert.Contains(t, c.Exec("capo"), "capot")
	pi.AssertNotCalled(t, "DeclareCapot")
	pi.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
