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

func mustPrsiOutputJSON(msg string) string {
	out := &controller.PrsiWebOutput{
		Players:       []*controller.PrsiWebOutputPlayer{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPrsiOutputJSON: %v", err))
	}
	return string(b)
}

func TestPrsiWebController_Exec(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockPrsiInteractor)
	siMock.On("ResetWithConfig", domain.DefaultPrsiConfig()).Return(mockOutput)
	siMock.On("Play", 3).Return(mockOutput)
	siMock.On("Draw").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewPrsiWebController(func() uc.PrsiInteractorIF { return siMock })
	defer ctrl.Stop()

	cases := []struct {
		name  string
		input controller.PrsiWebInput
		body  string
	}{
		{"reset r", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "r", SessionID: "s1"}}, mockOutput},
		{"reset long", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"}}, mockOutput},
		{"play p", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"}, CardIndex: ptrInt(3)}, mockOutput},
		{"play long", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"}, CardIndex: ptrInt(3)}, mockOutput},
		{"draw d", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"}}, mockOutput},
		{"draw long", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "draw", SessionID: "s1"}}, mockOutput},
		{"log", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "log", SessionID: "s1"}}, mockOutput},
		{"l shorthand", controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "l", SessionID: "s1"}}, mockOutput},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := tc.input
			recorded := execRequest(t, ctrl.Exec, &in)
			recorded.CodeIs(http.StatusOK)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(tc.body)
		})
	}

	t.Run("quit", func(t *testing.T) {
		in := controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustPrsiOutputJSON("bye."))
	})

	t.Run("play missing cardIndex", func(t *testing.T) {
		in := controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustPrsiOutputJSON("param error: cardIndex is required."))
	})

	t.Run("unsupported command", func(t *testing.T) {
		in := controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "zzz", SessionID: "s1"}}
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustPrsiOutputJSON("Unsupported command."))
	})
}

func TestPrsiWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom difficulty passed", func(t *testing.T) {
		diff := 2
		expected := domain.PrsiConfig{CpuDifficulty: domain.PrsiCpuDifficultyHard}
		siMock := new(usecase.MockPrsiInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewPrsiWebController(func() uc.PrsiInteractorIF { return siMock })
		defer ctrl.Stop()

		in := controller.PrsiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg1"},
			Config:       &controller.PrsiWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultPrsiConfig()
		siMock := new(usecase.MockPrsiInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewPrsiWebController(func() uc.PrsiInteractorIF { return siMock })
		defer ctrl.Stop()

		in := controller.PrsiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg2"},
			Config:       &controller.PrsiWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		expected := domain.DefaultPrsiConfig()
		siMock := new(usecase.MockPrsiInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewPrsiWebController(func() uc.PrsiInteractorIF { return siMock })
		defer ctrl.Stop()

		in := controller.PrsiWebInput{BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg3"}}
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func ptrInt(v int) *int { return &v }
