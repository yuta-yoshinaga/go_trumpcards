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

func mustEasthavenOutputJSON(msg string) string {
	out := &controller.EasthavenWebOutput{
		Tableau:       [][]*controller.KlondikeWebOutputTableauCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustEasthavenOutputJSON: %v", err))
	}
	return string(b)
}

func TestEasthavenWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`
	expectedBody := mockOutput

	eiMock := new(usecase.MockEasthavenInteractor)
	eiMock.On("Reset").Return(mockOutput)
	eiMock.On("Deal").Return(mockOutput)
	eiMock.On("GiveUp").Return(mockOutput)
	eiMock.On("Hint").Return(mockOutput)
	eiMock.On("AutoComplete").Return(mockOutput)
	eiMock.On("ActionLog").Return(mockOutput)
	eiMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	eiMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	eiMock.On("Undo").Return(mockOutput)
	eiMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.EasthavenInteractorIF { return eiMock }
	ctrl := controller.NewEasthavenWebController(factory)
	defer ctrl.Stop()

	cases := []struct {
		name string
		body string
		code int
		want string
	}{
		{"quit", `{"command":"q","sessionId":"s1"}`, http.StatusOK, mustEasthavenOutputJSON("bye.")},
		{"reset", `{"command":"r","sessionId":"s1"}`, http.StatusOK, expectedBody},
		{"deal", `{"command":"d","sessionId":"s1"}`, http.StatusOK, expectedBody},
		{"giveup", `{"command":"g","sessionId":"s1"}`, http.StatusOK, expectedBody},
		{"autocomplete", `{"command":"ac","sessionId":"s1"}`, http.StatusOK, expectedBody},
		{"undo", `{"command":"u","sessionId":"s1"}`, http.StatusOK, expectedBody},
		{"undo_n", `{"command":"undo_n","sessionId":"s1","n":3}`, http.StatusOK, expectedBody},
		{"undo_n missing n", `{"command":"undo_n","sessionId":"s1"}`, http.StatusBadRequest, ""},
		{"hint", `{"command":"h","sessionId":"s1"}`, http.StatusOK, expectedBody},
		{"log", `{"command":"l","sessionId":"s1"}`, http.StatusOK, expectedBody},
		{"move tt", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":0,"cardIndex":2},"to":{"zone":"tableau","col":4}}`, http.StatusOK, expectedBody},
		{"move tf", `{"command":"m","sessionId":"s1","from":{"zone":"tableau","col":1},"to":{"zone":"foundation"}}`, http.StatusOK, expectedBody},
		{"move missing", `{"command":"m","sessionId":"s1"}`, http.StatusBadRequest, ""},
		{"move invalid zones", `{"command":"m","sessionId":"s1","from":{"zone":"stock"},"to":{"zone":"tableau"}}`, http.StatusBadRequest, ""},
		{"move tt missing params", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau","col":4}}`, http.StatusBadRequest, ""},
		{"move tf missing col", `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`, http.StatusBadRequest, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.EasthavenWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(tc.code)
			if tc.want != "" {
				recorded.BodyIs(tc.want)
			}
		})
	}
}
