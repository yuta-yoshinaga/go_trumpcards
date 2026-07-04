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

func mustMichiganOutputJSON(msg string) string {
	out := &controller.MichiganWebOutput{
		Players:         make([]*controller.MichiganWebOutputPlayer, 0),
		Boodles:         make([]*controller.MichiganWebOutputBoodle, 0),
		PlayableIndices: make([]int, 0),
		WinnerIdx:       -1,
		MatchWinnerIdx:  -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMichiganOutputJSON: %v", err))
	}
	return string(b)
}

func TestMichiganWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	miMock := new(usecase.MockMichiganInteractor)
	miMock.On("ResetWithConfig", domain.DefaultMichiganConfig()).Return(mockOutput)
	miMock.On("Bet", mock2Ints()).Return(mockOutput)
	miMock.On("Play", 0).Return(mockOutput)
	miMock.On("NextRound").Return(mockOutput)
	miMock.On("Hint").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.MichiganInteractorIF { return miMock }
	ctrl := controller.NewMichiganWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.MichiganWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustMichiganOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bet", func(t *testing.T) {
		bets := []int{2, 2, 2, 2}
		input := controller.MichiganWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			BoodleBets:   &bets,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		miMock.AssertCalled(t, "Bet", bets)
	})
	t.Run("bet missing", func(t *testing.T) {
		run(t, `{"command":"bet","sessionId":"s1"}`, mustMichiganOutputJSON("param error: boodleBets is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		idx := 0
		input := controller.MichiganWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		miMock.AssertCalled(t, "Play", 0)
	})
	t.Run("play missing", func(t *testing.T) {
		run(t, `{"command":"play","sessionId":"s1"}`, mustMichiganOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustMichiganOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustMichiganOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestMichiganWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("custom config passed through", func(t *testing.T) {
		players, ante, chips, rounds := 5, 12, 500, 20
		expected := domain.MichiganConfig{PlayerCount: 5, Ante: 12, StartingChips: 500, TargetRounds: 20}
		miMock := new(usecase.MockMichiganInteractor)
		miMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMichiganWebController(func() uc.MichiganInteractorIF { return miMock })
		defer ctrl.Stop()

		input := controller.MichiganWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config: &controller.MichiganWebConfig{
				PlayerCount: &players, Ante: &ante, StartingChips: &chips, TargetRounds: &rounds,
			},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		miMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range values fall back to default", func(t *testing.T) {
		players := 99
		expected := domain.DefaultMichiganConfig()
		miMock := new(usecase.MockMichiganInteractor)
		miMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMichiganWebController(func() uc.MichiganInteractorIF { return miMock })
		defer ctrl.Stop()

		input := controller.MichiganWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.MichiganWebConfig{PlayerCount: &players},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		miMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestMichiganWebController_Stop(t *testing.T) {
	miMock := new(usecase.MockMichiganInteractor)
	c := controller.NewMichiganWebController(func() uc.MichiganInteractorIF { return miMock })
	c.Stop()
	c.Stop()
}

// mock2Ints returns the [2,2,2,2] boodle-bet slice used by the bet expectation.
func mock2Ints() []int { return []int{2, 2, 2, 2} }
