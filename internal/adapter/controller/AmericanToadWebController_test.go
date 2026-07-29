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

func mustAmericanToadOutputJSON(msg string) string {
	out := &controller.AmericanToadWebOutput{
		Reserve:       []*controller.WebOutputCard{},
		Tableau:       [][]*controller.AmericanToadWebOutputTableauCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustAmericanToadOutputJSON: %v", err))
	}
	return string(b)
}

func TestAmericanToadWebController_Method(t *testing.T) {
	mockOutput := `{"reserve":[],"tableau":[],"foundation":[],"stockCount":0,"waste":[],"baseRank":0,"passesUsed":0,"canRedeal":false,"phase":0,"moveCount":0,"message":""}`

	aiMock := new(usecase.MockAmericanToadInteractor)
	aiMock.On("Reset").Return(mockOutput)
	aiMock.On("Draw").Return(mockOutput)
	aiMock.On("GiveUp").Return(mockOutput)
	aiMock.On("Hint").Return(mockOutput)
	aiMock.On("AutoComplete").Return(mockOutput)
	aiMock.On("ActionLog").Return(mockOutput)
	aiMock.On("MoveReserveToFoundation").Return(mockOutput)
	aiMock.On("MoveReserveToTableau", 3).Return(mockOutput)
	aiMock.On("MoveWasteToFoundation").Return(mockOutput)
	aiMock.On("MoveWasteToTableau", 2).Return(mockOutput)
	aiMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	aiMock.On("MoveTableauToTableau", 0, -1, 5).Return(mockOutput)
	aiMock.On("MoveTableauToTableau", 0, 2, 5).Return(mockOutput)
	aiMock.On("Undo").Return(mockOutput)
	aiMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewAmericanToadWebController(func() uc.AmericanToadInteractorIF { return aiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.AmericanToadWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustAmericanToadOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"draw d", `{"command":"d","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"autocomplete ac", `{"command":"ac","sessionId":"s1"}`},
		{"undo u", `{"command":"u","sessionId":"s1"}`},
		{"undo_n", `{"command":"undo_n","sessionId":"s1","n":2}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"reserve to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"reserve"},"to":{"zone":"foundation"}}`},
		{"reserve to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"reserve"},"to":{"zone":"tableau","col":3}}`},
		{"waste to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`},
		{"waste to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","col":2}}`},
		{"tableau to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation"}}`},
		// cardIndex omitted means "the top card", which the domain reads as -1.
		{"tableau to tableau without cardIndex", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau","col":5}}`},
		{"tableau to tableau with a run head", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0,"cardIndex":2},"to":{"zone":"tableau","col":5}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	for _, tc := range []struct{ name, body string }{
		{"undo_n missing n", `{"command":"undo_n","sessionId":"s1"}`},
		{"move missing from/to", `{"command":"m","sessionId":"s1"}`},
		{"move invalid zones", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation"}}`},
		{"reserve to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"reserve"},"to":{"zone":"tableau"}}`},
		{"waste to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau"}}`},
		{"tableau to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`},
		{"tableau to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
