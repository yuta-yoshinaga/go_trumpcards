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

func TestPresidentWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	piMock := new(usecase.MockPresidentInteractor)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("Play", mock.Anything).Return(mockOutput)
	piMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.PresidentInteractorIF { return piMock }
	ctrl := controller.NewPresidentWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.PresidentWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset full with config", func(t *testing.T) {
		input := controller.PresidentWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.PresidentWebConfig{RevolutionEnabled: true, CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
		piMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play pass", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("play indices", func(t *testing.T) {
		input := controller.PresidentWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
			Indices:      []int{0, 1},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		piMock.AssertCalled(t, "Play", []int{0, 1})
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestPresidentWebConfig_ToConfig(t *testing.T) {
	wc := controller.PresidentWebConfig{
		RevolutionEnabled:     true,
		CardExchangeEnabled:   true,
		PassFieldFlushEnabled: false,
		CpuDifficulty:         2,
	}
	c := wc.ToConfig()
	if !c.RevolutionEnabled || !c.CardExchangeEnabled || c.PassFieldFlushEnabled || c.CpuDifficulty != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
