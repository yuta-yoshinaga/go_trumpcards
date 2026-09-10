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

func mustMarriageOutputJSON(msg string) string {
	out := &controller.MarriageWebOutput{
		Players:       []*controller.MarriageWebOutputPlayer{},
		WinnerIdx:     -1,
		DeclarerIdx:   -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMarriageOutputJSON: %v", err))
	}
	return string(b)
}

func TestMarriageWebOutput_MaalJSON(t *testing.T) {
	b, err := json.Marshal(&controller.MarriageWebOutputPlayer{Maal: 7})
	if err != nil || !strings.Contains(string(b), `"maal":7`) {
		t.Fatalf("maal is not exposed in JSON: %s (%v)", b, err)
	}
}

func TestMarriageWebController_Method(t *testing.T) {
	mockOutput := `{"ok":true}`

	siMock := new(usecase.MockMarriageInteractor)
	siMock.On("ResetWithConfig", domain.DefaultMarriageConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("Declare", 2).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.MarriageInteractorIF { return siMock }
	ctrl := controller.NewMarriageWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, input *controller.MarriageWebInput) *recorded {
		return execRequest(t, ctrl.Exec, input)
	}

	t.Run("q exits", func(t *testing.T) {
		input := &controller.MarriageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "ir-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mustMarriageOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			input := &controller.MarriageWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ir-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawstock", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			input := &controller.MarriageWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ir-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawdiscard", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			input := &controller.MarriageWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ir-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("discard missing cardIndex", func(t *testing.T) {
		input := &controller.MarriageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "ir-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustMarriageOutputJSON("param error: cardIndex is required."))
	})

	t.Run("discard ok", func(t *testing.T) {
		v := 3
		input := &controller.MarriageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "ir-1"},
			CardIndex:    &v,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("declare missing cardIndex", func(t *testing.T) {
		input := &controller.MarriageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "de", SessionID: "ir-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustMarriageOutputJSON("param error: cardIndex is required."))
	})

	t.Run("declare ok", func(t *testing.T) {
		v := 2
		input := &controller.MarriageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "de", SessionID: "ir-1"},
			CardIndex:    &v,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			input := &controller.MarriageWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ir-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			input := &controller.MarriageWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "ir-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		input := &controller.MarriageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bogus", SessionID: "ir-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustMarriageOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		input := &controller.MarriageWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "", SessionID: "ir-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustMarriageOutputJSON("param error."))
	})
}

func TestMarriageWebInput_ToConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		in := controller.MarriageWebInput{}
		got := in.ToConfig()
		want := domain.DefaultMarriageConfig()
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
	t.Run("custom values", func(t *testing.T) {
		pc := 3
		diff := int(domain.MarriageCpuDifficultyHard)
		rounds := 7
		in := controller.MarriageWebInput{
			Config: &controller.MarriageWebConfig{
				PlayerCount:   &pc,
				CpuDifficulty: &diff,
				TargetRounds:  &rounds,
			},
		}
		got := in.ToConfig()
		if got.PlayerCount != 3 {
			t.Errorf("playerCount = %d", got.PlayerCount)
		}
		if got.CpuDifficulty != domain.MarriageCpuDifficultyHard {
			t.Errorf("difficulty = %d", got.CpuDifficulty)
		}
		if got.TargetRounds != 7 {
			t.Errorf("targetRounds = %d", got.TargetRounds)
		}
	})
	t.Run("out of range clamps", func(t *testing.T) {
		pc := 99
		in := controller.MarriageWebInput{
			Config: &controller.MarriageWebConfig{PlayerCount: &pc},
		}
		got := in.ToConfig()
		if got.PlayerCount < domain.MarriagePlayerCountMin || got.PlayerCount > domain.MarriagePlayerCountMax {
			t.Errorf("playerCount out of range: %d", got.PlayerCount)
		}
	})
}
