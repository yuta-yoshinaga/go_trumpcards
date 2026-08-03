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

func mustNapoleonsSquareOutputJSON(msg string) string {
	out := &controller.NapoleonsSquareWebOutput{
		Tableau:       [][]*controller.NapoleonsSquareWebOutputTableauCard{},
		Waste:         []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustNapoleonsSquareOutputJSON: %v", err))
	}
	return string(b)
}

func TestNapoleonsSquareWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`

	niMock := new(usecase.MockNapoleonsSquareInteractor)
	niMock.On("Reset").Return(mockOutput)
	niMock.On("Draw").Return(mockOutput)
	niMock.On("GiveUp").Return(mockOutput)
	niMock.On("Hint").Return(mockOutput)
	niMock.On("AutoComplete").Return(mockOutput)
	niMock.On("ActionLog").Return(mockOutput)
	niMock.On("MoveWasteToTableau", 3).Return(mockOutput)
	niMock.On("MoveWasteToFoundation").Return(mockOutput)
	niMock.On("MoveTableauToTableau", 0, -1, 5).Return(mockOutput)
	niMock.On("MoveTableauToTableau", 0, 2, 5).Return(mockOutput)
	niMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	niMock.On("Undo").Return(mockOutput)
	niMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewNapoleonsSquareWebController(func() uc.NapoleonsSquareInteractorIF { return niMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.NapoleonsSquareWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustNapoleonsSquareOutputJSON("bye."))
	})

	ok := []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"draw d", `{"command":"d","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"autocomplete ac", `{"command":"ac","sessionId":"s1"}`},
		{"undo u", `{"command":"u","sessionId":"s1"}`},
		{"undo_n", `{"command":"undo_n","sessionId":"s1","n":2}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"waste to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","col":3}}`},
		{"waste to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`},
		{"tableau to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation"}}`},
		// cardIndex omitted means "the top card", which the domain reads as -1.
		{"tableau to tableau without cardIndex", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau","col":5}}`},
		{"tableau to tableau with a run head", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0,"cardIndex":2},"to":{"zone":"tableau","col":5}}`},
	}
	for _, tc := range ok {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	bad := []struct{ name, body string }{
		{"undo_n missing n", `{"command":"undo_n","sessionId":"s1"}`},
		{"move missing from/to", `{"command":"m","sessionId":"s1"}`},
		{"move invalid zones", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation"}}`},
		{"waste to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau"}}`},
		{"tableau to tableau missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau","col":5}}`},
		{"tableau to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau"}}`},
		{"tableau to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
