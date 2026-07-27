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

func mustTwoTenJackOutputJSON(msg string) string {
	out := &controller.TwoTenJackWebOutput{
		Players:       []*controller.TwoTenJackWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		TrumpSuit:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTwoTenJackOutputJSON: %v", err))
	}
	return string(b)
}

func TestTwoTenJackWebController_Exec(t *testing.T) {
	mockOutput := `{"ok":true}`

	siMock := new(usecase.MockTwoTenJackInteractor)
	siMock.On("ResetWithConfig", domain.DefaultTwoTenJackConfig()).Return(mockOutput)
	siMock.On("DeclareTrump", 1).Return(mockOutput)
	siMock.On("Play", 3).Return(mockOutput)
	siMock.On("NextTrick").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TwoTenJackInteractorIF { return siMock }
	ctrl := controller.NewTwoTenJackWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustTwoTenJackOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("declare", func(t *testing.T) {
		suit := 1
		input := controller.TwoTenJackWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "ttj-1"},
			TrumpSuit:    &suit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("declare missing field", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTwoTenJackOutputJSON("param error: trumpSuit is required."))
	})

	t.Run("play", func(t *testing.T) {
		idx := 3
		input := controller.TwoTenJackWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "ttj-1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing field", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTwoTenJackOutputJSON("param error: cardIndex is required."))
	})

	t.Run("next", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("unknown", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.TwoTenJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"ttj-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestTwoTenJackWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"ok":true}`

	diff := 2
	limit := 80
	expected := domain.TwoTenJackConfig{CpuDifficulty: domain.TwoTenJackCpuDifficultyHard, PointLimit: 80}
	siMock := new(usecase.MockTwoTenJackInteractor)
	siMock.On("ResetWithConfig", expected).Return(mockOutput)

	factory := func() uc.TwoTenJackInteractorIF { return siMock }
	ctrl := controller.NewTwoTenJackWebController(factory)
	defer ctrl.Stop()

	input := controller.TwoTenJackWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "ttj-cfg-1"},
		Config:       &controller.TwoTenJackWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	siMock.AssertCalled(t, "ResetWithConfig", expected)
}

func TestTwoTenJackWebInput_ToConfig_NilConfig(t *testing.T) {
	input := controller.TwoTenJackWebInput{}
	cfg := input.ToConfig()
	if cfg != domain.DefaultTwoTenJackConfig() {
		t.Errorf("expected default config, got %+v", cfg)
	}
}
