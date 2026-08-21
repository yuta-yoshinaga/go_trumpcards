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

func mustRankAndFileOutputJSON(msg string) string {
	out := &controller.RankAndFileWebOutput{
		Tableau:       [][]*controller.RankAndFileWebOutputTableauCard{},
		Waste:         []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustRankAndFileOutputJSON: %v", err))
	}
	return string(b)
}

func TestRankAndFileWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`
	expectedBody := mockOutput

	fiMock := new(usecase.MockRankAndFileInteractor)
	fiMock.On("Reset").Return(mockOutput)
	fiMock.On("Draw").Return(mockOutput)
	fiMock.On("GiveUp").Return(mockOutput)
	fiMock.On("Hint").Return(mockOutput)
	fiMock.On("AutoComplete").Return(mockOutput)
	fiMock.On("ActionLog").Return(mockOutput)
	fiMock.On("MoveWasteToTableau", 3).Return(mockOutput)
	fiMock.On("MoveWasteToFoundation").Return(mockOutput)
	fiMock.On("MoveTableauToTableau", 0, 3, 5).Return(mockOutput)
	fiMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	fiMock.On("Undo").Return(mockOutput)
	fiMock.On("UndoN", 2).Return(mockOutput)

	factory := func() uc.RankAndFileInteractorIF { return fiMock }
	ctrl := controller.NewRankAndFileWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustRankAndFileOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw d", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo u", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1","n":2}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo_n","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to tableau", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","col":3}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to foundation", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0,"cardIndex":3},"to":{"zone":"tableau","col":5}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to foundation", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move missing from/to", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"foundation"},"to":{"zone":"waste"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move waste to tableau missing col", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing params", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau","col":5}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to foundation missing col", func(t *testing.T) {
		var input controller.RankAndFileWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}
