//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustWaspOutputJSON(msg string) string {
	out := &controller.WaspWebOutput{
		Tableau:       [][]*controller.KlondikeWebOutputTableauCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustWaspOutputJSON: %v", err))
	}
	return string(b)
}

func TestWaspWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"phase":0,"moveCount":0,"message":""}`
	expectedBody := mockOutput

	siMock := new(usecase.MockWaspInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Deal").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("AutoComplete").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	siMock.On("Undo").Return(mockOutput)
	siMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.WaspInteractorIF { return siMock }
	ctrl := controller.NewWaspWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustWaspOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("deal", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1","n":3}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0,"cardIndex":2},"to":{"zone":"tableau","col":4}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move missing from/to", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"foundation"},"to":{"zone":"tableau"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing params", func(t *testing.T) {
		var input controller.WaspWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau","col":4}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
