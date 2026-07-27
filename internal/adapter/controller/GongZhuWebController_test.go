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

func mustGongZhuOutputJSON(msg string) string {
	out := &controller.GongZhuWebOutput{
		Players:          []*controller.GongZhuWebOutputPlayer{},
		CurrentTrick:     []*controller.WebOutputTrickCard{},
		ExposableIndices: []int{},
		WinnerIdx:        -1,
		WebOutputBase:    controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGongZhuOutputJSON: %v", err))
	}
	return string(b)
}

func TestGongZhuWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockGongZhuInteractor)
	giMock.On("ResetWithConfig", domain.DefaultGongZhuConfig()).Return(mockOutput)
	giMock.On("Expose", []int{0, 1}).Return(mockOutput)
	giMock.On("Expose", []int(nil)).Return(mockOutput)
	giMock.On("Play", 3).Return(mockOutput)
	giMock.On("NextTrick").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.GongZhuInteractorIF { return giMock }
	ctrl := controller.NewGongZhuWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		var input controller.GongZhuWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustGongZhuOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("expose with indices", func(t *testing.T) {
		run(t, `{"command":"expose","cardIndices":[0,1],"sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("expose empty", func(t *testing.T) {
		run(t, `{"command":"expose","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.GongZhuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})

	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustGongZhuOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustGongZhuOutputJSON("param error."), http.StatusBadRequest)
	})
	t.Run("play no cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustGongZhuOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
}

func TestGongZhuWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		limit := 500
		expected := domain.GongZhuConfig{CpuDifficulty: domain.GongZhuCpuDifficultyHard, PointLimit: 500}
		giMock := new(usecase.MockGongZhuInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGongZhuWebController(func() uc.GongZhuInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.GongZhuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.GongZhuWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range values fall back to defaults", func(t *testing.T) {
		diff := 9
		limit := 0
		expected := domain.DefaultGongZhuConfig()
		giMock := new(usecase.MockGongZhuInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGongZhuWebController(func() uc.GongZhuInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.GongZhuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.GongZhuWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultGongZhuConfig()
		giMock := new(usecase.MockGongZhuInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGongZhuWebController(func() uc.GongZhuInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.GongZhuWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestGongZhuWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockGongZhuInteractor)
	c := controller.NewGongZhuWebController(func() uc.GongZhuInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}
