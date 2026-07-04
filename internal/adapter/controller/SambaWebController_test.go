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

func mustSambaOutputJSON(msg string) string {
	out := &controller.SambaWebOutput{
		Players:       make([]*controller.SambaWebOutputPlayer, 0),
		TeamScores:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSambaOutputJSON: %v", err))
	}
	return string(b)
}

func TestSambaWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	expectedBody := mockOutput

	siMock := new(usecase.MockSambaInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSambaConfig()).Return(mockOutput)
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

	factory := func() uc.SambaInteractorIF { return siMock }
	ctrl := controller.NewSambaWebController(factory)
	defer ctrl.Stop()

	run := func(name, body string, input *controller.SambaWebInput, expect string, code int) {
		t.Run(name, func(t *testing.T) {
			in := input
			if in == nil {
				in = new(controller.SambaWebInput)
				_ = json.Unmarshal([]byte(body), in)
			}
			recorded := execRequest(t, ctrl.Exec, in)
			recorded.CodeIs(code)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(expect)
		})
	}

	run("quit", `{"command":"q","sessionId":"s1"}`, nil, mustSambaOutputJSON("bye."), http.StatusOK)
	run("reset", `{"command":"reset","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)
	run("drawstock", `{"command":"ds","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)
	run("drawdiscard no idx", `{"command":"dd","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)
	run("meld no groups", `{"command":"m","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)
	run("skipmeld", `{"command":"sm","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)
	run("goout", `{"command":"go","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)
	run("nextround", `{"command":"nr","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)
	run("log", `{"command":"l","sessionId":"s1"}`, nil, expectedBody, http.StatusOK)

	t.Run("drawdiscard with indices", func(t *testing.T) {
		input := controller.SambaWebInput{
			BaseWebInput:       controller.BaseWebInput{Command: "drawdiscard", SessionID: "s1"},
			NaturalPairIndices: []int{0, 1},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("meld with groups", func(t *testing.T) {
		input := controller.SambaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "meld", SessionID: "s1"},
			MeldGroups:   [][]int{{0, 1, 2}},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("discard with index", func(t *testing.T) {
		v := 3
		input := controller.SambaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndex:    &v,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	run("unknown command", `{"command":"other","sessionId":"s1"}`, nil, mustSambaOutputJSON("Unsupported command."), http.StatusBadRequest)
	run("empty command", `{"command":"","sessionId":"s1"}`, nil, mustSambaOutputJSON("param error."), http.StatusBadRequest)
	run("discard no cardIndex", `{"command":"d","sessionId":"s1"}`, nil, mustSambaOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
}

func TestSambaWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 7500
		expected := domain.SambaConfig{CpuDifficulty: domain.SambaCpuDifficultyHard, PointLimit: 7500}
		siMock := new(usecase.MockSambaInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		ctrl := controller.NewSambaWebController(func() uc.SambaInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.SambaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.SambaWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range difficulty ignored", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultSambaConfig()
		siMock := new(usecase.MockSambaInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		ctrl := controller.NewSambaWebController(func() uc.SambaInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.SambaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.SambaWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultSambaConfig()
		siMock := new(usecase.MockSambaInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		ctrl := controller.NewSambaWebController(func() uc.SambaInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.SambaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestSambaWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockSambaInteractor)
	c := controller.NewSambaWebController(func() uc.SambaInteractorIF { return siMock })
	c.Stop()
	c.Stop()
}
