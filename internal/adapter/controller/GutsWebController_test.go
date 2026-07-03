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

func mustGutsOutputJSON(msg string) string {
	out := &controller.GutsWebOutput{
		Players:        make([]*controller.GutsWebOutputPlayer, 0),
		Matchers:       make([]int, 0),
		WinnerIdx:      -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGutsOutputJSON: %v", err))
	}
	return string(b)
}

func TestGutsWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockGutsInteractor)
	biMock.On("ResetWithConfig", domain.DefaultGutsConfig()).Return(mockOutput)
	biMock.On("Declare", true).Return(mockOutput)
	biMock.On("Declare", false).Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.GutsInteractorIF { return biMock }
	ctrl := controller.NewGutsWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.GutsWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustGutsOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("declare in", func(t *testing.T) {
		decl := int(domain.GutsDeclarationIn)
		input := controller.GutsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "declare", SessionID: "s1"},
			Declaration:  &decl,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Declare", true)
	})
	t.Run("declare out", func(t *testing.T) {
		decl := int(domain.GutsDeclarationOut)
		input := controller.GutsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "declare", SessionID: "s1"},
			Declaration:  &decl,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Declare", false)
	})
	t.Run("declare missing declaration", func(t *testing.T) {
		run(t, `{"command":"declare","sessionId":"s1"}`, mustGutsOutputJSON("param error: declaration is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustGutsOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustGutsOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestGutsWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("custom config passed through", func(t *testing.T) {
		players, ante, chips, rounds := 6, 25, 500, 20
		expected := domain.GutsConfig{PlayerCount: 6, Ante: 25, StartingChips: 500, TargetRounds: 20}
		biMock := new(usecase.MockGutsInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGutsWebController(func() uc.GutsInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.GutsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config: &controller.GutsWebConfig{
				PlayerCount: &players, Ante: &ante, StartingChips: &chips, TargetRounds: &rounds,
			},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range values fall back to default", func(t *testing.T) {
		players := 99
		expected := domain.DefaultGutsConfig()
		biMock := new(usecase.MockGutsInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGutsWebController(func() uc.GutsInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.GutsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.GutsWebConfig{PlayerCount: &players},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestGutsWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockGutsInteractor)
	c := controller.NewGutsWebController(func() uc.GutsInteractorIF { return biMock })
	c.Stop()
	c.Stop()
}
