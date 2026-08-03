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

func mustAluetteOutputJSON(msg string) string {
	out := &controller.AluetteWebOutput{
		Players:         []*controller.AluetteWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		Luettes:         domain.AluetteLuetteTable(),
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustAluetteOutputJSON: %v", err))
	}
	return string(b)
}

func TestAluetteWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockAluetteInteractor)
	diMock.On("ResetWithConfig", domain.DefaultAluetteConfig()).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewAluetteWebController(func() uc.AluetteInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.AluetteWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustAluetteOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.AluetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`,
			mustAluetteOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	// **捨て札も入札も無い。**タロー系から写した scarto / bid が素通りしては
	// ならない。
	t.Run("scarto and bid are not commands here", func(t *testing.T) {
		for _, cmd := range []string{"s", "scarto", "d", "discard", "bid"} {
			run(t, fmt.Sprintf(`{"command":%q,"sessionId":"s1"}`, cmd),
				mustAluetteOutputJSON("Unsupported command."), http.StatusBadRequest)
		}
	})
	t.Run("next / nextround / log / hint", func(t *testing.T) {
		for _, cmd := range []string{"n", "nr", "log", "h"} {
			run(t, fmt.Sprintf(`{"command":%q,"sessionId":"s1"}`, cmd), mockOutput, http.StatusOK)
		}
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`,
			mustAluetteOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`,
			mustAluetteOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestAluetteWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	newCtrl := func(expected domain.AluetteConfig) (*usecase.MockAluetteInteractor, *controller.AluetteWebController) {
		diMock := new(usecase.MockAluetteInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		return diMock, controller.NewAluetteWebController(func() uc.AluetteInteractorIF { return diMock })
	}

	t.Run("custom config passed through", func(t *testing.T) {
		diff, points := 2, 8
		expected := domain.AluetteConfig{CpuDifficulty: domain.AluetteCpuDifficultyHard, TargetPoints: 8}
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.AluetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.AluetteWebConfig{CpuDifficulty: &diff, TargetPoints: &points},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultAluetteConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.AluetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.AluetteWebConfig{CpuDifficulty: &diff},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("zero target points falls back to default", func(t *testing.T) {
		points := 0
		expected := domain.DefaultAluetteConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.AluetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
			Config:       &controller.AluetteWebConfig{TargetPoints: &points},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultAluetteConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.AluetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c4"},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestAluetteWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockAluetteInteractor)
	c := controller.NewAluetteWebController(func() uc.AluetteInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}
