//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newWizardCuiMock() *mockUsecases.MockWizardInteractor {
	m := new(mockUsecases.MockWizardInteractor)
	m.On("GetConfig").Return(domain.DefaultWizardConfig())
	m.On("ResetWithConfig", domain.DefaultWizardConfig()).Return("reset")
	m.On("Bid", 3).Return("bid3")
	m.On("Play", 0).Return("play0")
	m.On("NextTrick").Return("next")
	m.On("NextRound").Return("nextround")
	m.On("Hint").Return("hint")
	m.On("ActionLog").Return("log")
	return m
}

func TestWizardCuiController_Exec(t *testing.T) {
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
			m := newWizardCuiMock()
			c := controller.NewWizardCuiController(m)
			result := c.Exec(tt.command)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWizardCuiController_Exec_Quit(t *testing.T) {
	m := newWizardCuiMock()
	c := controller.NewWizardCuiController(m)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestWizardCuiController_Exec_SetDifficulty(t *testing.T) {
	m := new(mockUsecases.MockWizardInteractor)
	m.On("GetConfig").Return(domain.DefaultWizardConfig())
	cfg := domain.DefaultWizardConfig()
	cfg.CpuDifficulty = domain.WizardCpuDifficultyHard
	m.On("ResetWithConfig", cfg).Return("hard")

	c := controller.NewWizardCuiController(m)
	assert.Equal(t, "hard", c.Exec("sd 2"))
}

func TestWizardCuiController_Exec_Errors(t *testing.T) {
	m := newWizardCuiMock()
	c := controller.NewWizardCuiController(m)

	// Missing args
	assert.Contains(t, c.Exec("b"), "required")
	assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	assert.Contains(t, c.Exec("sd"), "required")

	// Invalid args
	assert.Contains(t, c.Exec("b abc"), "Invalid")
	assert.Contains(t, c.Exec("sd 99"), "Invalid")
}

func TestWizardCuiController_Exec_UnknownCommand(t *testing.T) {
	m := newWizardCuiMock()
	c := controller.NewWizardCuiController(m)
	result := c.Exec("xyz")
	assert.Contains(t, result, "不明")
}
