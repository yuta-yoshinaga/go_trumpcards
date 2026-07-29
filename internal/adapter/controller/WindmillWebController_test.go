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

func mustWindmillOutputJSON(msg string) string {
	out := &controller.WindmillWebOutput{
		Sails:         []*controller.WebOutputCard{},
		Center:        []*controller.WebOutputCard{},
		Corners:       [][]*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustWindmillOutputJSON: %v", err))
	}
	return string(b)
}

func TestWindmillWebController_Method(t *testing.T) {
	mockOutput := `{"sails":[],"center":[],"corners":[],"stockCount":0,"waste":[],"transferBlocked":false,"phase":0,"moveCount":0,"message":""}`

	wiMock := new(usecase.MockWindmillInteractor)
	wiMock.On("Reset").Return(mockOutput)
	wiMock.On("Draw").Return(mockOutput)
	wiMock.On("GiveUp").Return(mockOutput)
	wiMock.On("Hint").Return(mockOutput)
	wiMock.On("AutoComplete").Return(mockOutput)
	wiMock.On("ActionLog").Return(mockOutput)
	wiMock.On("MoveSailToCenter", 3).Return(mockOutput)
	wiMock.On("MoveSailToCorner", 3, 1).Return(mockOutput)
	wiMock.On("MoveWasteToCenter").Return(mockOutput)
	wiMock.On("MoveWasteToCorner", 2).Return(mockOutput)
	wiMock.On("MoveCornerToCenter", 0).Return(mockOutput)
	wiMock.On("Undo").Return(mockOutput)
	wiMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewWindmillWebController(func() uc.WindmillInteractorIF { return wiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.WindmillWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustWindmillOutputJSON("bye."))
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
		{"sail to center", `{"command":"m","sessionId":"s1","from":{"zone":"sail","col":3},"to":{"zone":"center"}}`},
		{"sail to corner", `{"command":"m","sessionId":"s1","from":{"zone":"sail","col":3},"to":{"zone":"corner","col":1}}`},
		{"waste to center", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"center"}}`},
		{"waste to corner", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"corner","col":2}}`},
		{"corner to center", `{"command":"m","sessionId":"s1","from":{"zone":"corner","col":0},"to":{"zone":"center"}}`},
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
		{"move invalid zones", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"center"}}`},
		// The rescue runs one way only; there is no centre-to-corner move.
		{"center to corner is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"center"},"to":{"zone":"corner","col":0}}`},
		{"corner to corner is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"corner","col":0},"to":{"zone":"corner","col":1}}`},
		{"sail to center missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"sail"},"to":{"zone":"center"}}`},
		{"sail to corner missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"sail","col":3},"to":{"zone":"corner"}}`},
		{"waste to corner missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"corner"}}`},
		{"corner to center missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"corner"},"to":{"zone":"center"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
