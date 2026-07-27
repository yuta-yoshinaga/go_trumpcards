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

func mustNapoleonOutputJSON(msg string) string {
	out := &controller.NapoleonWebOutput{
		Players:       []*controller.NapoleonWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    domain.NapoleonWinnerUndecided,
		NapoleonIdx:   -1,
		AdjutantIdx:   -1,
		HighestBidder: -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustNapoleonOutputJSON: %v", err))
	}
	return string(b)
}

func TestNapoleonWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"bidPlayerIdx":0,"currentTrick":[],"trumpSuit":0,"napoleonIdx":-1,"adjutantIdx":-1,"adjutantRevealed":false,"highestBid":0,"highestBidder":-1,"gameEndFlag":false,"winnerTeam":-1,"leadPlayerIdx":0,"message":"","config":{"cpuDifficulty":0,"minBid":0,"pointLimit":0}}`
	expectedBody := mockOutput

	niMock := new(usecase.MockNapoleonInteractor)
	niMock.On("ResetWithConfig", domain.DefaultNapoleonConfig()).Return(mockOutput)
	niMock.On("Bid", 3).Return(mockOutput)
	niMock.On("DeclareTrump", 1, 2, 13).Return(mockOutput)
	niMock.On("ExchangeKitty", 3).Return(mockOutput)
	niMock.On("Play", 3).Return(mockOutput)
	niMock.On("NextTrick").Return(mockOutput)
	niMock.On("NextRound").Return(mockOutput)
	niMock.On("Hint").Return(mockOutput)
	niMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.NapoleonInteractorIF { return niMock }
	ctrl := controller.NewNapoleonWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec bid", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "test-session-1"},
			Bid:          func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec b bid shorthand", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "test-session-1"},
			Bid:          func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec trump", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput:  controller.BaseWebInput{Command: "trump", SessionID: "test-session-1"},
			TrumpSuit:     func() *int { v := 1; return &v }(),
			AdjutantSuit:  func() *int { v := 2; return &v }(),
			AdjutantValue: func() *int { v := 13; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec t trump shorthand", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput:  controller.BaseWebInput{Command: "t", SessionID: "test-session-1"},
			TrumpSuit:     func() *int { v := 1; return &v }(),
			AdjutantSuit:  func() *int { v := 2; return &v }(),
			AdjutantValue: func() *int { v := 13; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec exchange", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "exchange", SessionID: "test-session-1"},
			DiscardIndex: func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec e exchange shorthand", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "e", SessionID: "test-session-1"},
			DiscardIndex: func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec n next", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"next","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec h hint", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec hint", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("param error."))
	})

	t.Run("failed Exec bid without bid field", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"bid","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("param error: bid is required."))
	})

	t.Run("failed Exec trump without trumpSuit", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"trump","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("param error: trumpSuit, adjutantSuit, adjutantValue are required."))
	})

	t.Run("failed Exec exchange without discardIndex", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"exchange","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("param error: discardIndex is required."))
	})

	t.Run("failed Exec play no cardIndex", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNapoleonOutputJSON("param error: cardIndex is required."))
	})
}

func TestNapoleonWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		minBid := 15
		limit := 200
		expected := domain.NapoleonConfig{CpuDifficulty: domain.NapoleonCpuDifficultyHard, MinBid: 15, PointLimit: 200}
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.NapoleonWebConfig{CpuDifficulty: &diff, MinBid: &minBid, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored, uses default", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultNapoleonConfig()
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.NapoleonWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored, uses default", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultNapoleonConfig()
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.NapoleonWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("minBid below 1 is ignored, uses default", func(t *testing.T) {
		minBid := 0
		expected := domain.DefaultNapoleonConfig()
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.NapoleonWebConfig{MinBid: &minBid},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("minBid exceeding max is ignored", func(t *testing.T) {
		minBid := 18
		expected := domain.DefaultNapoleonConfig()
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-minbid-max"},
			Config:       &controller.NapoleonWebConfig{MinBid: &minBid},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored, uses default", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultNapoleonConfig()
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
			Config:       &controller.NapoleonWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultNapoleonConfig()
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.NapoleonWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultNapoleonConfig()
		niMock := new(usecase.MockNapoleonInteractor)
		niMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.NapoleonInteractorIF { return niMock }
		ctrl := controller.NewNapoleonWebController(factory)
		defer ctrl.Stop()

		input := controller.NapoleonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-6"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		niMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestNapoleonWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	mockA := new(usecase.MockNapoleonInteractor)
	mockA.On("ResetWithConfig", domain.DefaultNapoleonConfig()).Return(mockOutput)
	mockB := new(usecase.MockNapoleonInteractor)
	mockB.On("ResetWithConfig", domain.DefaultNapoleonConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewNapoleonWebController(func() uc.NapoleonInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultNapoleonConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultNapoleonConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultNapoleonConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
		var input controller.NapoleonWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestNapoleonWebController_Stop(t *testing.T) {
	niMock := new(usecase.MockNapoleonInteractor)
	factory := func() uc.NapoleonInteractorIF { return niMock }
	c := controller.NewNapoleonWebController(factory)
	c.Stop()
	c.Stop()
}
