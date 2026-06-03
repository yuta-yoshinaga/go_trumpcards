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

func mustBurracoOutputJSON(msg string) string {
	out := &controller.BurracoWebOutput{
		Players:       make([]*controller.BurracoWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBurracoOutputJSON: %v", err))
	}
	return string(b)
}

func TestBurracoWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"currentPlayerIdx":0,"discardTop":null,"drawPileCount":0,"discardPileCount":0,"isFrozen":false,"gameEndFlag":false,"winnerIdx":-1,"message":"","messageCode":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	siMock := new(usecase.MockBurracoInteractor)
	siMock.On("ResetWithConfig", domain.DefaultBurracoConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard", ([]int)(nil)).Return(mockOutput)
	siMock.On("DrawFromDiscard", []int{0, 1}).Return(mockOutput)
	siMock.On("Meld", ([][]int)(nil)).Return(mockOutput)
	siMock.On("Meld", [][]int{{0, 1, 2}}).Return(mockOutput)
	siMock.On("SkipMeld").Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("GoOut").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BurracoInteractorIF { return siMock }
	ctrl := controller.NewBurracoWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBurracoOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBurracoOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec ds drawstock", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"ds","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec drawstock", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"drawstock","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec dd drawdiscard no indices", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"dd","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec drawdiscard with indices", func(t *testing.T) {
		input := controller.BurracoWebInput{
			BaseWebInput:       controller.BaseWebInput{Command: "drawdiscard", SessionID: "test-session-1"},
			NaturalPairIndices: []int{0, 1},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec m meld", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec meld with groups", func(t *testing.T) {
		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "meld", SessionID: "test-session-1"},
			MeldGroups:   [][]int{{0, 1, 2}},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec sm skipmeld", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"sm","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec skipmeld", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"skipmeld","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec d discard", func(t *testing.T) {
		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec discard long", func(t *testing.T) {
		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "discard", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec go goout", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"go","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec goout", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"goout","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBurracoOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBurracoOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBurracoOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBurracoOutputJSON("param error."))
	})

	t.Run("failed Exec discard no cardIndex", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBurracoOutputJSON("param error: cardIndex is required."))
	})
}

func TestBurracoWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 7500
		expected := domain.BurracoConfig{CpuDifficulty: domain.BurracoCpuDifficultyHard, PointLimit: 7500}
		siMock := new(usecase.MockBurracoInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BurracoInteractorIF { return siMock }
		ctrl := controller.NewBurracoWebController(factory)
		defer ctrl.Stop()

		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.BurracoWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultBurracoConfig()
		siMock := new(usecase.MockBurracoInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BurracoInteractorIF { return siMock }
		ctrl := controller.NewBurracoWebController(factory)
		defer ctrl.Stop()

		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.BurracoWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultBurracoConfig()
		siMock := new(usecase.MockBurracoInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BurracoInteractorIF { return siMock }
		ctrl := controller.NewBurracoWebController(factory)
		defer ctrl.Stop()

		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.BurracoWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultBurracoConfig()
		siMock := new(usecase.MockBurracoInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BurracoInteractorIF { return siMock }
		ctrl := controller.NewBurracoWebController(factory)
		defer ctrl.Stop()

		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.BurracoWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultBurracoConfig()
		siMock := new(usecase.MockBurracoInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BurracoInteractorIF { return siMock }
		ctrl := controller.NewBurracoWebController(factory)
		defer ctrl.Stop()

		input := controller.BurracoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestBurracoWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[]}`

	mockA := new(usecase.MockBurracoInteractor)
	mockA.On("ResetWithConfig", domain.DefaultBurracoConfig()).Return(mockOutput)
	mockB := new(usecase.MockBurracoInteractor)
	mockB.On("ResetWithConfig", domain.DefaultBurracoConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewBurracoWebController(func() uc.BurracoInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultBurracoConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultBurracoConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultBurracoConfig())
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.BurracoWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestBurracoWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockBurracoInteractor)
	factory := func() uc.BurracoInteractorIF { return siMock }
	c := controller.NewBurracoWebController(factory)
	c.Stop()
	c.Stop()
}
