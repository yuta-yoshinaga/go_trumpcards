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

func mustThreeCardBragOutputJSON(msg string) string {
	out := &controller.ThreeCardBragWebOutput{
		Players:        []*controller.ThreeCardBragWebOutputPlayer{},
		RoundWinnerIdx: -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustThreeCardBragOutputJSON: %v", err))
	}
	return string(b)
}

func TestThreeCardBragWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockThreeCardBragInteractor)
	giMock.On("ResetWithConfig", domain.DefaultThreeCardBragConfig()).Return(mockOutput)
	giMock.On("See").Return(mockOutput)
	giMock.On("Bet").Return(mockOutput)
	giMock.On("Raise", 4).Return(mockOutput)
	giMock.On("Fold").Return(mockOutput)
	giMock.On("Show").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ThreeCardBragInteractorIF { return giMock }
	ctrl := controller.NewThreeCardBragWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.ThreeCardBragWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustThreeCardBragOutputJSON("bye."), http.StatusOK)
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
		input := controller.ThreeCardBragWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "raise", SessionID: "s1"},
			RaiseStake:   &stake,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("raise missing stake", func(t *testing.T) {
		run(t, `{"command":"raise","sessionId":"s1"}`, mustThreeCardBragOutputJSON("param error: raiseStake is required."), http.StatusBadRequest)
	})
	t.Run("fold", func(t *testing.T) {
		run(t, `{"command":"fold","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("show", func(t *testing.T) {
		run(t, `{"command":"show","sessionId":"s1"}`, mockOutput, http.StatusOK)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustThreeCardBragOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustThreeCardBragOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestThreeCardBragWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		ante := 3
		chips := 60
		expected := domain.ThreeCardBragConfig{
			CpuDifficulty: domain.ThreeCardBragCpuDifficultyHard,
			Ante:          3,
			StartingChips: 60,
		}
		giMock := new(usecase.MockThreeCardBragInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewThreeCardBragWebController(func() uc.ThreeCardBragInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.ThreeCardBragWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.ThreeCardBragWebConfig{CpuDifficulty: &diff, Ante: &ante, StartingChips: &chips},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultThreeCardBragConfig()
		giMock := new(usecase.MockThreeCardBragInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewThreeCardBragWebController(func() uc.ThreeCardBragInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.ThreeCardBragWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.ThreeCardBragWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultThreeCardBragConfig()
		giMock := new(usecase.MockThreeCardBragInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewThreeCardBragWebController(func() uc.ThreeCardBragInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.ThreeCardBragWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestThreeCardBragWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockThreeCardBragInteractor)
	c := controller.NewThreeCardBragWebController(func() uc.ThreeCardBragInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}
