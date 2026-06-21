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

func mustChinchonOutputJSON(msg string) string {
	out := &controller.ChinchonWebOutput{
		Players:       []*controller.ChinchonWebOutputPlayer{},
		WinnerIdx:     -1,
		KnockerIdx:    -1,
		KnockerMelds:  []*controller.ChinchonWebOutputMeld{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustChinchonOutputJSON: %v", err))
	}
	return string(b)
}

func TestChinchonWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`

	siMock := new(usecase.MockChinchonInteractor)
	siMock.On("ResetWithConfig", domain.DefaultChinchonConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("Knock", 0).Return(mockOutput)
	siMock.On("Layoff", mock.Anything).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ChinchonInteractorIF { return siMock }
	ctrl := controller.NewChinchonWebController(factory)
	defer ctrl.Stop()

	cases := []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset reset", `{"command":"reset","sessionId":"s1"}`},
		{"drawstock ds", `{"command":"ds","sessionId":"s1"}`},
		{"drawdiscard dd", `{"command":"dd","sessionId":"s1"}`},
		{"layoff lo", `{"command":"lo","sessionId":"s1","cardIndices":[0,1]}`},
		{"nextround nr", `{"command":"nr","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.ChinchonWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(mockOutput)
		})
	}

	t.Run("discard with index", func(t *testing.T) {
		input := controller.ChinchonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("knock with index", func(t *testing.T) {
		input := controller.ChinchonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "k", SessionID: "s1"},
			CardIndex:    func() *int { v := 0; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("discard no cardIndex", func(t *testing.T) {
		var input controller.ChinchonWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustChinchonOutputJSON("param error: cardIndex is required."))
	})

	t.Run("knock no cardIndex", func(t *testing.T) {
		var input controller.ChinchonWebInput
		_ = json.Unmarshal([]byte(`{"command":"k","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustChinchonOutputJSON("param error: cardIndex is required."))
	})

	t.Run("unsupported", func(t *testing.T) {
		var input controller.ChinchonWebInput
		_ = json.Unmarshal([]byte(`{"command":"zzz","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustChinchonOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.ChinchonWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustChinchonOutputJSON("param error."))
	})
}

func TestChinchonWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		players := 2
		knock := 3
		elim := 50
		expected := domain.ChinchonConfig{CpuDifficulty: domain.ChinchonCpuDifficultyHard, PlayerCount: 2, KnockThreshold: 3, EliminationLimit: 50}
		siMock := new(usecase.MockChinchonInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewChinchonWebController(func() uc.ChinchonInteractorIF { return siMock })
		defer ctrl.Stop()
		input := controller.ChinchonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.ChinchonWebConfig{CpuDifficulty: &diff, PlayerCount: &players, KnockThreshold: &knock, EliminationLimit: &elim},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range config falls back to defaults", func(t *testing.T) {
		diff := 99
		players := 9
		expected := domain.DefaultChinchonConfig()
		siMock := new(usecase.MockChinchonInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewChinchonWebController(func() uc.ChinchonInteractorIF { return siMock })
		defer ctrl.Stop()
		input := controller.ChinchonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.ChinchonWebConfig{CpuDifficulty: &diff, PlayerCount: &players},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestChinchonWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockChinchonInteractor)
	c := controller.NewChinchonWebController(func() uc.ChinchonInteractorIF { return siMock })
	c.Stop()
	c.Stop()
}
