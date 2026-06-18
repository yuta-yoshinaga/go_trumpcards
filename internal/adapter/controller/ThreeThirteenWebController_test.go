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

func mustThreeThirteenOutputJSON(msg string) string {
	out := &controller.ThreeThirteenWebOutput{
		Players:       []*controller.ThreeThirteenWebOutputPlayer{},
		WinnerIdx:     -1,
		KnockerIdx:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustThreeThirteenOutputJSON: %v", err))
	}
	return string(b)
}

func TestThreeThirteenWebController_Method(t *testing.T) {
	mockOutput := `{"ok":true}`

	siMock := new(usecase.MockThreeThirteenInteractor)
	siMock.On("ResetWithConfig", domain.DefaultThreeThirteenConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("Knock", 2).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ThreeThirteenInteractorIF { return siMock }
	ctrl := controller.NewThreeThirteenWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, input *controller.ThreeThirteenWebInput) *recorded {
		return execRequest(t, ctrl.Exec, input)
	}

	t.Run("q exits", func(t *testing.T) {
		input := &controller.ThreeThirteenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "tt-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mustThreeThirteenOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			input := &controller.ThreeThirteenWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "tt-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawstock", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			input := &controller.ThreeThirteenWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "tt-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("drawdiscard", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			input := &controller.ThreeThirteenWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "tt-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("discard missing cardIndex", func(t *testing.T) {
		input := &controller.ThreeThirteenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "tt-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustThreeThirteenOutputJSON("param error: cardIndex is required."))
	})

	t.Run("discard ok", func(t *testing.T) {
		v := 3
		input := &controller.ThreeThirteenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "tt-1"},
			CardIndex:    &v,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("knock missing cardIndex", func(t *testing.T) {
		input := &controller.ThreeThirteenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "k", SessionID: "tt-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustThreeThirteenOutputJSON("param error: cardIndex is required."))
	})

	t.Run("knock ok", func(t *testing.T) {
		v := 2
		input := &controller.ThreeThirteenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "k", SessionID: "tt-1"},
			CardIndex:    &v,
		}
		got := run(t, input)
		got.CodeIs(http.StatusOK)
		got.BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			input := &controller.ThreeThirteenWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "tt-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			input := &controller.ThreeThirteenWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "tt-1"},
			}
			got := run(t, input)
			got.CodeIs(http.StatusOK)
			got.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		input := &controller.ThreeThirteenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bogus", SessionID: "tt-1"},
		}
		got := run(t, input)
		got.CodeIs(http.StatusBadRequest)
		got.BodyIs(mustThreeThirteenOutputJSON("Unsupported command."))
	})
}

func TestThreeThirteenWebInput_ToConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		in := controller.ThreeThirteenWebInput{}
		assert := func(got, want domain.ThreeThirteenConfig) {
			if got != want {
				t.Errorf("got %+v, want %+v", got, want)
			}
		}
		assert(in.ToConfig(), domain.DefaultThreeThirteenConfig())
	})
	t.Run("custom values", func(t *testing.T) {
		diff := int(domain.ThreeThirteenCpuDifficultyHard)
		players := 2
		in := controller.ThreeThirteenWebInput{
			Config: &controller.ThreeThirteenWebConfig{
				CpuDifficulty: &diff,
				PlayerCount:   &players,
			},
		}
		got := in.ToConfig()
		if got.CpuDifficulty != domain.ThreeThirteenCpuDifficultyHard {
			t.Errorf("difficulty = %d", got.CpuDifficulty)
		}
		if got.PlayerCount != 2 {
			t.Errorf("playerCount = %d", got.PlayerCount)
		}
	})
	t.Run("out of range clamps", func(t *testing.T) {
		players := 99
		in := controller.ThreeThirteenWebInput{
			Config: &controller.ThreeThirteenWebConfig{PlayerCount: &players},
		}
		got := in.ToConfig()
		if got.PlayerCount < domain.ThreeThirteenMinPlayers || got.PlayerCount > domain.ThreeThirteenMaxPlayers {
			t.Errorf("playerCount out of range: %d", got.PlayerCount)
		}
	})
}
