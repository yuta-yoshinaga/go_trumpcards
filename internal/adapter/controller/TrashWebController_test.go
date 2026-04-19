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

func mustTrashOutputJSON(msg string) string {
	out := &controller.TrashWebOutput{
		Winner:        -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTrashOutputJSON: %v", err))
	}
	return string(b)
}

func TestTrashWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0,"current":0,"players":[{"slots":[{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false}],"isCpu":false},{"slots":[{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false},{"faceUp":false}],"isCpu":true}],"stockSize":34,"discardSize":0,"moveCount":0,"winner":-1,"message":""}`
	expectedBody := mockOutput

	tiMock := new(usecase.MockTrashInteractor)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Draw").Return(mockOutput)
	tiMock.On("PlaceWild", 3).Return(mockOutput)
	tiMock.On("CpuStep").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TrashInteractorIF { return tiMock }
	ctrl := controller.NewTrashWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustTrashOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("place", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1","position":3}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("place missing position", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("cpu", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"cpu","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.TrashWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
