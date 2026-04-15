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

func mustPageOneOutputJSON(msg string) string {
	out := &controller.PageOneWebOutput{
		Players:       []*controller.PageOneWebOutputPlayer{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPageOneOutputJSON: %v", err))
	}
	return string(b)
}

func TestPageOneWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"currentPlayerIdx":0,"discardTop":null,"drawPileCount":0,"gameEndFlag":false,"winnerIdx":-1,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`

	siMock := new(usecase.MockPageOneInteractor)
	siMock.On("ResetWithConfig", domain.DefaultPageOneConfig()).Return(mockOutput)
	siMock.On("Play", 3).Return(mockOutput)
	siMock.On("Draw").Return(mockOutput)
	siMock.On("Declare").Return(mockOutput)
	siMock.On("SkipDeclare").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.PageOneInteractorIF { return siMock }
	ctrl := controller.NewPageOneWebController(factory)
	defer ctrl.Stop()

	t.Run("Exec q", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustPageOneOutputJSON("bye."))
	})
	t.Run("Exec reset", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("Exec play", func(t *testing.T) {
		input := controller.PageOneWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("Exec play missing index", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustPageOneOutputJSON("param error: cardIndex is required."))
	})
	t.Run("Exec draw", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("Exec declare", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"declare","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("Exec skip", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"sk","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("Exec nextround", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("Exec log", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("Exec unknown", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustPageOneOutputJSON("Unsupported command."))
	})
	t.Run("Exec empty command", func(t *testing.T) {
		var input controller.PageOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustPageOneOutputJSON("param error."))
	})
}

func TestPageOneWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`
	t.Run("custom config", func(t *testing.T) {
		diff := 2
		limit := 300
		expected := domain.PageOneConfig{CpuDifficulty: domain.PageOneCpuDifficultyHard, PointLimit: 300}
		siMock := new(usecase.MockPageOneInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewPageOneWebController(func() uc.PageOneInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.PageOneWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.PageOneWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("invalid difficulty falls back to default", func(t *testing.T) {
		diff := 99
		expected := domain.DefaultPageOneConfig()
		siMock := new(usecase.MockPageOneInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewPageOneWebController(func() uc.PageOneInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.PageOneWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.PageOneWebConfig{CpuDifficulty: &diff},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}
