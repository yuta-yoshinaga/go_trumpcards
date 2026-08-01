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

func TestLiteratureWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	liMock := new(usecase.MockLiteratureInteractor)
	liMock.On("ResetWithConfig", domain.DefaultLiteratureConfig()).Return(mockOutput)
	liMock.On("Ask", 1, 3, 5).Return(mockOutput)
	liMock.On("Claim", 2, []int{0, 0, 2, 2, 4, 4}).Return(mockOutput)
	liMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.LiteratureInteractorIF { return liMock }
	ctrl := controller.NewLiteratureWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "log", "l"} {
			var input controller.LiteratureWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with parameters", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"a","sessionId":"s1","target":1,"suit":3,"value":5}`,
			`{"command":"ask","sessionId":"s1","target":1,"suit":3,"value":5}`,
			`{"command":"c","sessionId":"s1","halfSuit":2,"holders":[0,0,2,2,4,4]}`,
			`{"command":"claim","sessionId":"s1","halfSuit":2,"holders":[0,0,2,2,4,4]}`,
		} {
			var input controller.LiteratureWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// **要求は 3 つ、宣言は組と 6 席がそろって初めて成立する。**
	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"a","sessionId":"s1"}`,
			`{"command":"a","sessionId":"s1","target":1}`,
			`{"command":"a","sessionId":"s1","target":1,"suit":3}`,
			`{"command":"c","sessionId":"s1"}`,
			`{"command":"c","sessionId":"s1","halfSuit":2}`,
			`{"command":"c","sessionId":"s1","halfSuit":2,"holders":[0,0,2]}`,
			`{"command":"c","sessionId":"s1","halfSuit":2,"holders":[0,0,2,2,4,4,0]}`,
		} {
			var input controller.LiteratureWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.LiteratureWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestLiteratureWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.LiteratureWebConfig, expected domain.LiteratureConfig) {
		t.Helper()
		liMock := new(usecase.MockLiteratureInteractor)
		liMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.LiteratureInteractorIF { return liMock }
		ctrl := controller.NewLiteratureWebController(factory)
		defer ctrl.Stop()

		input := controller.LiteratureWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		liMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		run(t, "cfg-1", &controller.LiteratureWebConfig{CpuDifficulty: &diff}, domain.DefaultLiteratureConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-2", nil, domain.DefaultLiteratureConfig())
	})
}
