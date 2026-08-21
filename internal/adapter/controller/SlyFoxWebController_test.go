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

func mustSlyFoxOutputJSON(msg string) string {
	out := &controller.SlyFoxWebOutput{
		Tableau:             [][]*controller.WebOutputCard{},
		Foundation:          [][]*controller.WebOutputCard{},
		FoundationAscending: []bool{},
		DealCycle:           20,
		WebOutputBase:       controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSlyFoxOutputJSON: %v", err))
	}
	return string(b)
}

func TestSlyFoxWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"foundation":[],"stockCount":0,"dealtThisCycle":0,"dealCycle":20,"reserveLocked":false,"phase":0,"moveCount":0,"message":""}`

	ciMock := new(usecase.MockSlyFoxInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("DealToPile", 4).Return(mockOutput)
	ciMock.On("DealToFoundation", 1).Return(mockOutput)
	ciMock.On("GiveUp").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("AutoComplete").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)
	ciMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	ciMock.On("Undo").Return(mockOutput)
	ciMock.On("UndoN", 2).Return(mockOutput)

	ctrl := controller.NewSlyFoxWebController(func() uc.SlyFoxInteractorIF { return ciMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.SlyFoxWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustSlyFoxOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"deal onto a slot", `{"command":"d","sessionId":"s1","to":{"zone":"tableau","idx":4}}`},
		{"deal to a foundation", `{"command":"d","sessionId":"s1","to":{"zone":"foundation","idx":1}}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"autocomplete ac", `{"command":"ac","sessionId":"s1"}`},
		{"undo u", `{"command":"u","sessionId":"s1"}`},
		{"undo_n", `{"command":"undo_n","sessionId":"s1","n":2}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"tableau to foundation", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","idx":1},"to":{"zone":"foundation"}}`},
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
		// **捨て札も山札も移動元ではない。**クローン元のコロラドから引き継いだ
		// ゾーンをリネームだけして残すと、ヘルプに無い構文が半分動く。
		{"waste is not a zone", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`},
		{"waste to tableau is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","idx":2}}`},
		{"stock is not a move source", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"tableau","idx":3}}`},
		{"stock to foundation is not a move", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"foundation"}}`},
		// 配り先は必須。捨て札が無いので、行き先を決めずには配れない。
		{"deal without a destination", `{"command":"d","sessionId":"s1"}`},
		{"deal without an index", `{"command":"d","sessionId":"s1","to":{"zone":"tableau"}}`},
		{"tableau to foundation missing from.col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`},
		{"tableau to tableau missing to.idx", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","idx":0},"to":{"zone":"tableau"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
