package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestScopaWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	siMock := new(usecase.MockScopaInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
	siMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.ScopaInteractorIF { return siMock }
	ctrl := controller.NewScopaWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.ScopaWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.ScopaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.ScopaWebConfig{TargetScore: 11, CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		siMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play", func(t *testing.T) {
		input := controller.ScopaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			HandIndex:    0,
			TableIndices: []int{0, 1},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		siMock.AssertCalled(t, "Play", 0, []int{0, 1})
	})

	t.Run("next", func(t *testing.T) {
		input := controller.ScopaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		siMock.AssertCalled(t, "NextRound")
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestScopaWebConfig_ToConfig(t *testing.T) {
	wc := controller.ScopaWebConfig{TargetScore: 15, CpuDifficulty: 2}
	c := wc.ToConfig()
	if c.TargetScore != 15 || c.CpuDifficulty != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
