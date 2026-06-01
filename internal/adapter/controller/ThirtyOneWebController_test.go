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

func mustThirtyOneOutputJSON(msg string) string {
	out := &controller.ThirtyOneWebOutput{
		Players:        []*controller.ThirtyOneWebOutputPlayer{},
		WinnerIdx:      -1,
		KnockerIdx:     -1,
		ThirtyOneIdx:   -1,
		RoundWinnerIdx: -1,
		RoundLosers:    []int{},
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustThirtyOneOutputJSON: %v", err))
	}
	return string(b)
}

func TestThirtyOneWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockThirtyOneInteractor)
	siMock.On("ResetWithConfig", domain.DefaultThirtyOneConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("Knock").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ThirtyOneInteractorIF { return siMock }
	ctrl := controller.NewThirtyOneWebController(factory)
	defer ctrl.Stop()

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			var input controller.ThirtyOneWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("draw stock and discard", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock", "dd", "drawdiscard"} {
			var input controller.ThirtyOneWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("discard with index", func(t *testing.T) {
		for _, cmd := range []string{"d", "discard"} {
			var input controller.ThirtyOneWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","cardIndex":3,"sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("knock no index needed", func(t *testing.T) {
		for _, cmd := range []string{"k", "knock"} {
			var input controller.ThirtyOneWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("nextround and log", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround", "log", "l"} {
			var input controller.ThirtyOneWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.ThirtyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustThirtyOneOutputJSON("Unsupported command."))
	})

	t.Run("discard no cardIndex", func(t *testing.T) {
		var input controller.ThirtyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustThirtyOneOutputJSON("param error: cardIndex is required."))
	})
}

func TestThirtyOneWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		lives := 5
		expected := domain.ThirtyOneConfig{CpuDifficulty: domain.ThirtyOneCpuDifficultyHard, InitialLives: 5}
		siMock := new(usecase.MockThirtyOneInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.ThirtyOneInteractorIF { return siMock }
		ctrl := controller.NewThirtyOneWebController(factory)
		defer ctrl.Stop()

		input := controller.ThirtyOneWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.ThirtyOneWebConfig{CpuDifficulty: &diff, InitialLives: &lives},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		lives := 0
		expected := domain.DefaultThirtyOneConfig()
		siMock := new(usecase.MockThirtyOneInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.ThirtyOneInteractorIF { return siMock }
		ctrl := controller.NewThirtyOneWebController(factory)
		defer ctrl.Stop()

		input := controller.ThirtyOneWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.ThirtyOneWebConfig{CpuDifficulty: &diff, InitialLives: &lives},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultThirtyOneConfig()
		siMock := new(usecase.MockThirtyOneInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.ThirtyOneInteractorIF { return siMock }
		ctrl := controller.NewThirtyOneWebController(factory)
		defer ctrl.Stop()

		input := controller.ThirtyOneWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-4"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}
