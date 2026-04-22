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

func mustSevenBridgeOutputJSON(msg string) string {
	out := &controller.SevenBridgeWebOutput{
		Players:        []*controller.SevenBridgeWebOutputPlayer{},
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSevenBridgeOutputJSON: %v", err))
	}
	return string(b)
}

func TestSevenBridgeWebController_Method(t *testing.T) {
	mockOutput := `{"ok":true}`
	expected := mockOutput

	siMock := new(usecase.MockSevenBridgeInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSevenBridgeConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("ClaimPon", []int{0, 1}).Return(mockOutput)
	siMock.On("ClaimChi", []int{0, 1}).Return(mockOutput)
	siMock.On("Meld", []int{0, 1, 2}).Return(mockOutput)
	siMock.On("Layoff", 1, 0, 2).Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.SevenBridgeInteractorIF { return siMock }
	ctrl := controller.NewSevenBridgeWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, input *controller.SevenBridgeWebInput) *recorded {
		return execRequest(t, ctrl.Exec, input)
	}

	t.Run("q exits", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "sb-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mustSevenBridgeOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			input := &controller.SevenBridgeWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "sb-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(expected)
		}
	})

	t.Run("drawstock", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			input := &controller.SevenBridgeWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "sb-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(expected)
		}
	})

	t.Run("pon", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "pon", SessionID: "sb-1"},
			CardIndices:  []int{0, 1},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(expected)
	})

	t.Run("chi", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "chi", SessionID: "sb-1"},
			CardIndices:  []int{0, 1},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(expected)
	})

	t.Run("meld", func(t *testing.T) {
		for _, cmd := range []string{"m", "meld"} {
			input := &controller.SevenBridgeWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "sb-1"},
				CardIndices:  []int{0, 1, 2},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(expected)
		}
	})

	t.Run("layoff missing params", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "sb-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustSevenBridgeOutputJSON("param error: targetPlayerIdx, meldIdx, cardIndex are required."))
	})

	t.Run("layoff ok", func(t *testing.T) {
		target, meld, card := 1, 0, 2
		input := &controller.SevenBridgeWebInput{
			BaseWebInput:    controller.BaseWebInput{Command: "lo", SessionID: "sb-1"},
			TargetPlayerIdx: &target,
			MeldIdx:         &meld,
			CardIndex:       &card,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(expected)
	})

	t.Run("discard missing cardIndex", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "sb-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustSevenBridgeOutputJSON("param error: cardIndex is required."))
	})

	t.Run("discard ok", func(t *testing.T) {
		v := 3
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "sb-1"},
			CardIndex:    &v,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(expected)
	})

	t.Run("nextround", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			input := &controller.SevenBridgeWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "sb-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(expected)
		}
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			input := &controller.SevenBridgeWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "sb-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(expected)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bogus", SessionID: "sb-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustSevenBridgeOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "", SessionID: "sb-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustSevenBridgeOutputJSON("param error."))
	})

	t.Run("empty session id", func(t *testing.T) {
		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: ""},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustSevenBridgeOutputJSON("param error."))
	})
}

func TestSevenBridgeWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config propagated", func(t *testing.T) {
		diff := 2
		limit := 200
		expected := domain.SevenBridgeConfig{CpuDifficulty: domain.SevenBridgeCpuDifficultyHard, PointLimit: 200}
		siMock := new(usecase.MockSevenBridgeInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SevenBridgeInteractorIF { return siMock }
		ctrl := controller.NewSevenBridgeWebController(factory)
		defer ctrl.Stop()

		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "sb-cfg-1"},
			Config:       &controller.SevenBridgeWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		got := execRequest(t, ctrl.Exec, input)
		got.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out-of-range config clamped to defaults", func(t *testing.T) {
		diff := 9
		limit := 0
		expected := domain.DefaultSevenBridgeConfig()
		siMock := new(usecase.MockSevenBridgeInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SevenBridgeInteractorIF { return siMock }
		ctrl := controller.NewSevenBridgeWebController(factory)
		defer ctrl.Stop()

		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "sb-cfg-2"},
			Config:       &controller.SevenBridgeWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		got := execRequest(t, ctrl.Exec, input)
		got.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultSevenBridgeConfig()
		siMock := new(usecase.MockSevenBridgeInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SevenBridgeInteractorIF { return siMock }
		ctrl := controller.NewSevenBridgeWebController(factory)
		defer ctrl.Stop()

		input := &controller.SevenBridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "sb-cfg-3"},
		}
		got := execRequest(t, ctrl.Exec, input)
		got.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}
