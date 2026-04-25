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

func mustSpiteAndMaliceOutputJSON(msg string) string {
	out := &controller.SpiteAndMaliceWebOutput{
		Winner:        -1,
		GoalSize:      domain.SpiteAndMaliceGoalSizeDefault,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSpiteAndMaliceOutputJSON: %v", err))
	}
	return string(b)
}

func TestSpiteAndMaliceWebController_Method(t *testing.T) {
	mockOut := `{"phase":0,"current":0,"players":[{"hand":null,"goalSize":0,"sides":[null,null,null,null],"isCpu":false},{"hand":null,"goalSize":0,"sides":[null,null,null,null],"isCpu":false}],"foundations":[null,null,null,null],"foundationTops":[0,0,0,0],"stockSize":0,"completedSize":0,"moveCount":0,"winner":-1,"goalSize":20,"cpuDifficulty":0,"message":""}`
	expected := mockOut

	m := new(usecase.MockSpiteAndMaliceInteractor)
	m.On("Reset").Return(mockOut)
	m.On("Hint").Return(mockOut)
	m.On("ActionLog").Return(mockOut)
	m.On("CpuStep").Return(mockOut)
	m.On("PlayFromHand", 0, 1).Return(mockOut)
	m.On("PlayFromGoal", 2).Return(mockOut)
	m.On("PlayFromSide", 1, 3).Return(mockOut)
	m.On("Discard", 0, 2).Return(mockOut)

	factory := func() uc.SpiteAndMaliceInteractorIF { return m }
	ctrl := controller.NewSpiteAndMaliceWebController(factory)
	defer ctrl.Stop()

	decode := func(raw string) controller.SpiteAndMaliceWebInput {
		var input controller.SpiteAndMaliceWebInput
		_ = json.Unmarshal([]byte(raw), &input)
		return input
	}

	t.Run("quit", func(t *testing.T) {
		in := decode(`{"command":"q","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustSpiteAndMaliceOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		in := decode(`{"command":"r","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("cpu", func(t *testing.T) {
		in := decode(`{"command":"cpu","sessionId":"s1"}`)
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

	t.Run("move hand->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"hand","idx":0},"to":{"zone":"foundation","idx":1}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move goal->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"goal"},"to":{"zone":"foundation","idx":2}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move side->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"side","idx":1},"to":{"zone":"foundation","idx":3}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move missing from/to", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid to zone", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"hand","idx":0},"to":{"zone":"stock"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move hand missing from.idx", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"hand"},"to":{"zone":"foundation","idx":1}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move side missing from.idx", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"side"},"to":{"zone":"foundation","idx":1}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid from zone", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation","idx":0}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("discard", func(t *testing.T) {
		in := decode(`{"command":"d","sessionId":"s1","from":{"zone":"hand","idx":0},"to":{"zone":"side","idx":2}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("discard missing from/to", func(t *testing.T) {
		in := decode(`{"command":"d","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("discard invalid from zone", func(t *testing.T) {
		in := decode(`{"command":"d","sessionId":"s1","from":{"zone":"goal"},"to":{"zone":"side","idx":0}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("discard invalid to zone", func(t *testing.T) {
		in := decode(`{"command":"d","sessionId":"s1","from":{"zone":"hand","idx":0},"to":{"zone":"foundation","idx":0}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
