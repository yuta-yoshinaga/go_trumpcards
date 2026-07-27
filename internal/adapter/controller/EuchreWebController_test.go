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

func mustEuchreOutputJSON(msg string) string {
	out := &controller.EuchreWebOutput{
		Players:       []*controller.EuchreWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustEuchreOutputJSON: %v", err))
	}
	return string(b)
}

func TestEuchreWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"bidPlayerIdx":0,"dealerIdx":0,"trumpSuit":0,"faceUpCard":null,"makerTeam":0,"goingAlone":false,"goingAlonePlayerIdx":0,"currentTrick":[],"teamScores":[0,0],"gameEndFlag":false,"winnerTeam":0,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	eiMock := new(usecase.MockEuchreInteractor)
	eiMock.On("ResetWithConfig", domain.DefaultEuchreConfig()).Return(mockOutput)
	eiMock.On("PickUp", true, false).Return(mockOutput)
	eiMock.On("PickUp", true, true).Return(mockOutput)
	eiMock.On("CallTrump", 2, false).Return(mockOutput)
	eiMock.On("CallTrump", 2, true).Return(mockOutput)
	eiMock.On("Pass").Return(mockOutput)
	eiMock.On("Discard", 3).Return(mockOutput)
	eiMock.On("Play", 3).Return(mockOutput)
	eiMock.On("NextTrick").Return(mockOutput)
	eiMock.On("NextRound").Return(mockOutput)
	eiMock.On("Hint").Return(mockOutput)
	eiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.EuchreInteractorIF { return eiMock }
	ctrl := controller.NewEuchreWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec orderup", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "orderup", SessionID: "test-session-1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec o shorthand", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "o", SessionID: "test-session-1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec orderup with goAlone", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "orderup", SessionID: "test-session-1"},
			GoAlone:      func() *bool { v := true; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec calltrump", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "calltrump", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec c calltrump shorthand", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "c", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec calltrump with goAlone", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "calltrump", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
			GoAlone:      func() *bool { v := true; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec pass", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"pass","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec pa shorthand", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"pa","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec discard", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "discard", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec d discard shorthand", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec n next", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"next","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec h hint", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec hint", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("param error."))
	})

	t.Run("failed Exec calltrump without suit field", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"calltrump","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("param error: suit is required."))
	})

	t.Run("failed Exec discard no cardIndex", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"discard","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("param error: cardIndex is required."))
	})

	t.Run("failed Exec play no cardIndex", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustEuchreOutputJSON("param error: cardIndex is required."))
	})
}

func TestEuchreWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 20
		expected := domain.EuchreConfig{CpuDifficulty: domain.EuchreCpuDifficultyHard, PointLimit: 20}
		eiMock := new(usecase.MockEuchreInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.EuchreInteractorIF { return eiMock }
		ctrl := controller.NewEuchreWebController(factory)
		defer ctrl.Stop()

		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.EuchreWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored, uses default", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultEuchreConfig()
		eiMock := new(usecase.MockEuchreInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.EuchreInteractorIF { return eiMock }
		ctrl := controller.NewEuchreWebController(factory)
		defer ctrl.Stop()

		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.EuchreWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored, uses default", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultEuchreConfig()
		eiMock := new(usecase.MockEuchreInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.EuchreInteractorIF { return eiMock }
		ctrl := controller.NewEuchreWebController(factory)
		defer ctrl.Stop()

		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.EuchreWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored, uses default", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultEuchreConfig()
		eiMock := new(usecase.MockEuchreInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.EuchreInteractorIF { return eiMock }
		ctrl := controller.NewEuchreWebController(factory)
		defer ctrl.Stop()

		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.EuchreWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultEuchreConfig()
		eiMock := new(usecase.MockEuchreInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.EuchreInteractorIF { return eiMock }
		ctrl := controller.NewEuchreWebController(factory)
		defer ctrl.Stop()

		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.EuchreWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultEuchreConfig()
		eiMock := new(usecase.MockEuchreInteractor)
		eiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.EuchreInteractorIF { return eiMock }
		ctrl := controller.NewEuchreWebController(factory)
		defer ctrl.Stop()

		input := controller.EuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		eiMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestEuchreWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	mockA := new(usecase.MockEuchreInteractor)
	mockA.On("ResetWithConfig", domain.DefaultEuchreConfig()).Return(mockOutput)
	mockB := new(usecase.MockEuchreInteractor)
	mockB.On("ResetWithConfig", domain.DefaultEuchreConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewEuchreWebController(func() uc.EuchreInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultEuchreConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultEuchreConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultEuchreConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
		var input controller.EuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestEuchreWebController_Stop(t *testing.T) {
	eiMock := new(usecase.MockEuchreInteractor)
	factory := func() uc.EuchreInteractorIF { return eiMock }
	c := controller.NewEuchreWebController(factory)
	c.Stop()
	c.Stop()
}
