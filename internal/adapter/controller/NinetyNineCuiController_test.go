//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newNinetyNineCuiMock() *mockUsecases.MockNinetyNineInteractor {
	m := new(mockUsecases.MockNinetyNineInteractor)
	m.On("GetConfig").Return(domain.DefaultNinetyNineConfig())
	m.On("ResetWithConfig", domain.DefaultNinetyNineConfig()).Return("reset")
	m.On("Bid", []int{0, 1, 2}).Return("bid")
	m.On("Play", 0).Return("play0")
	m.On("NextTrick").Return("next")
	m.On("NextRound").Return("nextround")
	m.On("Hint").Return("hint")
	m.On("ActionLog").Return("log")
	return m
}

func TestNinetyNineCuiController_Exec(t *testing.T) {
	tests := []struct {
		command  string
		expected string
	}{
		{"r", "reset"},
		{"reset", "reset"},
		{"b 0 1 2", "bid"},
		{"bid 0 1 2", "bid"},
		{"p 0", "play0"},
		{"play 0", "play0"},
		{"n", "next"},
		{"next", "next"},
		{"nr", "nextround"},
		{"nextround", "nextround"},
		{"h", "hint"},
		{"hint", "hint"},
		{"log", "log"},
		{"l", "log"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			m := newNinetyNineCuiMock()
			c := controller.NewNinetyNineCuiController(m)
			assert.Equal(t, tt.expected, c.Exec(tt.command))
		})
	}
}

func TestNinetyNineCuiController_Exec_Quit(t *testing.T) {
	m := newNinetyNineCuiMock()
	c := controller.NewNinetyNineCuiController(m)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestNinetyNineCuiController_Exec_SetDifficulty(t *testing.T) {
	m := new(mockUsecases.MockNinetyNineInteractor)
	m.On("GetConfig").Return(domain.DefaultNinetyNineConfig())
	cfg := domain.DefaultNinetyNineConfig()
	cfg.CpuDifficulty = domain.NinetyNineCpuDifficultyHard
	m.On("ResetWithConfig", cfg).Return("hard")

	c := controller.NewNinetyNineCuiController(m)
	assert.Equal(t, "hard", c.Exec("sd 2"))
}

func TestNinetyNineCuiController_Exec_SetTarget(t *testing.T) {
	m := new(mockUsecases.MockNinetyNineInteractor)
	m.On("GetConfig").Return(domain.DefaultNinetyNineConfig())
	cfg := domain.DefaultNinetyNineConfig()
	cfg.TargetScore = 50
	m.On("ResetWithConfig", cfg).Return("target50")

	c := controller.NewNinetyNineCuiController(m)
	assert.Equal(t, "target50", c.Exec("st 50"))
}

func TestNinetyNineCuiController_Exec_Errors(t *testing.T) {
	m := newNinetyNineCuiMock()
	c := controller.NewNinetyNineCuiController(m)

	assert.True(t, msgRejected(c.Exec("b")))
	assert.True(t, msgRejected(c.Exec("b 0 1")))
	assert.True(t, msgRejected(c.Exec("b a b c")))
	assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
	assert.True(t, msgRejected(c.Exec("st")))

	assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	assert.Contains(t, c.Exec("sd 99"), msgInvalidCpuDifficultyPrefix())
	assert.True(t, msgRejected(c.Exec("st 5")))
	assert.True(t, msgRejected(c.Exec("st 9999")))
}

func TestNinetyNineCuiController_Exec_UnknownCommand(t *testing.T) {
	m := newNinetyNineCuiMock()
	c := controller.NewNinetyNineCuiController(m)
	assert.Contains(t, c.Exec("xyz"), "不明")
}
