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

func mustHandAndFootOutputJSON(msg string) string {
	out := &controller.HandAndFootWebOutput{
		Players:       make([]*controller.HandAndFootWebOutputPlayer, 0),
		Teams:         make([]*controller.HandAndFootWebOutputTeam, 0),
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHandAndFootOutputJSON: %v", err))
	}
	return string(b)
}

func TestHandAndFootWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	expectedBody := mockOutput

	siMock := new(usecase.MockHandAndFootInteractor)
	siMock.On("ResetWithConfig", domain.DefaultHandAndFootConfig()).Return(mockOutput)
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

	factory := func() uc.HandAndFootInteractorIF { return siMock }
	ctrl := controller.NewHandAndFootWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustHandAndFootOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec ds", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"ds","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec dd with indices", func(t *testing.T) {
		input := controller.HandAndFootWebInput{
			BaseWebInput:       controller.BaseWebInput{Command: "dd", SessionID: "s1"},
			NaturalPairIndices: []int{0, 1},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec m with groups", func(t *testing.T) {
		input := controller.HandAndFootWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			MeldGroups:   [][]int{{0, 1, 2}},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec sm", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"sm","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec d discard", func(t *testing.T) {
		input := controller.HandAndFootWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec go", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"go","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustHandAndFootOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustHandAndFootOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.HandAndFootWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustHandAndFootOutputJSON("param error."))
	})

	t.Run("failed Exec discard no cardIndex", func(t *testing.T) {
		var input controller.HandAndFootWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustHandAndFootOutputJSON("param error: cardIndex is required."))
	})
}

func TestHandAndFootWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 7500
		expected := domain.DefaultHandAndFootConfig()
		expected.CpuDifficulty = domain.HandAndFootCpuDifficultyHard
		expected.PointLimit = 7500
		siMock := new(usecase.MockHandAndFootInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HandAndFootInteractorIF { return siMock }
		ctrl := controller.NewHandAndFootWebController(factory)
		defer ctrl.Stop()

		input := controller.HandAndFootWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.HandAndFootWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max ignored", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultHandAndFootConfig()
		siMock := new(usecase.MockHandAndFootInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HandAndFootInteractorIF { return siMock }
		ctrl := controller.NewHandAndFootWebController(factory)
		defer ctrl.Stop()

		input := controller.HandAndFootWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.HandAndFootWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultHandAndFootConfig()
		siMock := new(usecase.MockHandAndFootInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.HandAndFootInteractorIF { return siMock }
		ctrl := controller.NewHandAndFootWebController(factory)
		defer ctrl.Stop()

		input := controller.HandAndFootWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestHandAndFootWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockHandAndFootInteractor)
	factory := func() uc.HandAndFootInteractorIF { return siMock }
	c := controller.NewHandAndFootWebController(factory)
	c.Stop()
	c.Stop()
}
