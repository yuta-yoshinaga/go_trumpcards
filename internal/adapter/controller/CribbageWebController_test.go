//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustCribbageOutputJSON(msg string) string {
	out := &controller.CribbageWebOutput{
		Players:        []*controller.CribbageWebOutputPlayer{},
		Crib:           []*controller.WebOutputCard{},
		PegPlayedCards: []*controller.WebOutputCard{},
		WinnerIdx:      -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCribbageOutputJSON: %v", err))
	}
	return string(b)
}

func TestCribbageWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"currentPlayerIdx":0,"dealerIdx":0,"crib":[],"starter":null,"pegCount":0,"pegPlayedCards":[],"showPhaseStep":0,"handScoreDetails":[null,null,null],"gameEndFlag":false,"winnerIdx":-1,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	siMock := new(usecase.MockCribbageInteractor)
	siMock.On("ResetWithConfig", domain.DefaultCribbageConfig()).Return(mockOutput)
	siMock.On("Discard", []int{1, 3}).Return(mockOutput)
	siMock.On("Cut").Return(mockOutput)
	siMock.On("Peg", 3).Return(mockOutput)
	siMock.On("Go").Return(mockOutput)
	siMock.On("ShowNext").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.CribbageInteractorIF { return siMock }
	ctrl := controller.NewCribbageWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCribbageOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCribbageOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec discard", func(t *testing.T) {
		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "test-session-1"},
			CardIndices:  []int{1, 3},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec discard long", func(t *testing.T) {
		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "discard", SessionID: "test-session-1"},
			CardIndices:  []int{1, 3},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec c cut", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"c","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec cut long", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"cut","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec peg", func(t *testing.T) {
		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec peg long", func(t *testing.T) {
		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "peg", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec go", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"go","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec sn shownext", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"sn","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec shownext", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"shownext","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCribbageOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCribbageOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCribbageOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCribbageOutputJSON("param error."))
	})

	t.Run("failed Exec peg no cardIndex", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCribbageOutputJSON("param error: cardIndex is required."))
	})
}

func TestCribbageWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 200
		expected := domain.CribbageConfig{CpuDifficulty: domain.CribbageCpuDifficultyHard, PointLimit: 200}
		siMock := new(usecase.MockCribbageInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CribbageInteractorIF { return siMock }
		ctrl := controller.NewCribbageWebController(factory)
		defer ctrl.Stop()

		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.CribbageWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultCribbageConfig()
		siMock := new(usecase.MockCribbageInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CribbageInteractorIF { return siMock }
		ctrl := controller.NewCribbageWebController(factory)
		defer ctrl.Stop()

		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.CribbageWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultCribbageConfig()
		siMock := new(usecase.MockCribbageInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CribbageInteractorIF { return siMock }
		ctrl := controller.NewCribbageWebController(factory)
		defer ctrl.Stop()

		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.CribbageWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultCribbageConfig()
		siMock := new(usecase.MockCribbageInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CribbageInteractorIF { return siMock }
		ctrl := controller.NewCribbageWebController(factory)
		defer ctrl.Stop()

		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.CribbageWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultCribbageConfig()
		siMock := new(usecase.MockCribbageInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CribbageInteractorIF { return siMock }
		ctrl := controller.NewCribbageWebController(factory)
		defer ctrl.Stop()

		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.CribbageWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultCribbageConfig()
		siMock := new(usecase.MockCribbageInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CribbageInteractorIF { return siMock }
		ctrl := controller.NewCribbageWebController(factory)
		defer ctrl.Stop()

		input := controller.CribbageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestCribbageWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[]}`

	mockA := new(usecase.MockCribbageInteractor)
	mockA.On("ResetWithConfig", domain.DefaultCribbageConfig()).Return(mockOutput)
	mockB := new(usecase.MockCribbageInteractor)
	mockB.On("ResetWithConfig", domain.DefaultCribbageConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewCribbageWebController(func() uc.CribbageInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultCribbageConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultCribbageConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultCribbageConfig())
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.CribbageWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestCribbageWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockCribbageInteractor)
	factory := func() uc.CribbageInteractorIF { return siMock }
	c := controller.NewCribbageWebController(factory)
	c.Stop()
	c.Stop()
}
