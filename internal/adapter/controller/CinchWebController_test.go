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

func mustCinchOutputJSON(msg string) string {
	out := &controller.CinchWebOutput{
		Players:         []*controller.CinchWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		LastTrick:       []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		RoundWinners:    []int{},
		TotalTricks:     domain.CinchTotalTricks,
		LastTrickWinner: -1,
		BidWinnerIdx:    -1,
		WinnerIdx:       -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCinchOutputJSON: %v", err))
	}
	return string(b)
}

func TestCinchWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	ciMock := new(usecase.MockCinchInteractor)
	ciMock.On("ResetWithConfig", domain.DefaultCinchConfig()).Return(mockOutput)
	ciMock.On("Bid", 3).Return(mockOutput)
	ciMock.On("NameTrump", domain.CardDesignHeart).Return(mockOutput)
	ciMock.On("Play", 3).Return(mockOutput)
	ciMock.On("NextRound").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.CinchInteractorIF { return ciMock }
	ctrl := controller.NewCinchWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.CinchWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustCinchOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid", func(t *testing.T) {
		input := controller.CinchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustCinchOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("trump", func(t *testing.T) {
		input := controller.CinchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "t", SessionID: "s1"},
			TrumpSuit:    func() *int { v := domain.CardDesignHeart; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("trump missing", func(t *testing.T) {
		run(t, `{"command":"t","sessionId":"s1"}`, mustCinchOutputJSON("param error: trumpSuit is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.CinchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustCinchOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("next alias", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustCinchOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustCinchOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestCinchWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		pts := 15
		expected := domain.CinchConfig{CpuDifficulty: domain.CinchDifficultyHard, PointLimit: 15}
		ciMock := new(usecase.MockCinchInteractor)
		ciMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewCinchWebController(func() uc.CinchInteractorIF { return ciMock })
		defer ctrl.Stop()

		input := controller.CinchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.CinchWebConfig{CpuDifficulty: &diff, PointLimit: &pts},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		ciMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultCinchConfig()
		ciMock := new(usecase.MockCinchInteractor)
		ciMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewCinchWebController(func() uc.CinchInteractorIF { return ciMock })
		defer ctrl.Stop()

		input := controller.CinchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.CinchWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		ciMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultCinchConfig()
		ciMock := new(usecase.MockCinchInteractor)
		ciMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewCinchWebController(func() uc.CinchInteractorIF { return ciMock })
		defer ctrl.Stop()

		input := controller.CinchWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		ciMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestCinchWebController_Stop(t *testing.T) {
	ciMock := new(usecase.MockCinchInteractor)
	c := controller.NewCinchWebController(func() uc.CinchInteractorIF { return ciMock })
	c.Stop()
	c.Stop()
}
