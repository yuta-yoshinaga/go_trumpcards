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

func mustAccordionOutputJSON(msg string) string {
	out := &controller.AccordionWebOutput{
		Piles:         []*controller.AccordionWebOutputPile{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustAccordionOutputJSON: %v", err))
	}
	return string(b)
}

func TestAccordionWebController_Method(t *testing.T) {
	mockOutput := `{"piles":[],"pileCount":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"undoToEscape":0,"message":""}`
	expectedBody := mockOutput

	aiMock := new(usecase.MockAccordionInteractor)
	aiMock.On("Reset").Return(mockOutput)
	aiMock.On("GiveUp").Return(mockOutput)
	aiMock.On("Hint").Return(mockOutput)
	aiMock.On("ActionLog").Return(mockOutput)
	aiMock.On("Move", 3, 0).Return(mockOutput)
	aiMock.On("Undo").Return(mockOutput)
	aiMock.On("UndoN", 3).Return(mockOutput)
	aiMock.On("AutoComplete").Return(mockOutput)

	factory := func() uc.AccordionInteractorIF { return aiMock }
	ctrl := controller.NewAccordionWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustAccordionOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1","n":3}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move pile to pile", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"pile","index":3},"to":{"zone":"pile","index":0}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move missing from/to", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"foundation"},"to":{"zone":"pile"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move missing index", func(t *testing.T) {
		var input controller.AccordionWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"pile"},"to":{"zone":"pile","index":0}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

// #5546: 独立 CUI には ac があり、Web GUI にもボタンがあるのに、Web の
// コマンド体系にだけ autocomplete が無く、CLI モードから呼べなかった。
func TestAccordionWebController_AutoComplete(t *testing.T) {
	mockOutput := `{"piles":[],"pileCount":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"undoToEscape":0,"message":""}`

	for _, cmd := range []string{"ac", "autocomplete"} {
		t.Run(cmd, func(t *testing.T) {
			aiMock := new(usecase.MockAccordionInteractor)
			aiMock.On("AutoComplete").Return(mockOutput)
			ctrl := controller.NewAccordionWebController(func() uc.AccordionInteractorIF { return aiMock })
			defer ctrl.Stop()

			var input controller.AccordionWebInput
			_ = json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"s-ac-`+cmd+`"}`), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
			aiMock.AssertCalled(t, "AutoComplete")
		})
	}
}
