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

func mustOpenFaceChineseOutputJSON(msg string) string {
	out := &controller.OpenFaceChineseWebOutput{
		Players:       []*controller.OpenFaceChineseWebOutputPlayer{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustOpenFaceChineseOutputJSON: %v", err))
	}
	return string(b)
}

func TestOpenFaceChineseWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockOpenFaceChineseInteractor)
	diMock.On("ResetWithConfig", domain.DefaultOpenFaceChineseConfig()).Return(mockOutput)
	diMock.On("Place", 2).Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.OpenFaceChineseInteractorIF { return diMock }
	ctrl := controller.NewOpenFaceChineseWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.OpenFaceChineseWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustOpenFaceChineseOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("place", func(t *testing.T) {
		input := controller.OpenFaceChineseWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			Row:          func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("place missing row", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustOpenFaceChineseOutputJSON("param error: row is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustOpenFaceChineseOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustOpenFaceChineseOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestOpenFaceChineseWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		players := 4
		rounds := 6
		expected := domain.OpenFaceChineseConfig{
			CpuDifficulty: domain.OpenFaceChineseCpuDifficultyHard,
			PlayerCount:   4,
			TargetRounds:  6,
		}
		diMock := new(usecase.MockOpenFaceChineseInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewOpenFaceChineseWebController(func() uc.OpenFaceChineseInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.OpenFaceChineseWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.OpenFaceChineseWebConfig{CpuDifficulty: &diff, PlayerCount: &players, TargetRounds: &rounds},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultOpenFaceChineseConfig()
		diMock := new(usecase.MockOpenFaceChineseInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewOpenFaceChineseWebController(func() uc.OpenFaceChineseInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.OpenFaceChineseWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.OpenFaceChineseWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultOpenFaceChineseConfig()
		diMock := new(usecase.MockOpenFaceChineseInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewOpenFaceChineseWebController(func() uc.OpenFaceChineseInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.OpenFaceChineseWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestOpenFaceChineseWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockOpenFaceChineseInteractor)
	c := controller.NewOpenFaceChineseWebController(func() uc.OpenFaceChineseInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}
