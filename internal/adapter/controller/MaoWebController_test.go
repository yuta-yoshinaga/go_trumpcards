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

func mustMaoOutputJSON(msg string) string {
	out := &controller.MaoWebOutput{
		Players:       []*controller.MaoWebOutputPlayer{},
		WinnerIdx:     -1,
		Direction:     1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMaoOutputJSON: %v", err))
	}
	return string(b)
}

func TestMaoWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`

	siMock := new(usecase.MockMaoInteractor)
	siMock.On("ResetWithConfig", domain.DefaultMaoConfig()).Return(mockOutput)
	siMock.On("Play", 3).Return(mockOutput)
	siMock.On("ChooseSuit", 2).Return(mockOutput)
	siMock.On("Draw").Return(mockOutput)
	siMock.On("Declare").Return(mockOutput)
	siMock.On("SkipDeclare").Return(mockOutput)
	siMock.On("DeclareWord", "spade").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewMaoWebController(func() uc.MaoInteractorIF { return siMock })
	defer ctrl.Stop()

	cases := []string{
		`{"command":"r","sessionId":"s1"}`,
		`{"command":"d","sessionId":"s1"}`,
		`{"command":"dc","sessionId":"s1"}`,
		`{"command":"sk","sessionId":"s1"}`,
		`{"command":"nr","sessionId":"s1"}`,
		`{"command":"log","sessionId":"s1"}`,
	}
	for _, body := range cases {
		var input controller.MaoWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	}

	t.Run("play with cardIndex", func(t *testing.T) {
		idx := 3
		input := controller.MaoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "s1"},
			CardIndex:    &idx,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("suit with suit param", func(t *testing.T) {
		s := 2
		input := controller.MaoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "suit", SessionID: "s1"},
			Suit:         &s,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("declareword with word param", func(t *testing.T) {
		w := "spade"
		input := controller.MaoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "declareword", SessionID: "s1"},
			Word:         &w,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing cardIndex", func(t *testing.T) {
		var input controller.MaoWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMaoOutputJSON("param error: cardIndex is required."))
	})

	t.Run("suit missing suit", func(t *testing.T) {
		var input controller.MaoWebInput
		_ = json.Unmarshal([]byte(`{"command":"suit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMaoOutputJSON("param error: suit is required."))
	})

	t.Run("declareword missing word", func(t *testing.T) {
		var input controller.MaoWebInput
		_ = json.Unmarshal([]byte(`{"command":"dw","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMaoOutputJSON("param error: word is required."))
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.MaoWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMaoOutputJSON("Unsupported command."))
	})
}

func TestMaoWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	diff := 2
	limit := 300
	expected := domain.MaoConfig{CpuDifficulty: domain.MaoCpuDifficultyHard, PointLimit: 300}
	siMock := new(usecase.MockMaoInteractor)
	siMock.On("ResetWithConfig", expected).Return(mockOutput)

	ctrl := controller.NewMaoWebController(func() uc.MaoInteractorIF { return siMock })
	defer ctrl.Stop()

	input := controller.MaoWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
		Config:       &controller.MaoWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	siMock.AssertCalled(t, "ResetWithConfig", expected)
}

func TestMaoWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockMaoInteractor)
	c := controller.NewMaoWebController(func() uc.MaoInteractorIF { return siMock })
	c.Stop()
	c.Stop()
}
