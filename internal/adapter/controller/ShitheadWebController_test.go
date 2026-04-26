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

func TestShitheadWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	siMock := new(usecase.MockShitheadInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Play", mock.Anything).Return(mockOutput)
	siMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.ShitheadInteractorIF { return siMock }
	ctrl := controller.NewShitheadWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.ShitheadWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.ShitheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.ShitheadWebConfig{MagicTwo: true, CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
		siMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play pickup", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("play indices", func(t *testing.T) {
		input := controller.ShitheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
			Indices:      []int{0, 1},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		siMock.AssertCalled(t, "Play", []int{0, 1})
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestShitheadWebConfig_ToConfig(t *testing.T) {
	wc := controller.ShitheadWebConfig{
		MagicTwo:        true,
		MagicSeven:      false,
		MagicEight:      true,
		MagicTen:        true,
		FourOfAKindBurn: true,
		CpuDifficulty:   2,
	}
	c := wc.ToConfig()
	if !c.MagicTwo || c.MagicSeven || !c.MagicEight || !c.MagicTen || !c.FourOfAKindBurn || c.CpuDifficulty != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
