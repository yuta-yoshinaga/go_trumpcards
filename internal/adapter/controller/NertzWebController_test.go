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

func mustNertzOutputJSON(msg string) string {
	out := &controller.NertzWebOutput{
		WinnerIdx:     -1,
		MatchWinner:   -1,
		PlayerCount:   domain.NertzPlayerCntDefault,
		DrawCount:     3,
		TargetScore:   domain.NertzTargetScoreDefault,
		CpuDifficulty: int(domain.NertzCpuDifficultyNormal),
		Players:       make([]*controller.NertzWebPlayer, 0),
		Foundations:   make([]*controller.NertzWebFoundation, 0),
		// 上限は盤面が無くても規則なので、既定の応答にも乗る (#5578)。
		FoundationMax: domain.NertzFoundationMax,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustNertzOutputJSON: %v", err))
	}
	return string(b)
}

func TestNertzWebController_Method(t *testing.T) {
	mockOut := mustNertzOutputJSON("")
	expected := mockOut

	m := new(usecase.MockNertzInteractor)
	m.On("Reset").Return(mockOut)
	m.On("ResetWithConfig", domain.NertzConfig{
		PlayerCount:   3,
		DrawCount:     1,
		TargetScore:   75,
		CpuDifficulty: domain.NertzCpuDifficultyHard,
		CpuTickMoves:  4,
	}).Return(mockOut)
	m.On("NextRound").Return(mockOut)
	m.On("Tick").Return(mockOut)
	m.On("Hint").Return(mockOut)
	m.On("ActionLog").Return(mockOut)
	m.On("Undo").Return(mockOut)
	m.On("Draw", 0).Return(mockOut)
	m.On("MoveNertzToFoundation", 0, 1).Return(mockOut)
	m.On("MoveNertzToTableau", 0, 2).Return(mockOut)
	m.On("MoveWasteToFoundation", 0, 0).Return(mockOut)
	m.On("MoveWasteToTableau", 0, 3).Return(mockOut)
	m.On("MoveTableauToFoundation", 0, 1, 2).Return(mockOut)
	m.On("MoveTableauToTableau", 0, 0, 1, 2).Return(mockOut)
	m.On("GetConfig").Return(domain.DefaultNertzConfig())

	factory := func() uc.NertzInteractorIF { return m }
	ctrl := controller.NewNertzWebController(factory)
	defer ctrl.Stop()

	decode := func(raw string) controller.NertzWebInput {
		var input controller.NertzWebInput
		_ = json.Unmarshal([]byte(raw), &input)
		return input
	}

	t.Run("quit", func(t *testing.T) {
		in := decode(`{"command":"q","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustNertzOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		in := decode(`{"command":"r","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("reset with config", func(t *testing.T) {
		in := decode(`{"command":"r","sessionId":"s1","config":{"playerCount":3,"drawCount":1,"targetScore":75,"cpuDifficulty":2,"cpuTickMoves":4}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("next round", func(t *testing.T) {
		in := decode(`{"command":"nr","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	// Issue #1532: long-form alias is lowercase to match Hearts / Skat / Whist.
	t.Run("next round (lowercase alias)", func(t *testing.T) {
		in := decode(`{"command":"nextround","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("tick", func(t *testing.T) {
		in := decode(`{"command":"tick","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("draw", func(t *testing.T) {
		in := decode(`{"command":"d","sessionId":"s1","playerIdx":0}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("hint", func(t *testing.T) {
		in := decode(`{"command":"h","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("log", func(t *testing.T) {
		in := decode(`{"command":"l","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("undo", func(t *testing.T) {
		in := decode(`{"command":"u","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move nertz->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"nertz"},"to":{"zone":"foundation","idx":1}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move nertz->tableau", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"nertz"},"to":{"zone":"tableau","col":2}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move waste->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation","idx":0}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move waste->tableau", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","col":3}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move tableau->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation","idx":2}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move tableau->tableau", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0,"cardIndex":1},"to":{"zone":"tableau","col":2}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move missing from/to", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move nertz->foundation missing idx", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"nertz"},"to":{"zone":"foundation"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move nertz->tableau missing col", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"nertz"},"to":{"zone":"tableau"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move waste->foundation missing idx", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move waste->tableau missing col", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau->foundation missing fields", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation","idx":0}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau->tableau missing fields", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau","col":1}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation","idx":0}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
