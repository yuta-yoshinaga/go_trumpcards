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

func mustSirTommyOutputJSON(msg string) string {
	out := &controller.SirTommyWebOutput{
		Foundations:   [][]*controller.WebOutputCard{},
		Wastes:        [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSirTommyOutputJSON: %v", err))
	}
	return string(b)
}

func TestSirTommyWebController_Method(t *testing.T) {
	mockOut := `{"foundations":[],"wastes":[],"stockCount":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"undoToEscape":0,"message":""}`
	expected := mockOut

	m := new(usecase.MockSirTommyInteractor)
	m.On("Reset").Return(mockOut)
	m.On("GiveUp").Return(mockOut)
	m.On("Hint").Return(mockOut)
	m.On("ActionLog").Return(mockOut)
	m.On("AutoComplete").Return(mockOut)
	m.On("Undo").Return(mockOut)
	m.On("UndoN", 2).Return(mockOut)
	m.On("PlayStockToFoundation", 1).Return(mockOut)
	m.On("PlayStockToWaste", 0).Return(mockOut)
	m.On("PlayWasteToFoundation", 2, 1).Return(mockOut)

	factory := func() uc.SirTommyInteractorIF { return m }
	ctrl := controller.NewSirTommyWebController(factory)
	defer ctrl.Stop()

	decode := func(raw string) controller.SirTommyWebInput {
		var input controller.SirTommyWebInput
		_ = json.Unmarshal([]byte(raw), &input)
		return input
	}

	t.Run("quit", func(t *testing.T) {
		in := decode(`{"command":"q","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustSirTommyOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		in := decode(`{"command":"r","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("giveup", func(t *testing.T) {
		in := decode(`{"command":"g","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("autocomplete", func(t *testing.T) {
		in := decode(`{"command":"ac","sessionId":"s1"}`)
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

	t.Run("undo_n", func(t *testing.T) {
		in := decode(`{"command":"undo_n","sessionId":"s1","n":2}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		in := decode(`{"command":"undo_n","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
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

	t.Run("move stock->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation","idx":1}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move stock->waste", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"waste","idx":0}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move waste->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"waste","idx":2},"to":{"zone":"foundation","idx":1}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expected)
	})

	t.Run("move missing from/to", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1"}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move missing to.idx stock->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move missing to.idx stock->waste", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"waste"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move missing idx waste->foundation", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		in := decode(`{"command":"m","sessionId":"s1","from":{"zone":"foundation"},"to":{"zone":"stock"}}`)
		recorded := execRequest(t, ctrl.Exec, &in)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
