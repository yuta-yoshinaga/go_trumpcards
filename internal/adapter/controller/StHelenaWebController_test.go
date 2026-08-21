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

func mustStHelenaOutputJSON(msg string) string {
	out := &controller.StHelenaWebOutput{
		Tableau:       [][]*controller.StHelenaWebOutputTableauCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustStHelenaOutputJSON: %v", err))
	}
	return string(b)
}

func TestStHelenaWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"foundation":[],"redealsRemaining":3,"phase":0,"moveCount":0,"message":""}`
	expectedBody := mockOutput

	ciMock := new(usecase.MockStHelenaInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("Redeal").Return(mockOutput)
	ciMock.On("GiveUp").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("AutoComplete").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)
	ciMock.On("MoveTableauToTableau", 0, 5).Return(mockOutput)
	ciMock.On("MoveTableauToFoundation", 1, 4).Return(mockOutput)
	ciMock.On("Undo").Return(mockOutput)
	ciMock.On("UndoN", 2).Return(mockOutput)

	factory := func() uc.StHelenaInteractorIF { return ciMock }
	ctrl := controller.NewStHelenaWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustStHelenaOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("redeal r", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("redeal full", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"redeal","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo u", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1","n":2}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau","col":5}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to foundation", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation","col":4}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move missing from/to", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"foundation","col":0},"to":{"zone":"tableau","col":1}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing col", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau","col":1}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to foundation missing col", func(t *testing.T) {
		var input controller.StHelenaWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
