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

func mustContractRummyOutputJSON(msg string) string {
	out := &controller.ContractRummyWebOutput{
		Players:        []*controller.ContractRummyWebOutputPlayer{},
		ContractSlots:  []*controller.ContractRummyWebOutputContractSlot{},
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustContractRummyOutputJSON: %v", err))
	}
	return string(b)
}

func TestContractRummyWebController_Method(t *testing.T) {
	mockOutput := `{"ok":true}`

	siMock := new(usecase.MockContractRummyInteractor)
	siMock.On("ResetWithConfig", domain.DefaultContractRummyConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("MeldContract", [][]int{{0, 1, 2}, {3, 4, 5}}).Return(mockOutput)
	siMock.On("MeldExtra", []int{0, 1, 2}).Return(mockOutput)
	siMock.On("Layoff", 1, 0, 2).Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ContractRummyInteractorIF { return siMock }
	ctrl := controller.NewContractRummyWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, input *controller.ContractRummyWebInput) *recorded {
		return execRequest(t, ctrl.Exec, input)
	}

	t.Run("q exits", func(t *testing.T) {
		input := &controller.ContractRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "cr-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mustContractRummyOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			input := &controller.ContractRummyWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "cr-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawstock", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			input := &controller.ContractRummyWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "cr-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawdiscard", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			input := &controller.ContractRummyWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "cr-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("meldcontract", func(t *testing.T) {
		for _, cmd := range []string{"mc", "meldcontract"} {
			input := &controller.ContractRummyWebInput{
				BaseWebInput:   controller.BaseWebInput{Command: cmd, SessionID: "cr-1"},
				IndicesPerSlot: [][]int{{0, 1, 2}, {3, 4, 5}},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("meldextra", func(t *testing.T) {
		for _, cmd := range []string{"me", "meldextra"} {
			input := &controller.ContractRummyWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "cr-1"},
				CardIndices:  []int{0, 1, 2},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("layoff missing params", func(t *testing.T) {
		input := &controller.ContractRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "cr-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustContractRummyOutputJSON("param error: targetPlayerIdx, meldIdx, cardIndex are required."))
	})

	t.Run("layoff ok", func(t *testing.T) {
		target, meld, card := 1, 0, 2
		input := &controller.ContractRummyWebInput{
			BaseWebInput:    controller.BaseWebInput{Command: "lo", SessionID: "cr-1"},
			TargetPlayerIdx: &target,
			MeldIdx:         &meld,
			CardIndex:       &card,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("discard missing cardIndex", func(t *testing.T) {
		input := &controller.ContractRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "cr-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustContractRummyOutputJSON("param error: cardIndex is required."))
	})

	t.Run("discard ok", func(t *testing.T) {
		v := 3
		input := &controller.ContractRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "cr-1"},
			CardIndex:    &v,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			input := &controller.ContractRummyWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "cr-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			input := &controller.ContractRummyWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "cr-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		input := &controller.ContractRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bogus", SessionID: "cr-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustContractRummyOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		input := &controller.ContractRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "", SessionID: "cr-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustContractRummyOutputJSON("param error."))
	})
}

func TestContractRummyWebInput_ToConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		in := controller.ContractRummyWebInput{}
		got := in.ToConfig()
		want := domain.DefaultContractRummyConfig()
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
	t.Run("custom values", func(t *testing.T) {
		diff := int(domain.ContractRummyCpuDifficultyHard)
		pen := 50
		in := controller.ContractRummyWebInput{
			Config: &controller.ContractRummyWebConfig{
				CpuDifficulty:       &diff,
				FailContractPenalty: &pen,
			},
		}
		got := in.ToConfig()
		if got.CpuDifficulty != domain.ContractRummyCpuDifficultyHard {
			t.Errorf("difficulty = %d", got.CpuDifficulty)
		}
		if got.FailContractPenalty != 50 {
			t.Errorf("penalty = %d", got.FailContractPenalty)
		}
	})
	t.Run("out of range clamps", func(t *testing.T) {
		diff := 99
		in := controller.ContractRummyWebInput{
			Config: &controller.ContractRummyWebConfig{CpuDifficulty: &diff},
		}
		got := in.ToConfig()
		// Out-of-range value falls back to default
		if got.CpuDifficulty < domain.ContractRummyCpuDifficultyEasy || got.CpuDifficulty > domain.ContractRummyCpuDifficultyHard {
			t.Errorf("difficulty out of range: %d", got.CpuDifficulty)
		}
	})
}
