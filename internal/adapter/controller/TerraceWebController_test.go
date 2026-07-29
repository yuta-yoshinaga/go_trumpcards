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

func mustTerraceOutputJSON(msg string) string {
	out := &controller.TerraceWebOutput{
		Reserve:       []*controller.WebOutputCard{},
		Tableau:       [][]*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTerraceOutputJSON: %v", err))
	}
	return string(b)
}

func TestTerraceWebController_Method(t *testing.T) {
	mockOutput := `{"reserve":[],"tableau":[],"foundation":[],"stockCount":0,"waste":[],"baseRank":0,"awaitingBaseRank":false,"phase":0,"moveCount":0,"message":""}`

	tiMock := new(usecase.MockTerraceInteractor)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Draw").Return(mockOutput)
	tiMock.On("GiveUp").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("AutoComplete").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)
	tiMock.On("MoveReserveToFoundation").Return(mockOutput)
	tiMock.On("MoveWasteToFoundation").Return(mockOutput)
	tiMock.On("MoveWasteToTableau", 2).Return(mockOutput)
	tiMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	tiMock.On("MoveTableauToTableau", 0, 5).Return(mockOutput)
	tiMock.On("Undo").Return(mockOutput)
	tiMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewTerraceWebController(func() uc.TerraceInteractorIF { return tiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.TerraceWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustTerraceOutputJSON("bye."))
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
		{"terrace to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"reserve"},"to":{"zone":"foundation"}}`},
		{"waste to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`},
		{"waste to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","col":2}}`},
		{"tableau to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation"}}`},
		{"tableau to tableau", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau","col":5}}`},
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
		// The terrace feeds the foundations and nothing else.
		{"terrace to tableau is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"reserve"},"to":{"zone":"tableau","col":0}}`},
		{"waste to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau"}}`},
		{"tableau to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`},
		{"tableau to tableau missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0},"to":{"zone":"tableau"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
