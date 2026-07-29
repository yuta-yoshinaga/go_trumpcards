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

func mustMissMilliganOutputJSON(msg string) string {
	out := &controller.MissMilliganWebOutput{
		Tableau:       [][]*controller.MissMilliganWebOutputTableauCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		Waived:        []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMissMilliganOutputJSON: %v", err))
	}
	return string(b)
}

func TestMissMilliganWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"foundation":[],"waived":[],"canWaive":false,"phase":0,"moveCount":0,"message":""}`

	miMock := new(usecase.MockMissMilliganInteractor)
	miMock.On("Reset").Return(mockOutput)
	miMock.On("Deal").Return(mockOutput)
	miMock.On("GiveUp").Return(mockOutput)
	miMock.On("Hint").Return(mockOutput)
	miMock.On("AutoComplete").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)
	miMock.On("MoveTableauToTableau", 0, -1, 5).Return(mockOutput)
	miMock.On("MoveTableauToTableau", 0, 2, 5).Return(mockOutput)
	miMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	miMock.On("Waive", 3, -1).Return(mockOutput)
	miMock.On("Waive", 3, 1).Return(mockOutput)
	miMock.On("PlaceWaived", 4).Return(mockOutput)
	miMock.On("MoveWaivedToFoundation").Return(mockOutput)
	miMock.On("Undo").Return(mockOutput)
	miMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewMissMilliganWebController(func() uc.MissMilliganInteractorIF { return miMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.MissMilliganWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustMissMilliganOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"deal d", `{"command":"d","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"autocomplete ac", `{"command":"ac","sessionId":"s1"}`},
		{"undo u", `{"command":"u","sessionId":"s1"}`},
		{"undo_n", `{"command":"undo_n","sessionId":"s1","n":2}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		// cardIndex omitted means "the top card", which the domain reads as -1.
		{"waive without index", `{"command":"wv","sessionId":"s1","from":{"zone":"tableau","col":3}}`},
		{"waive with a run head", `{"command":"waive","sessionId":"s1","from":{"zone":"tableau","col":3,"cardIndex":1}}`},
		{"tableau to tableau without cardIndex", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau","col":5}}`},
		{"tableau to tableau with a run head", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0,"cardIndex":2},"to":{"zone":"tableau","col":5}}`},
		{"tableau to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation"}}`},
		{"waived back to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"waived"},"to":{"zone":"tableau","col":4}}`},
		{"waived to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"waived"},"to":{"zone":"foundation"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	for _, tc := range []struct{ name, body string }{
		{"undo_n missing n", `{"command":"undo_n","sessionId":"s1"}`},
		{"waive missing col", `{"command":"wv","sessionId":"s1"}`},
		{"move missing from/to", `{"command":"m","sessionId":"s1"}`},
		{"move invalid zones", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation"}}`},
		{"tableau to tableau missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau","col":5}}`},
		{"tableau to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau"}}`},
		{"tableau to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`},
		{"waived to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"waived"},"to":{"zone":"tableau"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
