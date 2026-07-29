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

func mustBraidOutputJSON(msg string) string {
	out := &controller.BraidWebOutput{
		Braid:         []*controller.WebOutputCard{},
		Fields:        []*controller.WebOutputCard{},
		Helpers:       []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBraidOutputJSON: %v", err))
	}
	return string(b)
}

func TestBraidWebController_Method(t *testing.T) {
	mockOutput := `{"braid":[],"fields":[],"helpers":[],"foundation":[],"stockCount":0,"waste":[],"baseRank":0,"direction":0,"awaitingDirection":false,"redealsLeft":2,"canRedeal":false,"phase":0,"moveCount":0,"message":""}`

	biMock := new(usecase.MockBraidInteractor)
	biMock.On("Reset").Return(mockOutput)
	biMock.On("Draw").Return(mockOutput)
	biMock.On("GiveUp").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("AutoComplete").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)
	biMock.On("ChooseDirection", true).Return(mockOutput)
	biMock.On("ChooseDirection", false).Return(mockOutput)
	biMock.On("MoveBraidToFoundation").Return(mockOutput)
	biMock.On("MoveFieldToFoundation", 2).Return(mockOutput)
	biMock.On("MoveHelperToFoundation", 5).Return(mockOutput)
	biMock.On("MoveWasteToFoundation").Return(mockOutput)
	biMock.On("MoveWasteToHelper", 3).Return(mockOutput)
	biMock.On("Undo").Return(mockOutput)
	biMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewBraidWebController(func() uc.BraidInteractorIF { return biMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.BraidWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustBraidOutputJSON("bye."))
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
		{"direction up", `{"command":"dir","sessionId":"s1","ascending":true}`},
		{"direction down", `{"command":"direction","sessionId":"s1","ascending":false}`},
		{"braid to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"braid"},"to":{"zone":"foundation"}}`},
		{"field to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"field","col":2},"to":{"zone":"foundation"}}`},
		{"helper to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"helper","col":5},"to":{"zone":"foundation"}}`},
		{"waste to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`},
		{"waste to helper", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"helper","col":3}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	for _, tc := range []struct{ name, body string }{
		{"undo_n missing n", `{"command":"undo_n","sessionId":"s1"}`},
		{"direction missing ascending", `{"command":"dir","sessionId":"s1"}`},
		{"move missing from/to", `{"command":"m","sessionId":"s1"}`},
		{"move invalid zones", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation"}}`},
		// 枠同士の移動は存在しない。行き先は基礎札か、捨て札からのヘルパーだけ。
		{"field to helper is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"field","col":0},"to":{"zone":"helper","col":1}}`},
		{"braid to helper is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"braid"},"to":{"zone":"helper","col":0}}`},
		{"field to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"field"},"to":{"zone":"foundation"}}`},
		{"helper to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"helper"},"to":{"zone":"foundation"}}`},
		{"waste to helper missing to.col", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"helper"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
