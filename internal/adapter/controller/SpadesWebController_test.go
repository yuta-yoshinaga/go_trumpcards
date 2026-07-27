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

func mustSpadesOutputJSON(msg string) string {
	out := &controller.SpadesWebOutput{
		Players:       []*controller.SpadesWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSpadesOutputJSON: %v", err))
	}
	return string(b)
}

func TestSpadesWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"bidPlayerIdx":0,"currentTrick":[],"spadesBroken":false,"gameEndFlag":false,"winnerIdx":-1,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"pointLimit":0,"nilBonus":0,"bagPenaltyThreshold":0}}`
	expectedBody := mockOutput

	siMock := new(usecase.MockSpadesInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSpadesConfig()).Return(mockOutput)
	siMock.On("Bid", 3).Return(mockOutput)
	siMock.On("Play", 3).Return(mockOutput)
	siMock.On("NextTrick").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.SpadesInteractorIF { return siMock }
	ctrl := controller.NewSpadesWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec bid", func(t *testing.T) {
		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "test-session-1"},
			Bid:          func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec b bid shorthand", func(t *testing.T) {
		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "test-session-1"},
			Bid:          func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play", func(t *testing.T) {
		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec n next", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"next","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec h hint", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec hint", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("param error."))
	})

	t.Run("failed Exec bid without bid field", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"bid","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("param error: bid is required."))
	})

	t.Run("failed Exec play no cardIndex", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustSpadesOutputJSON("param error: cardIndex is required."))
	})
}

func TestSpadesWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 300
		nilBonus := 200
		bagThreshold := 5
		expected := domain.SpadesConfig{CpuDifficulty: domain.SpadesCpuDifficultyHard, PointLimit: 300, NilBonus: 200, BagPenaltyThreshold: 5}
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.SpadesWebConfig{CpuDifficulty: &diff, PointLimit: &limit, NilBonus: &nilBonus, BagPenaltyThreshold: &bagThreshold},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored, uses default", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.SpadesWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored, uses default", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.SpadesWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored, uses default", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.SpadesWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.SpadesWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nilBonus below 0 is ignored, uses default", func(t *testing.T) {
		nilBonus := -1
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-nilbonus"},
			Config:       &controller.SpadesWebConfig{NilBonus: &nilBonus},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nilBonus exceeding 500 is ignored", func(t *testing.T) {
		nilBonus := 501
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-nilbonus-max"},
			Config:       &controller.SpadesWebConfig{NilBonus: &nilBonus},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("bagPenaltyThreshold below 1 is ignored, uses default", func(t *testing.T) {
		bagThreshold := 0
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-bagthresh"},
			Config:       &controller.SpadesWebConfig{BagPenaltyThreshold: &bagThreshold},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("bagPenaltyThreshold exceeding 100 is ignored", func(t *testing.T) {
		bagThreshold := 101
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-bagthresh-max"},
			Config:       &controller.SpadesWebConfig{BagPenaltyThreshold: &bagThreshold},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultSpadesConfig()
		siMock := new(usecase.MockSpadesInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SpadesInteractorIF { return siMock }
		ctrl := controller.NewSpadesWebController(factory)
		defer ctrl.Stop()

		input := controller.SpadesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestSpadesWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	mockA := new(usecase.MockSpadesInteractor)
	mockA.On("ResetWithConfig", domain.DefaultSpadesConfig()).Return(mockOutput)
	mockB := new(usecase.MockSpadesInteractor)
	mockB.On("ResetWithConfig", domain.DefaultSpadesConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewSpadesWebController(func() uc.SpadesInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultSpadesConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultSpadesConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultSpadesConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
		var input controller.SpadesWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestSpadesWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockSpadesInteractor)
	factory := func() uc.SpadesInteractorIF { return siMock }
	c := controller.NewSpadesWebController(factory)
	c.Stop()
	c.Stop()
}
