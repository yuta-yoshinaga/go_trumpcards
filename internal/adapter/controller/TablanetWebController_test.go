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

func mustTablanetOutputJSON(msg string) string {
	out := &controller.TablanetWebOutput{
		Players:         []*controller.TablanetWebOutputPlayer{},
		TableCards:      []*controller.WebOutputCard{},
		PlayableIndices: []int{},
		CaptureOptions:  map[int][]int{},
		Winners:         []int{},
		LastCaptureIdx:  -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTablanetOutputJSON: %v", err))
	}
	return string(b)
}

func TestTablanetWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockTablanetInteractor)
	biMock.On("ResetWithConfig", domain.DefaultTablanetConfig()).Return(mockOutput)
	biMock.On("Play", 3, []int(nil)).Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TablanetInteractorIF { return biMock }
	ctrl := controller.NewTablanetWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.TablanetWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTablanetOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.TablanetWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustTablanetOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustTablanetOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustTablanetOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestTablanetWebController_PlayWithTableIndices(t *testing.T) {
	mockOutput := `{"phase":0}`
	biMock := new(usecase.MockTablanetInteractor)
	biMock.On("Play", 0, []int{1, 2}).Return(mockOutput)
	ctrl := controller.NewTablanetWebController(func() uc.TablanetInteractorIF { return biMock })
	defer ctrl.Stop()

	input := controller.TablanetWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
		CardIndex:    func() *int { v := 0; return &v }(),
		TableIndices: []int{1, 2},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	biMock.AssertCalled(t, "Play", 0, []int{1, 2})
}

func TestTablanetWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		expected := domain.TablanetConfig{CpuDifficulty: domain.TablanetCpuDifficultyHard}
		biMock := new(usecase.MockTablanetInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTablanetWebController(func() uc.TablanetInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.TablanetWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.TablanetWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultTablanetConfig()
		biMock := new(usecase.MockTablanetInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTablanetWebController(func() uc.TablanetInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.TablanetWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.TablanetWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTablanetConfig()
		biMock := new(usecase.MockTablanetInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTablanetWebController(func() uc.TablanetInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.TablanetWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestTablanetWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockTablanetInteractor)
	c := controller.NewTablanetWebController(func() uc.TablanetInteractorIF { return biMock })
	c.Stop()
	c.Stop()
}
