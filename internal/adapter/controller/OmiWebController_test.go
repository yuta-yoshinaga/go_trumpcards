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

func mustOmiOutputJSON(msg string) string {
	out := &controller.OmiWebOutput{
		Players:       []*controller.OmiWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustOmiOutputJSON: %v", err))
	}
	return string(b)
}

func TestOmiWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"bidPlayerIdx":0,"dealerIdx":0,"trumpSuit":0,"faceUpCard":null,"makerTeam":0,"goingAlone":false,"goingAlonePlayerIdx":0,"currentTrick":[],"teamScores":[0,0],"gameEndFlag":false,"winnerTeam":0,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	eiMock := new(usecase.MockOmiInteractor)
	eiMock.On("ResetWithConfig", domain.DefaultOmiConfig()).Return(mockOutput)
	eiMock.On("CallTrump", 2).Return(mockOutput)
	eiMock.On("Play", 3).Return(mockOutput)
	eiMock.On("NextTrick").Return(mockOutput)
	eiMock.On("NextRound").Return(mockOutput)
	eiMock.On("Hint").Return(mockOutput)
	eiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.OmiInteractorIF { return eiMock }
	ctrl := controller.NewOmiWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// removed commands are rejected as unsupported
	t.Run("removed commands rejected as unsupported", func(t *testing.T) {
		removedCmds := []string{"orderup", "o", "pass", "pa", "discard", "d"}
		for _, cmd := range removedCmds {
			input := controller.OmiWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "test-session-1"},
			}
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(mustOmiOutputJSON("Unsupported command."))
		}
	})

	t.Run("success Exec calltrump", func(t *testing.T) {
		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "calltrump", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
		eiMock.AssertCalled(t, "CallTrump", 2)
	})

	t.Run("success Exec c calltrump shorthand", func(t *testing.T) {
		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "c", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec trump", func(t *testing.T) {
		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "trump", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec t shorthand", func(t *testing.T) {
		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "t", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play", func(t *testing.T) {
		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec n next", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"next","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec h hint", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec hint", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("param error."))
	})

	t.Run("failed Exec calltrump without suit field", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"calltrump","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("param error: suit is required."))
	})

	t.Run("failed Exec play no cardIndex", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustOmiOutputJSON("param error: cardIndex is required."))
	})
}

func TestOmiWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 20
		expected := domain.OmiConfig{CpuDifficulty: domain.OmiCpuDifficultyHard, PointLimit: 20}
		eiMock := new(usecase.MockOmiInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.OmiInteractorIF { return eiMock }
		ctrl := controller.NewOmiWebController(factory)
		defer ctrl.Stop()

		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.OmiWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored, uses default", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultOmiConfig()
		eiMock := new(usecase.MockOmiInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.OmiInteractorIF { return eiMock }
		ctrl := controller.NewOmiWebController(factory)
		defer ctrl.Stop()

		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.OmiWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored, uses default", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultOmiConfig()
		eiMock := new(usecase.MockOmiInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.OmiInteractorIF { return eiMock }
		ctrl := controller.NewOmiWebController(factory)
		defer ctrl.Stop()

		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.OmiWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored, uses default", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultOmiConfig()
		eiMock := new(usecase.MockOmiInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.OmiInteractorIF { return eiMock }
		ctrl := controller.NewOmiWebController(factory)
		defer ctrl.Stop()

		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.OmiWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultOmiConfig()
		eiMock := new(usecase.MockOmiInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.OmiInteractorIF { return eiMock }
		ctrl := controller.NewOmiWebController(factory)
		defer ctrl.Stop()

		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.OmiWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultOmiConfig()
		eiMock := new(usecase.MockOmiInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.OmiInteractorIF { return eiMock }
		ctrl := controller.NewOmiWebController(factory)
		defer ctrl.Stop()

		input := controller.OmiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestOmiWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	mockA := new(usecase.MockOmiInteractor)
	mockA.On("ResetWithConfig", domain.DefaultOmiConfig()).Return(mockOutput)
	mockB := new(usecase.MockOmiInteractor)
	mockB.On("ResetWithConfig", domain.DefaultOmiConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewOmiWebController(func() uc.OmiInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultOmiConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultOmiConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultOmiConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
		var input controller.OmiWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestOmiWebController_Stop(t *testing.T) {
	eiMock := new(usecase.MockOmiInteractor)
	factory := func() uc.OmiInteractorIF { return eiMock }
	c := controller.NewOmiWebController(factory)
	c.Stop()
	c.Stop()
}
