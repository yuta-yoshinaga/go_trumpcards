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

func mustSultanOutputJSON(msg string) string {
	out := &controller.SultanWebOutput{
		Foundation:    [][]*controller.WebOutputCard{},
		Divan:         []*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSultanOutputJSON: %v", err))
	}
	return string(b)
}

func TestSultanWebController_Method(t *testing.T) {
	mockOutput := `{"foundation":[],"divan":[],"stockCount":0,"waste":[],"redealCount":0,"canRedeal":false,"phase":0,"moveCount":0,"message":""}`
	expectedBody := mockOutput

	siMock := new(usecase.MockSultanInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Draw").Return(mockOutput)
	siMock.On("Redeal").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("AutoComplete").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("MoveDivanToFoundation", 3).Return(mockOutput)
	siMock.On("MoveWasteToFoundation").Return(mockOutput)
	siMock.On("Undo").Return(mockOutput)
	siMock.On("UndoN", 2).Return(mockOutput)

	factory := func() uc.SultanInteractorIF { return siMock }
	ctrl := controller.NewSultanWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustSultanOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw d", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("redeal rd", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"rd","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo u", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1","n":2}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move divan to foundation", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"divan","divanIdx":3}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to foundation", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"waste"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move missing from", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move divan missing index", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"divan"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zone", func(t *testing.T) {
		var input controller.SultanWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"foundation"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
