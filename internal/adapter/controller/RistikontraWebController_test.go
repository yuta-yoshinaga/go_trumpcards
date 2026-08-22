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

func TestRistikontraWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	piMock := new(usecase.MockRistikontraInteractor)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("NextRound").Return(mockOutput)
	piMock.On("Play", mock.Anything).Return(mockOutput)
	piMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.RistikontraInteractorIF { return piMock }
	ctrl := controller.NewRistikontraWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.RistikontraWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.RistikontraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.RistikontraWebConfig{PlayerCnt: 3, CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		piMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play", func(t *testing.T) {
		input := controller.RistikontraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			HandIndex:    1,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		piMock.AssertCalled(t, "Play", 1)
	})

	t.Run("next", func(t *testing.T) {
		input := controller.RistikontraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		piMock.AssertCalled(t, "NextRound")
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestRistikontraWebConfig_ToConfig(t *testing.T) {
	wc := controller.RistikontraWebConfig{PlayerCnt: 2, CpuDifficulty: 2}
	c := wc.ToConfig()
	if c.PlayerCnt != 2 || c.CpuDifficulty != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
