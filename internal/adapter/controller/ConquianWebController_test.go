//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustConquianOutputJSON(msg string) string {
	out := &controller.ConquianWebOutput{
		Players:       []*controller.ConquianWebOutputPlayer{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustConquianOutputJSON: %v", err))
	}
	return string(b)
}

func TestConquianWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`

	siMock := new(usecase.MockConquianInteractor)
	siMock.On("ResetWithConfig", domain.DefaultConquianConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Meld", mock.Anything).Return(mockOutput)
	siMock.On("MeldWithTargets", mock.Anything, mock.Anything).Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ConquianInteractorIF { return siMock }
	ctrl := controller.NewConquianWebController(factory)
	defer ctrl.Stop()

	cases := []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset reset", `{"command":"reset","sessionId":"s1"}`},
		{"drawstock ds", `{"command":"ds","sessionId":"s1"}`},
		{"drawdiscard dd", `{"command":"dd","sessionId":"s1"}`},
		{"meld m", `{"command":"m","sessionId":"s1","meldGroups":[[0,1,2]]}`},
		{"nextround nr", `{"command":"nr","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.ConquianWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(mockOutput)
		})
	}

	t.Run("discard with index", func(t *testing.T) {
		input := controller.ConquianWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("discard no cardIndex", func(t *testing.T) {
		var input controller.ConquianWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustConquianOutputJSON("param error: cardIndex is required."))
	})

	t.Run("unsupported", func(t *testing.T) {
		var input controller.ConquianWebInput
		_ = json.Unmarshal([]byte(`{"command":"zzz","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustConquianOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.ConquianWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustConquianOutputJSON("param error."))
	})
}

func TestConquianWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		wins := 3
		expected := domain.ConquianConfig{CpuDifficulty: domain.ConquianCpuDifficultyHard, TargetWins: 3}
		siMock := new(usecase.MockConquianInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewConquianWebController(func() uc.ConquianInteractorIF { return siMock })
		defer ctrl.Stop()
		input := controller.ConquianWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.ConquianWebConfig{CpuDifficulty: &diff, TargetWins: &wins},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range config falls back to defaults", func(t *testing.T) {
		diff := 99
		wins := 0
		expected := domain.DefaultConquianConfig()
		siMock := new(usecase.MockConquianInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewConquianWebController(func() uc.ConquianInteractorIF { return siMock })
		defer ctrl.Stop()
		input := controller.ConquianWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.ConquianWebConfig{CpuDifficulty: &diff, TargetWins: &wins},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestConquianWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockConquianInteractor)
	c := controller.NewConquianWebController(func() uc.ConquianInteractorIF { return siMock })
	c.Stop()
	c.Stop()
}
