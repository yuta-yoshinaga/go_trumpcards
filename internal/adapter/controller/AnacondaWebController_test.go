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

func mustAnacondaOutputJSON(msg string) string {
	out := &controller.AnacondaWebOutput{
		Players:        make([]*controller.AnacondaWebOutputPlayer, 0),
		WinnerIdx:      -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustAnacondaOutputJSON: %v", err))
	}
	return string(b)
}

func TestAnacondaWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockAnacondaInteractor)
	biMock.On("ResetWithConfig", domain.DefaultAnacondaConfig()).Return(mockOutput)
	biMock.On("Pass", []int{0, 1, 2}).Return(mockOutput)
	biMock.On("Keep", []int{0, 1, 2, 3, 4}).Return(mockOutput)
	biMock.On("Call").Return(mockOutput)
	biMock.On("Raise").Return(mockOutput)
	biMock.On("Fold").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.AnacondaInteractorIF { return biMock }
	ctrl := controller.NewAnacondaWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.AnacondaWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustAnacondaOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("pass", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1","cardIndices":[0,1,2]}`, mockOutput, http.StatusOK)
		biMock.AssertCalled(t, "Pass", []int{0, 1, 2})
	})
	t.Run("pass missing", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1"}`, mustAnacondaOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("keep", func(t *testing.T) {
		run(t, `{"command":"keep","sessionId":"s1","cardIndices":[0,1,2,3,4]}`, mockOutput, http.StatusOK)
		biMock.AssertCalled(t, "Keep", []int{0, 1, 2, 3, 4})
	})
	t.Run("keep missing", func(t *testing.T) {
		run(t, `{"command":"keep","sessionId":"s1"}`, mustAnacondaOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("bet call", func(t *testing.T) {
		action := "call"
		input := controller.AnacondaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Action:       &action,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Call")
	})
	t.Run("bet raise", func(t *testing.T) {
		action := "raise"
		input := controller.AnacondaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Action:       &action,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "Raise")
	})
	t.Run("bet fold", func(t *testing.T) {
		action := "fold"
		input := controller.AnacondaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Action:       &action,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "Fold")
	})
	t.Run("bet missing action", func(t *testing.T) {
		run(t, `{"command":"bet","sessionId":"s1"}`, mustAnacondaOutputJSON("param error: action is required."), http.StatusBadRequest)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustAnacondaOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
}

func TestAnacondaWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("custom config passed through", func(t *testing.T) {
		players, ante, chips, rounds := 6, 25, 500, 20
		expected := domain.AnacondaConfig{PlayerCount: 6, Ante: 25, StartingChips: 500, TargetRounds: 20}
		biMock := new(usecase.MockAnacondaInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewAnacondaWebController(func() uc.AnacondaInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.AnacondaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config: &controller.AnacondaWebConfig{
				PlayerCount: &players, Ante: &ante, StartingChips: &chips, TargetRounds: &rounds,
			},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range values fall back to default", func(t *testing.T) {
		players := 99
		expected := domain.DefaultAnacondaConfig()
		biMock := new(usecase.MockAnacondaInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewAnacondaWebController(func() uc.AnacondaInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.AnacondaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.AnacondaWebConfig{PlayerCount: &players},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestAnacondaWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockAnacondaInteractor)
	c := controller.NewAnacondaWebController(func() uc.AnacondaInteractorIF { return biMock })
	c.Stop()
	c.Stop()
}
