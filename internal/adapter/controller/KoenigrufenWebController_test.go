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

func mustKoenigrufenOutputJSON(msg string) string {
	out := &controller.KoenigrufenWebOutput{
		Players:         []*controller.KoenigrufenWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		Talon:           []*controller.WebOutputCard{},
		PlayableIndices: []int{},
		DeclarerIdx:     -1,
		HighestBidder:   -1,
		CalledKing:      -1,
		PartnerIdx:      -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustKoenigrufenOutputJSON: %v", err))
	}
	return string(b)
}

func TestKoenigrufenWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockKoenigrufenInteractor)
	diMock.On("ResetWithConfig", domain.DefaultKoenigrufenConfig()).Return(mockOutput)
	diMock.On("Bid", domain.KoenigrufenBidRufer).Return(mockOutput)
	diMock.On("Pass").Return(mockOutput)
	diMock.On("CallKing", 3).Return(mockOutput)
	diMock.On("Discard", []int{0, 1, 2, 3, 4, 5}).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.KoenigrufenInteractorIF { return diMock }
	ctrl := controller.NewKoenigrufenWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.KoenigrufenWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustKoenigrufenOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid", func(t *testing.T) {
		input := controller.KoenigrufenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          func() *string { v := "rufer"; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing bid", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustKoenigrufenOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("pass", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("callking", func(t *testing.T) {
		input := controller.KoenigrufenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "ck", SessionID: "s1"},
			CallSuit:     func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("callking missing suit", func(t *testing.T) {
		run(t, `{"command":"ck","sessionId":"s1"}`, mustKoenigrufenOutputJSON("param error: callSuit is required."), http.StatusBadRequest)
	})
	t.Run("discard", func(t *testing.T) {
		input := controller.KoenigrufenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndices:  []int{0, 1, 2, 3, 4, 5},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("discard missing indices", func(t *testing.T) {
		run(t, `{"command":"d","sessionId":"s1"}`, mustKoenigrufenOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.KoenigrufenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustKoenigrufenOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustKoenigrufenOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
}

func TestKoenigrufenWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		deals := 7
		expected := domain.KoenigrufenConfig{CpuDifficulty: domain.KoenigrufenCpuDifficultyHard, TargetDeals: 7}
		diMock := new(usecase.MockKoenigrufenInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewKoenigrufenWebController(func() uc.KoenigrufenInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.KoenigrufenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.KoenigrufenWebConfig{CpuDifficulty: &diff, TargetDeals: &deals},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultKoenigrufenConfig()
		diMock := new(usecase.MockKoenigrufenInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewKoenigrufenWebController(func() uc.KoenigrufenInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.KoenigrufenWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestKoenigrufenWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockKoenigrufenInteractor)
	c := controller.NewKoenigrufenWebController(func() uc.KoenigrufenInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}
