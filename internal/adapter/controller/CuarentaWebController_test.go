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

func TestCuarentaWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	ciMock := new(usecase.MockCuarentaInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("NextRound").Return(mockOutput)
	ciMock.On("Play", mock.Anything).Return(mockOutput)
	ciMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.CuarentaInteractorIF { return ciMock }
	ctrl := controller.NewCuarentaWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.CuarentaWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.CuarentaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.CuarentaWebConfig{TargetScore: 40, CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play", func(t *testing.T) {
		input := controller.CuarentaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			HandIndex:    1,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "Play", 1)
	})

	t.Run("next", func(t *testing.T) {
		input := controller.CuarentaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "NextRound")
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestCuarentaWebConfig_ToConfig(t *testing.T) {
	wc := controller.CuarentaWebConfig{TargetScore: 35, CpuDifficulty: 2}
	c := wc.ToConfig()
	if c.TargetScore != 35 || c.CpuDifficulty != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
