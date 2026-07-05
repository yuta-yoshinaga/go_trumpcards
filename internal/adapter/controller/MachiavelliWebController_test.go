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

func mustMachiavelliOutputJSON(msg string) string {
	out := &controller.MachiavelliWebOutput{
		Players:        []*controller.MachiavelliWebOutputPlayer{},
		Table:          []*controller.MachiavelliWebOutputMeld{},
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMachiavelliOutputJSON: %v", err))
	}
	return string(b)
}

func TestMachiavelliWebController_Method(t *testing.T) {
	mockOutput := `{"ok":true}`

	refs := [][]domain.MachiavelliCardRef{{{Design: 1, Value: 3}, {Design: 1, Value: 4}, {Design: 1, Value: 5}}}

	siMock := new(usecase.MockMachiavelliInteractor)
	siMock.On("ResetWithConfig", domain.DefaultMachiavelliConfig()).Return(mockOutput)
	siMock.On("Draw").Return(mockOutput)
	siMock.On("Play", refs, []int{0, 1, 2}).Return(mockOutput)
	siMock.On("NewMeld", []int{0, 1, 2}).Return(mockOutput)
	siMock.On("Layoff", 0, 2).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.MachiavelliInteractorIF { return siMock }
	ctrl := controller.NewMachiavelliWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, input *controller.MachiavelliWebInput) *recorded {
		return execRequest(t, ctrl.Exec, input)
	}

	t.Run("q exits", func(t *testing.T) {
		input := &controller.MachiavelliWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "ma-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mustMachiavelliOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			input := &controller.MachiavelliWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ma-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("draw", func(t *testing.T) {
		for _, cmd := range []string{"dr", "draw"} {
			input := &controller.MachiavelliWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ma-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("play", func(t *testing.T) {
		for _, cmd := range []string{"p", "play"} {
			input := &controller.MachiavelliWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ma-1"},
				TableMelds: [][]controller.MachiavelliCardRefInput{{
					{Design: 1, Value: 3}, {Design: 1, Value: 4}, {Design: 1, Value: 5},
				}},
				HandIndices: []int{0, 1, 2},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("newmeld", func(t *testing.T) {
		for _, cmd := range []string{"nm", "newmeld"} {
			input := &controller.MachiavelliWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ma-1"},
				HandIndices:  []int{0, 1, 2},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("layoff missing params", func(t *testing.T) {
		input := &controller.MachiavelliWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "ma-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustMachiavelliOutputJSON("param error: meldIdx and handIndex are required."))
	})

	t.Run("layoff ok", func(t *testing.T) {
		meld, hand := 0, 2
		input := &controller.MachiavelliWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "ma-1"},
			MeldIdx:      &meld,
			HandIndex:    &hand,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			input := &controller.MachiavelliWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ma-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			input := &controller.MachiavelliWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ma-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		input := &controller.MachiavelliWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bogus", SessionID: "ma-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustMachiavelliOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		input := &controller.MachiavelliWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "", SessionID: "ma-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustMachiavelliOutputJSON("param error."))
	})
}

func TestMachiavelliWebInput_ToConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		in := controller.MachiavelliWebInput{}
		got := in.ToConfig()
		want := domain.DefaultMachiavelliConfig()
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
	t.Run("custom values", func(t *testing.T) {
		pc := 5
		diff := int(domain.MachiavelliCpuDifficultyHard)
		rounds := 4
		in := controller.MachiavelliWebInput{
			Config: &controller.MachiavelliWebConfig{
				PlayerCount:   &pc,
				CpuDifficulty: &diff,
				TargetRounds:  &rounds,
			},
		}
		got := in.ToConfig()
		if got.PlayerCount != 5 {
			t.Errorf("playerCount = %d", got.PlayerCount)
		}
		if got.CpuDifficulty != domain.MachiavelliCpuDifficultyHard {
			t.Errorf("difficulty = %d", got.CpuDifficulty)
		}
		if got.TargetRounds != 4 {
			t.Errorf("rounds = %d", got.TargetRounds)
		}
	})
	t.Run("out of range clamps", func(t *testing.T) {
		pc := 99
		in := controller.MachiavelliWebInput{
			Config: &controller.MachiavelliWebConfig{PlayerCount: &pc},
		}
		got := in.ToConfig()
		if got.PlayerCount < domain.MachiavelliPlayerCountMin || got.PlayerCount > domain.MachiavelliPlayerCountMax {
			t.Errorf("playerCount out of range: %d", got.PlayerCount)
		}
	})
}
