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

func mustCourtPieceOutputJSON(msg string) string {
	out := &controller.CourtPieceWebOutput{
		Players:       []*controller.CourtPieceWebOutputPlayer{},
		TeamScores:    []int{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCourtPieceOutputJSON: %v", err))
	}
	return string(b)
}

func TestCourtPieceWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	tiMock := new(usecase.MockCourtPieceInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultCourtPieceConfig()).Return(mockOutput)
	tiMock.On("DeclareTrump", domain.CardDesignSpade).Return(mockOutput)
	tiMock.On("Play", 3).Return(mockOutput)
	tiMock.On("NextTrick").Return(mockOutput)
	tiMock.On("NextRound").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.CourtPieceInteractorIF { return tiMock }
	ctrl := controller.NewCourtPieceWebController(factory)
	defer ctrl.Stop()

	t.Run("q quits", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"cp1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mustCourtPieceOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"cp1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("trump valid", func(t *testing.T) {
		suit := domain.CardDesignSpade
		input := controller.CourtPieceWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "trump", SessionID: "cp1"},
			TrumpSuit:    &suit,
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("play valid", func(t *testing.T) {
		idx := 3
		input := controller.CourtPieceWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "cp1"},
			CardIndex:    &idx,
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("next n", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"cp1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"cp1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"cp1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"cp1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"cp1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCourtPieceOutputJSON("Unsupported command."))
	})

	t.Run("trump missing", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"trump","sessionId":"cp1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCourtPieceOutputJSON("param error: trumpSuit is required."))
	})

	t.Run("play missing cardIndex", func(t *testing.T) {
		var input controller.CourtPieceWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"cp1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCourtPieceOutputJSON("param error: cardIndex is required."))
	})

	t.Run("session id too long", func(t *testing.T) {
		input := controller.CourtPieceWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCourtPieceOutputJSON("param error."))
	})
}

func TestCourtPieceWebController_CustomConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config", func(t *testing.T) {
		diff := 2
		limit := 9
		expected := domain.CourtPieceConfig{
			CpuDifficulty: domain.CourtPieceCpuDifficultyHard,
			PointLimit:    9,
		}
		tiMock := new(usecase.MockCourtPieceInteractor)
		tiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CourtPieceInteractorIF { return tiMock }
		ctrl := controller.NewCourtPieceWebController(factory)
		defer ctrl.Stop()

		input := controller.CourtPieceWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cp-cfg-1"},
			Config: &controller.CourtPieceWebConfig{
				CpuDifficulty: &diff,
				PointLimit:    &limit,
			},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		tiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		tiMock := new(usecase.MockCourtPieceInteractor)
		tiMock.On("ResetWithConfig", domain.DefaultCourtPieceConfig()).Return(mockOutput)

		factory := func() uc.CourtPieceInteractorIF { return tiMock }
		ctrl := controller.NewCourtPieceWebController(factory)
		defer ctrl.Stop()

		input := controller.CourtPieceWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cp-cfg-2"},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		tiMock.AssertCalled(t, "ResetWithConfig", domain.DefaultCourtPieceConfig())
	})
}
