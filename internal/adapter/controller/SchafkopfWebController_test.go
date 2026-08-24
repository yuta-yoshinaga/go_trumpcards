//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustSchafkopfOutputJSON(msg string) string {
	out := &controller.SchafkopfWebOutput{
		Players:         []*controller.SchafkopfWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		CallableSuits:   []int{},
		PlayableIndices: []int{},
		PickerIdx:       -1,
		PartnerIdx:      -1,
		WinnerIdx:       -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSchafkopfOutputJSON: %v", err))
	}
	return string(b)
}

func TestSchafkopfWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockSchafkopfInteractor)
	giMock.On("ResetWithConfig", domain.DefaultSchafkopfConfig()).Return(mockOutput)
	giMock.On("Declare", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
	giMock.On("Call", 2).Return(mockOutput)
	giMock.On("Play", 3).Return(mockOutput)
	giMock.On("NextTrick").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.SchafkopfInteractorIF { return giMock }
	ctrl := controller.NewSchafkopfWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.SchafkopfWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustSchafkopfOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("pick true defaults to Rufspiel", func(t *testing.T) {
		pickVal := true
		input := controller.SchafkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "pick", SessionID: "s1"},
			Pick:         &pickVal,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		giMock.AssertCalled(t, "Declare", true, domain.SchafkopfContractRufspiel, 0)
	})
	// **契約とスートは素通しでなければならない。**片方でも落ちると、宣言した
	// 契約と盤面の切り札が食い違う。
	t.Run("pick forwards the contract and its solo suit", func(t *testing.T) {
		pickVal := true
		contract := int(domain.SchafkopfContractSolo)
		soloSuit := domain.CardDesignHeart
		input := controller.SchafkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "pick", SessionID: "s1"},
			Pick:         &pickVal,
			Contract:     &contract,
			SoloSuit:     &soloSuit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		giMock.AssertCalled(t, "Declare", true, domain.SchafkopfContractSolo, domain.CardDesignHeart)
	})
	t.Run("pick missing param", func(t *testing.T) {
		run(t, `{"command":"pick","sessionId":"s1"}`, mustSchafkopfOutputJSON("param error: pick is required."), http.StatusBadRequest)
	})
	t.Run("call", func(t *testing.T) {
		suit := 2
		input := controller.SchafkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "call", SessionID: "s1"},
			CallSuit:     &suit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("call missing param", func(t *testing.T) {
		run(t, `{"command":"call","sessionId":"s1"}`, mustSchafkopfOutputJSON("param error: callSuit is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.SchafkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustSchafkopfOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustSchafkopfOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustSchafkopfOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestSchafkopfWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		base := 3
		expected := domain.SchafkopfConfig{
			CpuDifficulty: domain.SchafkopfCpuDifficultyHard,
			BaseChips:     3,
			StartChips:    20,
			TargetChips:   40,
		}
		giMock := new(usecase.MockSchafkopfInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewSchafkopfWebController(func() uc.SchafkopfInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.SchafkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.SchafkopfWebConfig{CpuDifficulty: &diff, BaseChips: &base},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultSchafkopfConfig()
		giMock := new(usecase.MockSchafkopfInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewSchafkopfWebController(func() uc.SchafkopfInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.SchafkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.SchafkopfWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultSchafkopfConfig()
		giMock := new(usecase.MockSchafkopfInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewSchafkopfWebController(func() uc.SchafkopfInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.SchafkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestSchafkopfWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockSchafkopfInteractor)
	c := controller.NewSchafkopfWebController(func() uc.SchafkopfInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}
