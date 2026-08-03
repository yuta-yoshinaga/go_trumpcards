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

func TestGuandanWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":1}`

	giMock := new(usecase.MockGuandanInteractor)
	giMock.On("ResetWithConfig", domain.DefaultGuandanConfig()).Return(mockOutput)
	giMock.On("PlayCards", []int{0, 1, 2}).Return(mockOutput)
	giMock.On("Pass").Return(mockOutput)
	giMock.On("ReturnTribute", 3).Return(mockOutput)
	giMock.On("NextHand").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.GuandanInteractorIF { return giMock }
	ctrl := controller.NewGuandanWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "ps", "pass", "n", "next", "log", "l"} {
			var input controller.GuandanWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with parameters", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"p","sessionId":"s1","cardIndexes":[0,1,2]}`,
			`{"command":"play","sessionId":"s1","cardIndexes":[0,1,2]}`,
			`{"command":"t","sessionId":"s1","cardIndex":3}`,
			`{"command":"tribute","sessionId":"s1","cardIndex":3}`,
		} {
			var input controller.GuandanWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// **出す札も還貢の札も省略できない。**
	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"p","sessionId":"s1"}`,
			`{"command":"p","sessionId":"s1","cardIndexes":[]}`,
			`{"command":"t","sessionId":"s1"}`,
		} {
			var input controller.GuandanWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	// **0 は有効な添字。**ポインタで受けていないと省略と区別できない。
	t.Run("card index zero is accepted", func(t *testing.T) {
		giMock.On("ReturnTribute", 0).Return(mockOutput)
		var input controller.GuandanWebInput
		_ = json.Unmarshal([]byte(`{"command":"t","sessionId":"s1","cardIndex":0}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ReturnTribute", 0)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.GuandanWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestGuandanWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.GuandanWebConfig, expected domain.GuandanConfig) {
		t.Helper()
		giMock := new(usecase.MockGuandanInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.GuandanInteractorIF { return giMock }
		ctrl := controller.NewGuandanWebController(factory)
		defer ctrl.Stop()

		input := controller.GuandanWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		run(t, "cfg-1", &controller.GuandanWebConfig{CpuDifficulty: &diff}, domain.DefaultGuandanConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-2", nil, domain.DefaultGuandanConfig())
	})
}
