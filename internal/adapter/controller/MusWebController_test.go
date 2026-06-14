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

func mustMusOutputJSON(msg string) string {
	out := &controller.MusWebOutput{
		Players:       []*controller.MusWebOutputPlayer{},
		WinnerTeam:    -1,
		HumanTeam:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMusOutputJSON: %v", err))
	}
	return string(b)
}

func TestMusWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockMusInteractor)
	giMock.On("ResetWithConfig", domain.DefaultMusConfig()).Return(mockOutput)
	giMock.On("Mus", true).Return(mockOutput)
	giMock.On("Mus", false).Return(mockOutput)
	giMock.On("Discard", []int{}).Return(mockOutput)
	giMock.On("Discard", []int{0, 2}).Return(mockOutput)
	giMock.On("Bet", domain.MusActionPaso, 0).Return(mockOutput)
	giMock.On("Bet", domain.MusActionEnvido, 4).Return(mockOutput)
	giMock.On("Bet", domain.MusActionOrdago, 0).Return(mockOutput)
	giMock.On("Bet", domain.MusActionQuiero, 0).Return(mockOutput)
	giMock.On("Bet", domain.MusActionNoQuiero, 0).Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.MusInteractorIF { return giMock }
	ctrl := controller.NewMusWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.MusWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustMusOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("mus true", func(t *testing.T) {
		musVal := true
		input := controller.MusWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "mus", SessionID: "s1"},
			Mus:          &musVal,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("mus false", func(t *testing.T) {
		musVal := false
		input := controller.MusWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "mus", SessionID: "s1"},
			Mus:          &musVal,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("mus missing param", func(t *testing.T) {
		run(t, `{"command":"mus","sessionId":"s1"}`, mustMusOutputJSON("param error: mus is required."), http.StatusBadRequest)
	})
	t.Run("discard no indices", func(t *testing.T) {
		run(t, `{"command":"discard","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("discard with indices", func(t *testing.T) {
		input := controller.MusWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "discard", SessionID: "s1"},
			DiscardIndices: []int{0, 2},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bet paso", func(t *testing.T) {
		action := domain.MusActionPaso
		input := controller.MusWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			BetAction:    &action,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bet envido with amount", func(t *testing.T) {
		action := domain.MusActionEnvido
		amount := 4
		input := controller.MusWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			BetAction:    &action,
			BetAmount:    &amount,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bet missing betAction", func(t *testing.T) {
		run(t, `{"command":"bet","sessionId":"s1"}`, mustMusOutputJSON("param error: betAction is required."), http.StatusBadRequest)
	})
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("next alias", func(t *testing.T) {
		run(t, `{"command":"next","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustMusOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustMusOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestMusWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		target := 30
		expected := domain.MusConfig{
			CpuDifficulty:   domain.MusCpuDifficultyHard,
			TargetAmarrakos: 30,
		}
		giMock := new(usecase.MockMusInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMusWebController(func() uc.MusInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.MusWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.MusWebConfig{CpuDifficulty: &diff, TargetAmarrakos: &target},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultMusConfig()
		giMock := new(usecase.MockMusInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMusWebController(func() uc.MusInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.MusWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.MusWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultMusConfig()
		giMock := new(usecase.MockMusInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMusWebController(func() uc.MusInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.MusWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestMusWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockMusInteractor)
	c := controller.NewMusWebController(func() uc.MusInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}
