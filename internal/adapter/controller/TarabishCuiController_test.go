package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockTarabishInteractor() *mockusecase.MockTarabishInteractor {
	return new(mockusecase.MockTarabishInteractor)
}

func TestTarabishCuiControllerSimpleCommands(t *testing.T) {
	for _, tc := range []struct {
		method  string
		aliases []string
	}{
		{"TakeTrump", []string{"t", "take"}},
		{"PassTrump", []string{"pass"}},
		{"NextRound", []string{"n", "next"}},
		{"GiveUp", []string{"g", "giveup"}},
		{"Hint", []string{"h", "hint"}},
		{"ActionLog", []string{"log", "l"}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			ti := newMockTarabishInteractor()
			c := NewTarabishCuiController(ti)
			ti.On(tc.method).Return("output")
			for _, alias := range tc.aliases {
				assert.Equal(t, "output", c.Exec(alias), "alias %q", alias)
			}
		})
	}
}

// **引き受けと見送りを取り違えない。** ラウンドの切り札が変わる。
func TestTarabishCuiControllerTakeAndPassAreDistinct(t *testing.T) {
	ti := newMockTarabishInteractor()
	c := NewTarabishCuiController(ti)
	ti.On("TakeTrump").Return("take")
	ti.On("PassTrump").Return("pass")

	assert.Equal(t, "take", c.Exec("t"))
	assert.Equal(t, "pass", c.Exec("pass"))
	ti.AssertNumberOfCalls(t, "TakeTrump", 1)
	ti.AssertNumberOfCalls(t, "PassTrump", 1)
}

func TestTarabishCuiControllerResetKeepsConfig(t *testing.T) {
	ti := newMockTarabishInteractor()
	c := NewTarabishCuiController(ti)
	cfg := domain.TarabishConfig{Target: 300}
	ti.On("GetConfig").Return(cfg)
	ti.On("ResetWithConfig", cfg).Return("reset")

	for _, alias := range []string{"r", "reset"} {
		assert.Equal(t, "reset", c.Exec(alias))
	}
	ti.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestTarabishCuiControllerQuit(t *testing.T) {
	c := NewTarabishCuiController(newMockTarabishInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestTarabishCuiControllerPlay(t *testing.T) {
	for _, alias := range []string{"p", "play"} {
		t.Run(alias, func(t *testing.T) {
			ti := newMockTarabishInteractor()
			c := NewTarabishCuiController(ti)
			ti.On("Play", 3).Return("played")
			assert.Equal(t, "played", c.Exec(alias+" 3"))
			ti.AssertCalled(t, "Play", 3)
		})
	}
}

func TestTarabishCuiControllerPlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"missing index", "p", msgCardIndexRequired()},
		{"non-numeric", "p abc", msgInvalidCardIndex("abc")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ti := newMockTarabishInteractor()
			c := NewTarabishCuiController(ti)
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ti.AssertNotCalled(t, "Play", mock.Anything)
		})
	}
}

func TestTarabishCuiControllerUnknownCommand(t *testing.T) {
	ti := newMockTarabishInteractor()
	c := NewTarabishCuiController(ti)
	assert.Contains(t, c.Exec("pas"), "pass")
	ti.AssertNotCalled(t, "PassTrump")
	ti.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
}
