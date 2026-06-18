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

func mustKalookiOutputJSON(msg string) string {
	out := &controller.KalookiWebOutput{
		Players:        []*controller.KalookiWebOutputPlayer{},
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustKalookiOutputJSON: %v", err))
	}
	return string(b)
}

func TestKalookiWebController_Method(t *testing.T) {
	mockOutput := `{"ok":true}`

	siMock := new(usecase.MockKalookiInteractor)
	siMock.On("ResetWithConfig", domain.DefaultKalookiConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Meld", [][]int{{0, 1, 2}, {3, 4, 5}}).Return(mockOutput)
	siMock.On("Layoff", 1, 0, 2).Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.KalookiInteractorIF { return siMock }
	ctrl := controller.NewKalookiWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, input *controller.KalookiWebInput) *recorded {
		return execRequest(t, ctrl.Exec, input)
	}

	t.Run("q exits", func(t *testing.T) {
		input := &controller.KalookiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "kl-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mustKalookiOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			input := &controller.KalookiWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "kl-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawstock", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			input := &controller.KalookiWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "kl-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawdiscard", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			input := &controller.KalookiWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "kl-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("meld", func(t *testing.T) {
		for _, cmd := range []string{"m", "meld"} {
			input := &controller.KalookiWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "kl-1"},
				MeldGroups:   [][]int{{0, 1, 2}, {3, 4, 5}},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("layoff missing params", func(t *testing.T) {
		input := &controller.KalookiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "kl-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustKalookiOutputJSON("param error: targetPlayerIdx, meldIdx, cardIndex are required."))
	})

	t.Run("layoff ok", func(t *testing.T) {
		target, meld, card := 1, 0, 2
		input := &controller.KalookiWebInput{
			BaseWebInput:    controller.BaseWebInput{Command: "lo", SessionID: "kl-1"},
			TargetPlayerIdx: &target,
			MeldIdx:         &meld,
			CardIndex:       &card,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("discard missing cardIndex", func(t *testing.T) {
		input := &controller.KalookiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "kl-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustKalookiOutputJSON("param error: cardIndex is required."))
	})

	t.Run("discard ok", func(t *testing.T) {
		v := 3
		input := &controller.KalookiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "kl-1"},
			CardIndex:    &v,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			input := &controller.KalookiWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "kl-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			input := &controller.KalookiWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "kl-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		input := &controller.KalookiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bogus", SessionID: "kl-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustKalookiOutputJSON("Unsupported command."))
	})
}

func TestKalookiWebInput_ToConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		in := controller.KalookiWebInput{}
		got := in.ToConfig()
		want := domain.DefaultKalookiConfig()
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
	t.Run("custom values", func(t *testing.T) {
		diff := int(domain.KalookiCpuDifficultyHard)
		players := 2
		threshold := 41
		in := controller.KalookiWebInput{
			Config: &controller.KalookiWebConfig{
				CpuDifficulty:    &diff,
				PlayerCount:      &players,
				OpeningThreshold: &threshold,
			},
		}
		got := in.ToConfig()
		if got.CpuDifficulty != domain.KalookiCpuDifficultyHard {
			t.Errorf("difficulty = %d", got.CpuDifficulty)
		}
		if got.PlayerCount != 2 {
			t.Errorf("playerCount = %d", got.PlayerCount)
		}
		if got.OpeningThreshold != 41 {
			t.Errorf("threshold = %d", got.OpeningThreshold)
		}
	})
	t.Run("out of range clamps", func(t *testing.T) {
		players := 99
		in := controller.KalookiWebInput{
			Config: &controller.KalookiWebConfig{PlayerCount: &players},
		}
		got := in.ToConfig()
		if got.PlayerCount < domain.KalookiMinPlayers || got.PlayerCount > domain.KalookiMaxPlayers {
			t.Errorf("playerCount out of range: %d", got.PlayerCount)
		}
	})
}
