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

func mustYanivOutputJSON(msg string) string {
	out := &controller.YanivWebOutput{
		Players:       []*controller.YanivWebOutputPlayer{},
		PickupCards:   []*controller.WebOutputCard{},
		WinnerIdx:     -1,
		CallerIdx:     -1,
		AsafWinnerIdx: -1,
		RoundScores:   []int{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustYanivOutputJSON: %v", err))
	}
	return string(b)
}

func TestYanivWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockYanivInteractor)
	siMock.On("ResetWithConfig", domain.DefaultYanivConfig()).Return(mockOutput)
	siMock.On("Discard", []int{1, 2}).Return(mockOutput)
	siMock.On("DeclareYaniv").Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromPickup", 0).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.YanivInteractorIF { return siMock }
	ctrl := controller.NewYanivWebController(factory)
	defer ctrl.Stop()

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			var input controller.YanivWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("discard with indices", func(t *testing.T) {
		for _, cmd := range []string{"d", "discard"} {
			var input controller.YanivWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","cardIndices":[1,2],"sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("yaniv", func(t *testing.T) {
		for _, cmd := range []string{"y", "yaniv"} {
			var input controller.YanivWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("draw stock and pickup", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			var input controller.YanivWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
		for _, cmd := range []string{"dp", "drawpickup"} {
			var input controller.YanivWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","end":0,"sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("nextround and log", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround", "log", "l"} {
			var input controller.YanivWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.YanivWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustYanivOutputJSON("Unsupported command."))
	})

	t.Run("drawpickup no end", func(t *testing.T) {
		var input controller.YanivWebInput
		_ = json.Unmarshal([]byte(`{"command":"dp","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustYanivOutputJSON("param error: end is required."))
	})
}

func TestYanivWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 100
		expected := domain.YanivConfig{CpuDifficulty: domain.YanivCpuDifficultyHard, ScoreLimit: 100}
		siMock := new(usecase.MockYanivInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.YanivInteractorIF { return siMock }
		ctrl := controller.NewYanivWebController(factory)
		defer ctrl.Stop()

		input := controller.YanivWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.YanivWebConfig{CpuDifficulty: &diff, ScoreLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		limit := 0
		expected := domain.DefaultYanivConfig()
		siMock := new(usecase.MockYanivInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.YanivInteractorIF { return siMock }
		ctrl := controller.NewYanivWebController(factory)
		defer ctrl.Stop()

		input := controller.YanivWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.YanivWebConfig{CpuDifficulty: &diff, ScoreLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultYanivConfig()
		siMock := new(usecase.MockYanivInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.YanivInteractorIF { return siMock }
		ctrl := controller.NewYanivWebController(factory)
		defer ctrl.Stop()

		input := controller.YanivWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-4"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}
