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

func mustGrandfathersClockOutputJSON(msg string) string {
	out := &controller.GrandfathersClockWebOutput{
		Tableau:       [][]*controller.GrandfathersClockWebOutputTableauCard{},
		Foundation:    []*controller.GrandfathersClockWebOutputFoundation{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGrandfathersClockOutputJSON: %v", err))
	}
	return string(b)
}

func TestGrandfathersClockWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`

	giMock := new(usecase.MockGrandfathersClockInteractor)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("GiveUp").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("AutoComplete").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)
	giMock.On("MoveTableauToFoundation", 1, 7).Return(mockOutput)
	giMock.On("MoveTableauToTableau", 0, 5).Return(mockOutput)
	giMock.On("Undo").Return(mockOutput)
	giMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewGrandfathersClockWebController(func() uc.GrandfathersClockInteractorIF { return giMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.GrandfathersClockWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustGrandfathersClockOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"autocomplete ac", `{"command":"ac","sessionId":"s1"}`},
		{"undo u", `{"command":"u","sessionId":"s1"}`},
		{"undo_n", `{"command":"undo_n","sessionId":"s1","n":2}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"tableau to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau","col":5}}`},
		{"tableau to clock face", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation","col":7}}`},
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
		{"move from non-tableau zone", `{"command":"m","sessionId":"s1","from":{"zone":"foundation","col":0},"to":{"zone":"tableau","col":5}}`},
		{"move missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau","col":5}}`},
		{"move to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau"}}`},
		// Twelve faces can hold the same suit, so the index cannot be derived.
		{"move to foundation missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"foundation"}}`},
		{"move invalid destination zone", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"stock"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
