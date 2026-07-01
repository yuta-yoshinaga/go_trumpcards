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

func TestKingWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	kiMock := new(usecase.MockKingInteractor)
	kiMock.On("Reset").Return(mockOutput)
	kiMock.On("NextDeal").Return(mockOutput)
	kiMock.On("SelectContract", mock.Anything, mock.Anything).Return(mockOutput)
	kiMock.On("Play", mock.Anything).Return(mockOutput)
	kiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	kiMock.On("Hint").Return(mockOutput)
	kiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.KingInteractorIF { return kiMock }
	ctrl := controller.NewKingWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.KingWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
		kiMock.AssertCalled(t, "Reset")
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.KingWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.KingWebConfig{CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		kiMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("select contract", func(t *testing.T) {
		input := controller.KingWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "c", SessionID: "s1"},
			Contract:     6,
			TrumpSuit:    1,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		kiMock.AssertCalled(t, "SelectContract", 6, 1)
	})

	t.Run("play", func(t *testing.T) {
		input := controller.KingWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			HandIndex:    3,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		kiMock.AssertCalled(t, "Play", 3)
	})

	t.Run("next", func(t *testing.T) {
		input := controller.KingWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		kiMock.AssertCalled(t, "NextDeal")
	})

	t.Run("hint", func(t *testing.T) {
		input := controller.KingWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "hint", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		kiMock.AssertCalled(t, "Hint")
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestKingWebConfig_ToConfig(t *testing.T) {
	wc := controller.KingWebConfig{CpuDifficulty: 2}
	c := wc.ToConfig()
	if int(c.CpuDifficulty) != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
