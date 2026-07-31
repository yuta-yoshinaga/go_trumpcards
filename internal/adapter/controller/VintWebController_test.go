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

func TestVintWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockVintInteractor)
	siMock.On("ResetWithConfig", domain.DefaultVintConfig()).Return(mockOutput)
	siMock.On("Bid", 3, domain.VintDenomHeart).Return(mockOutput)
	siMock.On("PassBid").Return(mockOutput)
	siMock.On("PlayCard", 4).Return(mockOutput)
	siMock.On("NextHand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.VintInteractorIF { return siMock }
	ctrl := controller.NewVintWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "ps", "pass", "n", "next", "log", "l"} {
			var input controller.VintWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with a parameter", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1","level":3,"denom":3}`,
			`{"command":"bid","sessionId":"s1","level":3,"denom":3}`,
			`{"command":"p","sessionId":"s1","cardIndex":4}`,
			`{"command":"play","sessionId":"s1","cardIndex":4}`,
		} {
			var input controller.VintWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// **level と denom は両方必須。**片方だけでは宣言が決まらない。
	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1"}`,
			`{"command":"b","sessionId":"s1","level":3}`,
			`{"command":"b","sessionId":"s1","denom":3}`,
			`{"command":"p","sessionId":"s1"}`,
		} {
			var input controller.VintWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.VintWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestVintWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.VintWebConfig, expected domain.VintConfig) {
		t.Helper()
		siMock := new(usecase.MockVintInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.VintInteractorIF { return siMock }
		ctrl := controller.NewVintWebController(factory)
		defer ctrl.Stop()

		input := controller.VintWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		run(t, "cfg-1", &controller.VintWebConfig{CpuDifficulty: &diff}, domain.DefaultVintConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-2", nil, domain.DefaultVintConfig())
	})
}
