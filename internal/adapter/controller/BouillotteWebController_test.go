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

func mustBouillotteOutputJSON(msg string) string {
	out := &controller.BouillotteWebOutput{
		Players:        make([]*controller.BouillotteWebOutputPlayer, 0),
		WinnerIdx:      -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBouillotteOutputJSON: %v", err))
	}
	return string(b)
}

func TestBouillotteWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockBouillotteInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBouillotteConfig()).Return(mockOutput)
	biMock.On("Bet", "call").Return(mockOutput)
	biMock.On("Bet", "raise").Return(mockOutput)
	biMock.On("Bet", "fold").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BouillotteInteractorIF { return biMock }
	ctrl := controller.NewBouillotteWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.BouillotteWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustBouillotteOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bet call", func(t *testing.T) {
		action := "call"
		input := controller.BouillotteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Action:       &action,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Bet", "call")
	})
	t.Run("bet raise", func(t *testing.T) {
		action := "raise"
		input := controller.BouillotteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Action:       &action,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Bet", "raise")
	})
	t.Run("bet missing action", func(t *testing.T) {
		run(t, `{"command":"bet","sessionId":"s1"}`, mustBouillotteOutputJSON("param error: action is required."), http.StatusBadRequest)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("next alias", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustBouillotteOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustBouillotteOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestBouillotteWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("custom config passed through", func(t *testing.T) {
		players, ante, chips, rounds := 3, 25, 500, 20
		expected := domain.BouillotteConfig{PlayerCount: 3, Ante: 25, StartingChips: 500, TargetRounds: 20}
		biMock := new(usecase.MockBouillotteInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewBouillotteWebController(func() uc.BouillotteInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.BouillotteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config: &controller.BouillotteWebConfig{
				PlayerCount: &players, Ante: &ante, StartingChips: &chips, TargetRounds: &rounds,
			},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range values fall back to default", func(t *testing.T) {
		players := 99
		expected := domain.DefaultBouillotteConfig()
		biMock := new(usecase.MockBouillotteInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewBouillotteWebController(func() uc.BouillotteInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.BouillotteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.BouillotteWebConfig{PlayerCount: &players},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestBouillotteWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockBouillotteInteractor)
	c := controller.NewBouillotteWebController(func() uc.BouillotteInteractorIF { return biMock })
	c.Stop()
	c.Stop()
}
