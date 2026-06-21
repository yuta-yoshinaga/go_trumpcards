//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestCuckooWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockCuckooInteractor)
	siMock.On("ResetWithConfig", domain.DefaultCuckooConfig()).Return(mockOutput)
	siMock.On("Keep").Return(mockOutput)
	siMock.On("Swap").Return(mockOutput)
	siMock.On("Refuse").Return(mockOutput)
	siMock.On("AcceptSwap").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.CuckooInteractorIF { return siMock }
	ctrl := controller.NewCuckooWebController(factory)
	defer ctrl.Stop()

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			var input controller.CuckooWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("keep swap refuse accept", func(t *testing.T) {
		for _, cmd := range []string{"k", "keep", "s", "swap", "rf", "refuse", "ac", "accept"} {
			var input controller.CuckooWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("nextround and log", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround", "log", "l"} {
			var input controller.CuckooWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.CuckooWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestCuckooWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		lives := 5
		expected := domain.CuckooConfig{CpuDifficulty: domain.CuckooCpuDifficultyHard, InitialLives: 5}
		siMock := new(usecase.MockCuckooInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CuckooInteractorIF { return siMock }
		ctrl := controller.NewCuckooWebController(factory)
		defer ctrl.Stop()

		input := controller.CuckooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.CuckooWebConfig{CpuDifficulty: &diff, InitialLives: &lives},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		lives := 0
		expected := domain.DefaultCuckooConfig()
		siMock := new(usecase.MockCuckooInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CuckooInteractorIF { return siMock }
		ctrl := controller.NewCuckooWebController(factory)
		defer ctrl.Stop()

		input := controller.CuckooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.CuckooWebConfig{CpuDifficulty: &diff, InitialLives: &lives},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultCuckooConfig()
		siMock := new(usecase.MockCuckooInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CuckooInteractorIF { return siMock }
		ctrl := controller.NewCuckooWebController(factory)
		defer ctrl.Stop()

		input := controller.CuckooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-4"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}
