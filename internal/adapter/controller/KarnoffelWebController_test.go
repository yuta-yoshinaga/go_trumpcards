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

func TestKarnoffelWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	kiMock := new(usecase.MockKarnoffelInteractor)
	kiMock.On("ResetWithConfig", domain.DefaultKarnoffelConfig()).Return(mockOutput)
	kiMock.On("PlayCard", 3).Return(mockOutput)
	kiMock.On("NextHand").Return(mockOutput)
	kiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.KarnoffelInteractorIF { return kiMock }
	ctrl := controller.NewKarnoffelWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "n", "next", "log", "l"} {
			var input controller.KarnoffelWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with a parameter", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"p","sessionId":"s1","cardIndex":3}`,
			`{"command":"play","sessionId":"s1","cardIndex":3}`,
		} {
			var input controller.KarnoffelWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		var input controller.KarnoffelWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.KarnoffelWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestKarnoffelWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.KarnoffelWebConfig, expected domain.KarnoffelConfig) {
		t.Helper()
		kiMock := new(usecase.MockKarnoffelInteractor)
		kiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.KarnoffelInteractorIF { return kiMock }
		ctrl := controller.NewKarnoffelWebController(factory)
		defer ctrl.Stop()

		input := controller.KarnoffelWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		kiMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff, hands := 9, 99
		run(t, "cfg-1", &controller.KarnoffelWebConfig{CpuDifficulty: &diff, TargetHands: &hands},
			domain.DefaultKarnoffelConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-2", nil, domain.DefaultKarnoffelConfig())
	})

	t.Run("a valid target comes through", func(t *testing.T) {
		hands := 5
		want := domain.DefaultKarnoffelConfig()
		want.TargetHands = hands
		run(t, "cfg-3", &controller.KarnoffelWebConfig{TargetHands: &hands}, want)
	})
}
