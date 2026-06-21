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

func TestScoponeWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	siMock := new(usecase.MockScoponeInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
	siMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.ScoponeInteractorIF { return siMock }
	ctrl := controller.NewScoponeWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.ScoponeWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.ScoponeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.ScoponeWebConfig{TargetScore: 11, CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		siMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("config command", func(t *testing.T) {
		input := controller.ScoponeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "config", SessionID: "s1"},
			Config:       &controller.ScoponeWebConfig{TargetScore: 21, CpuDifficulty: 1},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play", func(t *testing.T) {
		input := controller.ScoponeWebInput{
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
		input := controller.ScoponeWebInput{
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

func TestScoponeWebConfig_ToConfig(t *testing.T) {
	wc := controller.ScoponeWebConfig{TargetScore: 15, CpuDifficulty: 2}
	c := wc.ToConfig()
	if c.TargetScore != 15 || c.CpuDifficulty != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
