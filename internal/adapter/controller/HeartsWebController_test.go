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

func mustHeartsOutputJSON(msg string) string {
	out := &controller.HeartsWebOutput{
		Players:       []*controller.HeartsWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHeartsOutputJSON: %v", err))
	}
	return string(b)
}

func TestHeartsWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"currentTrick":[],"heartsBroken":false,"passDirection":0,"gameEndFlag":false,"winnerIdx":-1,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	hiMock := new(usecase.MockHeartsInteractor)
	hiMock.On("ResetWithConfig", domain.DefaultHeartsConfig()).Return(mockOutput)
	hiMock.On("Pass", []int{0, 1, 2}).Return(mockOutput)
	hiMock.On("Play", 3).Return(mockOutput)
	hiMock.On("NextTrick").Return(mockOutput)
	hiMock.On("NextRound").Return(mockOutput)
	hiMock.On("Hint").Return(mockOutput)
	hiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.HeartsInteractorIF { return hiMock }
	ctrl := controller.NewHeartsWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec pass", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"pass","cardIndices":[0,1,2],"sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play", func(t *testing.T) {
		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec n next", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"next","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec h hint", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec hint", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("param error."))
	})

	t.Run("failed Exec pass wrong count", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"pass","cardIndices":[0,1],"sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("param error: pass requires exactly 3 card indices."))
	})

	t.Run("failed Exec play no cardIndex", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustHeartsOutputJSON("param error: cardIndex is required."))
	})
}

func TestHeartsWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom cpuDifficulty and pointLimit are passed", func(t *testing.T) {
		diff := 2
		limit := 50
		expected := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficultyHard, PointLimit: 50}
		hiMock := new(usecase.MockHeartsInteractor)
		hiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HeartsInteractorIF { return hiMock }
		ctrl := controller.NewHeartsWebController(factory)
		defer ctrl.Stop()

		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.HeartsWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		hiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max (3) is ignored, uses default", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultHeartsConfig()
		hiMock := new(usecase.MockHeartsInteractor)
		hiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HeartsInteractorIF { return hiMock }
		ctrl := controller.NewHeartsWebController(factory)
		defer ctrl.Stop()

		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.HeartsWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		hiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min (-1) is ignored, uses default", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultHeartsConfig()
		hiMock := new(usecase.MockHeartsInteractor)
		hiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HeartsInteractorIF { return hiMock }
		ctrl := controller.NewHeartsWebController(factory)
		defer ctrl.Stop()

		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.HeartsWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		hiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored, uses default", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultHeartsConfig()
		hiMock := new(usecase.MockHeartsInteractor)
		hiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HeartsInteractorIF { return hiMock }
		ctrl := controller.NewHeartsWebController(factory)
		defer ctrl.Stop()

		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.HeartsWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		hiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("point limit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultHeartsConfig()
		hiMock := new(usecase.MockHeartsInteractor)
		hiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HeartsInteractorIF { return hiMock }
		ctrl := controller.NewHeartsWebController(factory)
		defer ctrl.Stop()

		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.HeartsWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		hiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultHeartsConfig()
		hiMock := new(usecase.MockHeartsInteractor)
		hiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HeartsInteractorIF { return hiMock }
		ctrl := controller.NewHeartsWebController(factory)
		defer ctrl.Stop()

		input := controller.HeartsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		hiMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestHeartsWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	mockA := new(usecase.MockHeartsInteractor)
	mockA.On("ResetWithConfig", domain.DefaultHeartsConfig()).Return(mockOutput)
	mockB := new(usecase.MockHeartsInteractor)
	mockB.On("ResetWithConfig", domain.DefaultHeartsConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewHeartsWebController(func() uc.HeartsInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultHeartsConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultHeartsConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultHeartsConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
		var input controller.HeartsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestHeartsWebController_Stop(t *testing.T) {
	hiMock := new(usecase.MockHeartsInteractor)
	factory := func() uc.HeartsInteractorIF { return hiMock }
	c := controller.NewHeartsWebController(factory)
	c.Stop()
	c.Stop()
}
