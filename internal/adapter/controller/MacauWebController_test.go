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

func mustMacauOutputJSON(msg string) string {
	out := &controller.MacauWebOutput{
		Players:       []*controller.MacauWebOutputPlayer{},
		WinnerIdx:     -1,
		Direction:     1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMacauOutputJSON: %v", err))
	}
	return string(b)
}

func TestMacauWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	expectedBody := mockOutput

	siMock := new(usecase.MockMacauInteractor)
	siMock.On("ResetWithConfig", domain.DefaultMacauConfig()).Return(mockOutput)
	siMock.On("Play", 3).Return(mockOutput)
	siMock.On("ChooseSuit", 2).Return(mockOutput)
	siMock.On("Draw").Return(mockOutput)
	siMock.On("Declare").Return(mockOutput)
	siMock.On("SkipDeclare").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.MacauInteractorIF { return siMock }
	ctrl := controller.NewMacauWebController(factory)
	defer ctrl.Stop()

	cases := []struct {
		name string
		body string
		in   *controller.MacauWebInput
	}{
		{name: "reset r", body: `{"command":"r","sessionId":"s1"}`},
		{name: "reset", body: `{"command":"reset","sessionId":"s1"}`},
		{name: "draw d", body: `{"command":"d","sessionId":"s1"}`},
		{name: "draw", body: `{"command":"draw","sessionId":"s1"}`},
		{name: "declare dc", body: `{"command":"dc","sessionId":"s1"}`},
		{name: "declare", body: `{"command":"declare","sessionId":"s1"}`},
		{name: "skipdeclare sk", body: `{"command":"sk","sessionId":"s1"}`},
		{name: "skipdeclare", body: `{"command":"skipdeclare","sessionId":"s1"}`},
		{name: "nextround nr", body: `{"command":"nr","sessionId":"s1"}`},
		{name: "log", body: `{"command":"log","sessionId":"s1"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.MacauWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(expectedBody)
		})
	}

	t.Run("play with cardIndex", func(t *testing.T) {
		idx := 3
		input := controller.MacauWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("suit with suit param", func(t *testing.T) {
		s := 2
		input := controller.MacauWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "suit", SessionID: "s1"},
			Suit:         &s,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("q says bye", func(t *testing.T) {
		var input controller.MacauWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustMacauOutputJSON("bye."))
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.MacauWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMacauOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.MacauWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMacauOutputJSON("param error."))
	})

	t.Run("sessionId too long", func(t *testing.T) {
		input := controller.MacauWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMacauOutputJSON("param error."))
	})

	t.Run("play missing cardIndex", func(t *testing.T) {
		var input controller.MacauWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMacauOutputJSON("param error: cardIndex is required."))
	})

	t.Run("suit missing suit", func(t *testing.T) {
		var input controller.MacauWebInput
		_ = json.Unmarshal([]byte(`{"command":"suit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMacauOutputJSON("param error: suit is required."))
	})
}

func TestMacauWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values passed", func(t *testing.T) {
		diff := 2
		limit := 300
		expected := domain.MacauConfig{CpuDifficulty: domain.MacauCpuDifficultyHard, PointLimit: 300}
		siMock := new(usecase.MockMacauInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		ctrl := controller.NewMacauWebController(func() uc.MacauInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.MacauWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.MacauWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range values fall back to default", func(t *testing.T) {
		diff := 5
		limit := 0
		expected := domain.DefaultMacauConfig()
		siMock := new(usecase.MockMacauInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		ctrl := controller.NewMacauWebController(func() uc.MacauInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.MacauWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.MacauWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultMacauConfig()
		siMock := new(usecase.MockMacauInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		ctrl := controller.NewMacauWebController(func() uc.MacauInteractorIF { return siMock })
		defer ctrl.Stop()

		input := controller.MacauWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestMacauWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockMacauInteractor)
	c := controller.NewMacauWebController(func() uc.MacauInteractorIF { return siMock })
	c.Stop()
	c.Stop()
}
