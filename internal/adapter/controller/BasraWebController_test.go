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

func mustBasraOutputJSON(msg string) string {
	out := &controller.BasraWebOutput{
		Players:         []*controller.BasraWebOutputPlayer{},
		TableCards:      []*controller.WebOutputCard{},
		PlayableIndices: []int{},
		CaptureOptions:  map[int][]int{},
		Winners:         []int{},
		LastCaptureIdx:  -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBasraOutputJSON: %v", err))
	}
	return string(b)
}

func TestBasraWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockBasraInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBasraConfig()).Return(mockOutput)
	biMock.On("Play", 3, []int(nil)).Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BasraInteractorIF { return biMock }
	ctrl := controller.NewBasraWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.BasraWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustBasraOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.BasraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustBasraOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustBasraOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustBasraOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestBasraWebController_PlayWithTableIndices(t *testing.T) {
	mockOutput := `{"phase":0}`
	biMock := new(usecase.MockBasraInteractor)
	biMock.On("Play", 0, []int{1, 2}).Return(mockOutput)
	ctrl := controller.NewBasraWebController(func() uc.BasraInteractorIF { return biMock })
	defer ctrl.Stop()

	input := controller.BasraWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		CardIndex:    func() *int { v := 0; return &v }(),
		TableIndices: []int{1, 2},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	biMock.AssertCalled(t, "Play", 0, []int{1, 2})
}

func TestBasraWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		expected := domain.BasraConfig{CpuDifficulty: domain.BasraCpuDifficultyHard}
		biMock := new(usecase.MockBasraInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewBasraWebController(func() uc.BasraInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.BasraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.BasraWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultBasraConfig()
		biMock := new(usecase.MockBasraInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewBasraWebController(func() uc.BasraInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.BasraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.BasraWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultBasraConfig()
		biMock := new(usecase.MockBasraInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewBasraWebController(func() uc.BasraInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.BasraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestBasraWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockBasraInteractor)
	c := controller.NewBasraWebController(func() uc.BasraInteractorIF { return biMock })
	c.Stop()
	c.Stop()
}
