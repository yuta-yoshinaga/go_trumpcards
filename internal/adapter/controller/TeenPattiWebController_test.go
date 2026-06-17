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

func mustTeenPattiOutputJSON(msg string) string {
	out := &controller.TeenPattiWebOutput{
		Players:           []*controller.TeenPattiWebOutputPlayer{},
		RoundWinnerIdx:    -1,
		MatchWinnerIdx:    -1,
		SideShowRequester: -1,
		SideShowTarget:    -1,
		WebOutputBase:     controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTeenPattiOutputJSON: %v", err))
	}
	return string(b)
}

func TestTeenPattiWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockTeenPattiInteractor)
	giMock.On("ResetWithConfig", domain.DefaultTeenPattiConfig()).Return(mockOutput)
	giMock.On("See").Return(mockOutput)
	giMock.On("Bet").Return(mockOutput)
	giMock.On("Raise", 4).Return(mockOutput)
	giMock.On("Fold").Return(mockOutput)
	giMock.On("Show").Return(mockOutput)
	giMock.On("RequestSideShow").Return(mockOutput)
	giMock.On("RespondSideShow", true).Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TeenPattiInteractorIF { return giMock }
	ctrl := controller.NewTeenPattiWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.TeenPattiWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTeenPattiOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("see", func(t *testing.T) {
		run(t, `{"command":"see","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bet", func(t *testing.T) {
		run(t, `{"command":"bet","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("raise with stake", func(t *testing.T) {
		stake := 4
		input := controller.TeenPattiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "raise", SessionID: "s1"},
			RaiseStake:   &stake,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("raise missing stake", func(t *testing.T) {
		run(t, `{"command":"raise","sessionId":"s1"}`, mustTeenPattiOutputJSON("param error: raiseStake is required."), http.StatusBadRequest)
	})
	t.Run("fold", func(t *testing.T) {
		run(t, `{"command":"fold","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("show", func(t *testing.T) {
		run(t, `{"command":"show","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("sideshow", func(t *testing.T) {
		run(t, `{"command":"sideshow","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("respond with accept", func(t *testing.T) {
		accept := true
		input := controller.TeenPattiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "respond", SessionID: "s1"},
			Accept:       &accept,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("respond missing accept", func(t *testing.T) {
		run(t, `{"command":"respond","sessionId":"s1"}`, mustTeenPattiOutputJSON("param error: accept is required."), http.StatusBadRequest)
	})
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("next alias", func(t *testing.T) {
		run(t, `{"command":"next","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustTeenPattiOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustTeenPattiOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestTeenPattiWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		ante := 3
		chips := 60
		expected := domain.TeenPattiConfig{
			CpuDifficulty: domain.TeenPattiCpuDifficultyHard,
			Ante:          3,
			StartingChips: 60,
		}
		giMock := new(usecase.MockTeenPattiInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTeenPattiWebController(func() uc.TeenPattiInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TeenPattiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.TeenPattiWebConfig{CpuDifficulty: &diff, Ante: &ante, StartingChips: &chips},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultTeenPattiConfig()
		giMock := new(usecase.MockTeenPattiInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTeenPattiWebController(func() uc.TeenPattiInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TeenPattiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.TeenPattiWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTeenPattiConfig()
		giMock := new(usecase.MockTeenPattiInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTeenPattiWebController(func() uc.TeenPattiInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TeenPattiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestTeenPattiWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockTeenPattiInteractor)
	c := controller.NewTeenPattiWebController(func() uc.TeenPattiInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}
