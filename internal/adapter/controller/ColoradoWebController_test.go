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

func mustColoradoOutputJSON(msg string) string {
	out := &controller.ColoradoWebOutput{
		Tableau:             [][]*controller.WebOutputCard{},
		Foundation:          [][]*controller.WebOutputCard{},
		FoundationAscending: []bool{},
		Waste:               []*controller.WebOutputCard{},
		WebOutputBase:       controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustColoradoOutputJSON: %v", err))
	}
	return string(b)
}

func TestColoradoWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"foundation":[],"stockCount":0,"waste":[],"phase":0,"moveCount":0,"message":""}`

	ciMock := new(usecase.MockColoradoInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("Draw").Return(mockOutput)
	ciMock.On("GiveUp").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("AutoComplete").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)
	ciMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	ciMock.On("MoveWasteToFoundation").Return(mockOutput)
	ciMock.On("MoveWasteToTableau", 2).Return(mockOutput)
	ciMock.On("MoveStockToTableau", 3).Return(mockOutput)
	ciMock.On("Undo").Return(mockOutput)
	ciMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewColoradoWebController(func() uc.ColoradoInteractorIF { return ciMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.ColoradoWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustColoradoOutputJSON("bye."))
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
		{"tableau to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","idx":1},"to":{"zone":"foundation"}}`},
		{"waste to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`},
		{"waste to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","idx":2}}`},
		// The stock fills a gap directly, without spending a turn on the waste.
		{"stock to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"tableau","idx":3}}`},
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
		{"move invalid zones", `{"command":"m","sessionId":"s1","from":{"zone":"foundation"},"to":{"zone":"tableau","idx":0}}`},
		// A tableau card has exactly one legal destination: a foundation.
		{"tableau to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","idx":0},"to":{"zone":"tableau","idx":5}}`},
		// The stock never reaches a foundation directly.
		{"stock to foundation is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation"}}`},
		{"tableau to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`},
		{"tableau to tableau missing to.idx", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","idx":0},"to":{"zone":"tableau"}}`},
		{"waste to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau"}}`},
		{"stock to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"tableau"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
