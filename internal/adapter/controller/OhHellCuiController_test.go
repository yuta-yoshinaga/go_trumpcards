//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newOhHellCuiMock() *mockUsecases.MockOhHellInteractor {
	m := new(mockUsecases.MockOhHellInteractor)
	m.On("GetConfig").Return(domain.DefaultOhHellConfig())
	m.On("ResetWithConfig", domain.DefaultOhHellConfig()).Return("reset")
	m.On("Bid", 3).Return("bid3")
	m.On("Play", 0).Return("play0")
	m.On("NextTrick").Return("next")
	m.On("NextRound").Return("nextround")
	m.On("Hint").Return("hint")
	m.On("ActionLog").Return("log")
	return m
}

func TestOhHellCuiController_Exec(t *testing.T) {
	tests := []struct {
		command  string
		expected string
	}{
		{"r", "reset"},
		{"reset", "reset"},
		{"b 3", "bid3"},
		{"bid 3", "bid3"},
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
			m := newOhHellCuiMock()
			c := controller.NewOhHellCuiController(m)
			result := c.Exec(tt.command)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOhHellCuiController_Exec_Quit(t *testing.T) {
	m := newOhHellCuiMock()
	c := controller.NewOhHellCuiController(m)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestOhHellCuiController_Exec_SetDifficulty(t *testing.T) {
	m := new(mockUsecases.MockOhHellInteractor)
	m.On("GetConfig").Return(domain.DefaultOhHellConfig())
	cfg := domain.DefaultOhHellConfig()
	cfg.CpuDifficulty = domain.OhHellCpuDifficultyHard
	m.On("ResetWithConfig", cfg).Return("hard")

	c := controller.NewOhHellCuiController(m)
	assert.Equal(t, "hard", c.Exec("sd 2"))
}

func TestOhHellCuiController_Exec_SetMaxHand(t *testing.T) {
	m := new(mockUsecases.MockOhHellInteractor)
	m.On("GetConfig").Return(domain.DefaultOhHellConfig())
	cfg := domain.DefaultOhHellConfig()
	cfg.MaxHandSize = 5
	m.On("ResetWithConfig", cfg).Return("maxhand5")

	c := controller.NewOhHellCuiController(m)
	assert.Equal(t, "maxhand5", c.Exec("sm 5"))
}

func TestOhHellCuiController_Exec_Errors(t *testing.T) {
	m := newOhHellCuiMock()
	c := controller.NewOhHellCuiController(m)

	// Missing args
	assert.Contains(t, c.Exec("b"), "required")
	assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	assert.Contains(t, c.Exec("sd"), "required")
	assert.Contains(t, c.Exec("sm"), "required")

	// Invalid args
	assert.Contains(t, c.Exec("b abc"), "Invalid")
	assert.Contains(t, c.Exec("sd 99"), "Invalid")
	assert.Contains(t, c.Exec("sm 0"), "Invalid")
	assert.Contains(t, c.Exec("sm 14"), "Invalid")
}

func TestOhHellCuiController_Exec_UnknownCommand(t *testing.T) {
	m := newOhHellCuiMock()
	c := controller.NewOhHellCuiController(m)
	result := c.Exec("xyz")
	assert.Contains(t, result, "不明")
}
