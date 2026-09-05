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

func mustBatakOutputJSON(msg string) string {
	out := &controller.BatakWebOutput{
		Players:          []*controller.BatakWebOutputPlayer{},
		CurrentTrick:     []*controller.WebOutputTrickCard{},
		ValidPlayIndices: []int{},
		WinnerIdx:        -1,
		DeclarerIdx:      -1,
		HighBid:          0,
		MinLegalBid:      0,
		WebOutputBase:    controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBatakOutputJSON: %v", err))
	}
	return string(b)
}

func TestBatakWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	ciMock := new(usecase.MockBatakInteractor)
	ciMock.On("ResetWithConfig", domain.DefaultBatakConfig()).Return(mockOutput)
	ciMock.On("Bid", 3).Return(mockOutput)
	ciMock.On("Bid", 0).Return(mockOutput)
	ciMock.On("Play", 3).Return(mockOutput)
	ciMock.On("NextTrick").Return(mockOutput)
	ciMock.On("NextRound").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BatakInteractorIF { return ciMock }
	ctrl := controller.NewBatakWebController(factory)
	defer ctrl.Stop()

	t.Run("q quits", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"cb1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mustBatakOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"cb1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("bid valid", func(t *testing.T) {
		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "cb1"},
			Bid:          func() *int { v := 3; return &v }(),
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("bid pass 0", func(t *testing.T) {
		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "cb1"},
			Bid:          func() *int { v := 0; return &v }(),
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
		ciMock.AssertCalled(t, "Bid", 0)
	})

	t.Run("pass command", func(t *testing.T) {
		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "pass", SessionID: "cb1"},
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
		ciMock.AssertCalled(t, "Bid", domain.BatakPassBid)
	})

	t.Run("bid b shorthand", func(t *testing.T) {
		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "cb1"},
			Bid:          func() *int { v := 3; return &v }(),
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("play valid", func(t *testing.T) {
		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "cb1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("next n", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"cb1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"cb1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"cb1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"cb1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"cb1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBatakOutputJSON("Unsupported command."))
	})

	t.Run("bid missing bid field", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"bid","sessionId":"cb1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBatakOutputJSON("param error: bid is required."))
	})

	t.Run("play missing cardIndex", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"cb1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBatakOutputJSON("param error: cardIndex is required."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.BatakWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"cb1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBatakOutputJSON("param error."))
	})

	t.Run("session id too long", func(t *testing.T) {
		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBatakOutputJSON("param error."))
	})
}

func TestBatakWebController_CustomConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom max rounds and difficulty", func(t *testing.T) {
		diff := 2
		rounds := 10
		expected := domain.BatakConfig{CpuDifficulty: domain.BatakCpuDifficultyHard, MaxRounds: 10}
		ciMock := new(usecase.MockBatakInteractor)
		ciMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BatakInteractorIF { return ciMock }
		ctrl := controller.NewBatakWebController(factory)
		defer ctrl.Stop()

		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cb-cfg-1"},
			Config:       &controller.BatakWebConfig{CpuDifficulty: &diff, MaxRounds: &rounds},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		ciMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpu difficulty above max falls back to default", func(t *testing.T) {
		diff := 99
		expected := domain.DefaultBatakConfig()
		ciMock := new(usecase.MockBatakInteractor)
		ciMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BatakInteractorIF { return ciMock }
		ctrl := controller.NewBatakWebController(factory)
		defer ctrl.Stop()

		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cb-cfg-2"},
			Config:       &controller.BatakWebConfig{CpuDifficulty: &diff},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		ciMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		ciMock := new(usecase.MockBatakInteractor)
		ciMock.On("ResetWithConfig", domain.DefaultBatakConfig()).Return(mockOutput)

		factory := func() uc.BatakInteractorIF { return ciMock }
		ctrl := controller.NewBatakWebController(factory)
		defer ctrl.Stop()

		input := controller.BatakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cb-cfg-3"},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	})
}
