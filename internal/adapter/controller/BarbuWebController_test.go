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

func TestBarbuWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	biMock := new(usecase.MockBarbuInteractor)
	biMock.On("Reset").Return(mockOutput)
	biMock.On("NextDeal").Return(mockOutput)
	biMock.On("SelectContract", mock.Anything, mock.Anything).Return(mockOutput)
	biMock.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
	biMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.BarbuInteractorIF { return biMock }
	ctrl := controller.NewBarbuWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.BarbuWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.BarbuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.BarbuWebConfig{CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		biMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("select contract", func(t *testing.T) {
		input := controller.BarbuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "c", SessionID: "s1"},
			Contract:     5,
			TrumpSuit:    1,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		biMock.AssertCalled(t, "SelectContract", 5, 1)
	})

	t.Run("play", func(t *testing.T) {
		input := controller.BarbuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			HandIndex:    3,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Play", 3, []int(nil))
	})

	t.Run("next", func(t *testing.T) {
		input := controller.BarbuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		biMock.AssertCalled(t, "NextDeal")
	})

	t.Run("too many indices returns 400", func(t *testing.T) {
		idxs := make([]int, 20)
		input := controller.BarbuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			TableIndices: idxs,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestBarbuWebConfig_ToConfig(t *testing.T) {
	wc := controller.BarbuWebConfig{CpuDifficulty: 2}
	c := wc.ToConfig()
	if int(c.CpuDifficulty) != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
