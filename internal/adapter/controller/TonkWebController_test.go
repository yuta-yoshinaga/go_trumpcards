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

func mustTonkOutputJSON(msg string) string {
	out := &controller.TonkWebOutput{
		Players:          []*controller.TonkWebOutputPlayer{},
		WinnerIdx:        -1,
		KnockerIdx:       -1,
		KnockerMelds:     []*controller.TonkWebOutputMeld{},
		KnockerDeadwood:  []*controller.WebOutputCard{},
		OpponentMelds:    []*controller.TonkWebOutputMeld{},
		OpponentDeadwood: []*controller.WebOutputCard{},
		// 閾値は盤面が無くても規則なので、既定の応答にも乗る (#5582)。
		UndercutRiskMax: domain.TonkUndercutRiskMax,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTonkOutputJSON: %v", err))
	}
	return string(b)
}

func TestTonkWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockTonkInteractor)
	siMock.On("ResetWithConfig", domain.DefaultTonkConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("Knock", 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TonkInteractorIF { return siMock }
	ctrl := controller.NewTonkWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.TonkWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustTonkOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.TonkWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustTonkOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			var input controller.TonkWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("draw stock and discard", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock", "dd", "drawdiscard"} {
			var input controller.TonkWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("discard with index", func(t *testing.T) {
		for _, cmd := range []string{"d", "discard"} {
			var input controller.TonkWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","cardIndex":3,"sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("knock with index", func(t *testing.T) {
		for _, cmd := range []string{"k", "knock"} {
			var input controller.TonkWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","cardIndex":3,"sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("nextround", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			var input controller.TonkWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			var input controller.TonkWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.TonkWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTonkOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.TonkWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTonkOutputJSON("param error."))
	})

	t.Run("empty sessionId", func(t *testing.T) {
		var input controller.TonkWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTonkOutputJSON("param error."))
	})

	t.Run("sessionId too long", func(t *testing.T) {
		input := controller.TonkWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTonkOutputJSON("param error."))
	})

	t.Run("discard no cardIndex", func(t *testing.T) {
		var input controller.TonkWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTonkOutputJSON("param error: cardIndex is required."))
	})

	t.Run("knock no cardIndex", func(t *testing.T) {
		var input controller.TonkWebInput
		_ = json.Unmarshal([]byte(`{"command":"k","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTonkOutputJSON("param error: cardIndex is required."))
	})
}

func TestTonkWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 100
		expected := domain.TonkConfig{CpuDifficulty: domain.TonkCpuDifficultyHard, PointLimit: 100}
		siMock := new(usecase.MockTonkInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.TonkInteractorIF { return siMock }
		ctrl := controller.NewTonkWebController(factory)
		defer ctrl.Stop()

		input := controller.TonkWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
			Config:       &controller.TonkWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range cpuDifficulty falls back to default", func(t *testing.T) {
		diff := 5
		expected := domain.DefaultTonkConfig()
		siMock := new(usecase.MockTonkInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.TonkInteractorIF { return siMock }
		ctrl := controller.NewTonkWebController(factory)
		defer ctrl.Stop()

		input := controller.TonkWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-2"},
			Config:       &controller.TonkWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit out of range is ignored", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultTonkConfig()
		siMock := new(usecase.MockTonkInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.TonkInteractorIF { return siMock }
		ctrl := controller.NewTonkWebController(factory)
		defer ctrl.Stop()

		input := controller.TonkWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-3"},
			Config:       &controller.TonkWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTonkConfig()
		siMock := new(usecase.MockTonkInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.TonkInteractorIF { return siMock }
		ctrl := controller.NewTonkWebController(factory)
		defer ctrl.Stop()

		input := controller.TonkWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-4"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}
