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

func TestShengJiWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockShengJiInteractor)
	siMock.On("ResetWithConfig", domain.DefaultShengJiConfig()).Return(mockOutput)
	siMock.On("Declare", 3).Return(mockOutput)
	siMock.On("Declare", 0).Return(mockOutput)
	siMock.On("BuryKitty", []int{0, 1, 2, 3, 4, 5, 6, 7}).Return(mockOutput)
	siMock.On("Play", []int{0, 1}).Return(mockOutput)
	siMock.On("NextHand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ShengJiInteractorIF { return siMock }
	ctrl := controller.NewShengJiWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "n", "next", "log", "l"} {
			var input controller.ShengJiWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with parameters", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"d","sessionId":"s1","suit":3}`,
			`{"command":"declare","sessionId":"s1","suit":3}`,
			`{"command":"b","sessionId":"s1","cardIndexes":[0,1,2,3,4,5,6,7]}`,
			`{"command":"bury","sessionId":"s1","cardIndexes":[0,1,2,3,4,5,6,7]}`,
			`{"command":"p","sessionId":"s1","cardIndexes":[0,1]}`,
			`{"command":"play","sessionId":"s1","cardIndexes":[0,1]}`,
		} {
			var input controller.ShengJiWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// **0 はパスという意味を持つ。**省略と区別できないと亮牌を降りられない。
	t.Run("suit zero is a pass, not an omission", func(t *testing.T) {
		var input controller.ShengJiWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1","suit":0}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "Declare", 0)
	})

	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"d","sessionId":"s1"}`,
			`{"command":"b","sessionId":"s1"}`,
			`{"command":"b","sessionId":"s1","cardIndexes":[]}`,
			`{"command":"p","sessionId":"s1"}`,
			`{"command":"p","sessionId":"s1","cardIndexes":[]}`,
		} {
			var input controller.ShengJiWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.ShengJiWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestShengJiWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.ShengJiWebConfig, expected domain.ShengJiConfig) {
		t.Helper()
		siMock := new(usecase.MockShengJiInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.ShengJiInteractorIF { return siMock }
		ctrl := controller.NewShengJiWebController(factory)
		defer ctrl.Stop()

		input := controller.ShengJiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		run(t, "cfg-1", &controller.ShengJiWebConfig{CpuDifficulty: &diff}, domain.DefaultShengJiConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-2", nil, domain.DefaultShengJiConfig())
	})
}
