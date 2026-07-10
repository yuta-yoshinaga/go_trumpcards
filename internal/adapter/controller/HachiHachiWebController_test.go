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

func mustHachiHachiOutputJSON(msg string) string {
	out := &controller.HachiHachiWebOutput{
		Players:         []*controller.HachiHachiWebOutputPlayer{},
		FieldCards:      []*controller.WebOutputCard{},
		PlayableIndices: []int{},
		CaptureOptions:  map[int][]int{},
		Winner:          -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHachiHachiOutputJSON: %v", err))
	}
	return string(b)
}

func TestHachiHachiWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	hiMock := new(usecase.MockHachiHachiInteractor)
	hiMock.On("ResetWithConfig", domain.DefaultHachiHachiConfig()).Return(mockOutput)
	hiMock.On("Play", 3, -1).Return(mockOutput)
	hiMock.On("NextRound").Return(mockOutput)
	hiMock.On("Hint").Return(mockOutput)
	hiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.HachiHachiInteractorIF { return hiMock }
	ctrl := controller.NewHachiHachiWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.HachiHachiWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustHachiHachiOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.HachiHachiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustHachiHachiOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustHachiHachiOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustHachiHachiOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestHachiHachiWebController_PlayWithFieldIndex(t *testing.T) {
	mockOutput := `{"phase":0}`
	hiMock := new(usecase.MockHachiHachiInteractor)
	hiMock.On("Play", 0, 2).Return(mockOutput)
	ctrl := controller.NewHachiHachiWebController(func() uc.HachiHachiInteractorIF { return hiMock })
	defer ctrl.Stop()

	input := controller.HachiHachiWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		CardIndex:    func() *int { v := 0; return &v }(),
		FieldIndex:   func() *int { v := 2; return &v }(),
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	hiMock.AssertCalled(t, "Play", 0, 2)
}

func TestHachiHachiWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`
	diff := 2
	rounds := 6
	expected := domain.HachiHachiConfig{
		CpuDifficulty: domain.HachiHachiCpuDifficultyHard,
		TargetRounds:  6,
	}
	hiMock := new(usecase.MockHachiHachiInteractor)
	hiMock.On("ResetWithConfig", expected).Return(mockOutput)
	ctrl := controller.NewHachiHachiWebController(func() uc.HachiHachiInteractorIF { return hiMock })
	defer ctrl.Stop()

	input := controller.HachiHachiWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
		Config:       &controller.HachiHachiWebConfig{CpuDifficulty: &diff, TargetRounds: &rounds},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	hiMock.AssertCalled(t, "ResetWithConfig", expected)
}
