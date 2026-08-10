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

func mustFourSeasonsOutputJSON(msg string) string {
	out := &controller.FourSeasonsWebOutput{
		Tableau:       [][]*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustFourSeasonsOutputJSON: %v", err))
	}
	return string(b)
}

func TestFourSeasonsWebController_Method(t *testing.T) {
	mockOut := `{"tableau":[],"foundation":[],"stockCount":0,"waste":[],"baseRank":7,"phase":0,"moveCount":0,"canUndo":false,"message":""}`
	expected := mockOut

	m := new(usecase.MockFourSeasonsInteractor)
	m.On("Reset").Return(mockOut)
	m.On("Draw").Return(mockOut)
	m.On("GiveUp").Return(mockOut)
	m.On("Hint").Return(mockOut)
	m.On("ActionLog").Return(mockOut)
	m.On("AutoComplete").Return(mockOut)
	m.On("Undo").Return(mockOut)
	m.On("UndoN", 2).Return(mockOut)
	m.On("MoveWasteToTableau", 3).Return(mockOut)
	m.On("MoveWasteToFoundation", 1).Return(mockOut)
	m.On("MoveTableauToTableau", 0, 2).Return(mockOut)
	m.On("MoveTableauToFoundation", 4, 0).Return(mockOut)

	ctrl := controller.NewFourSeasonsWebController(func() uc.FourSeasonsInteractorIF { return m })
	defer ctrl.Stop()

	decode := func(raw string) controller.FourSeasonsWebInput {
		var input controller.FourSeasonsWebInput
		_ = json.Unmarshal([]byte(raw), &input)
		return input
	}

	okCases := map[string]string{
		"quit":                `{"command":"q","sessionId":"s1"}`,
		"reset":               `{"command":"r","sessionId":"s1"}`,
		"draw":                `{"command":"d","sessionId":"s1"}`,
		"draw long":           `{"command":"draw","sessionId":"s1"}`,
		"giveup":              `{"command":"g","sessionId":"s1"}`,
		"autocomplete":        `{"command":"ac","sessionId":"s1"}`,
		"undo":                `{"command":"u","sessionId":"s1"}`,
		"undo_n":              `{"command":"undo_n","sessionId":"s1","n":2}`,
		"hint":                `{"command":"h","sessionId":"s1"}`,
		"log":                 `{"command":"l","sessionId":"s1"}`,
		"waste->tableau":      `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau","idx":3}}`,
		"waste->foundation":   `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation","idx":1}}`,
		"tableau->tableau":    `{"command":"m","sessionId":"s1","from":{"zone":"tableau","idx":0},"to":{"zone":"tableau","idx":2}}`,
		"tableau->foundation": `{"command":"m","sessionId":"s1","from":{"zone":"tableau","idx":4},"to":{"zone":"foundation","idx":0}}`,
	}
	for name, raw := range okCases {
		t.Run(name, func(t *testing.T) {
			in := decode(raw)
			recorded := execRequest(t, ctrl.Exec, &in)
			recorded.CodeIs(http.StatusOK)
			if name == "quit" {
				recorded.BodyIs(mustFourSeasonsOutputJSON("bye."))
			} else {
				recorded.BodyIs(expected)
			}
		})
	}

	// A foundation destination always needs idx: which corner opens is decided
	// by the order they are started, so it cannot be implied.
	badCases := map[string]string{
		"undo_n missing n":           `{"command":"undo_n","sessionId":"s1"}`,
		"move missing from/to":       `{"command":"m","sessionId":"s1"}`,
		"waste->tableau no idx":      `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"tableau"}}`,
		"waste->foundation no idx":   `{"command":"m","sessionId":"s1","from":{"zone":"waste"},"to":{"zone":"foundation"}}`,
		"tableau->tableau no idx":    `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"tableau"}}`,
		"tableau->foundation no idx": `{"command":"m","sessionId":"s1","from":{"zone":"tableau"},"to":{"zone":"foundation"}}`,
		"invalid zones":              `{"command":"m","sessionId":"s1","from":{"zone":"foundation","idx":0},"to":{"zone":"waste"}}`,
	}
	for name, raw := range badCases {
		t.Run(name, func(t *testing.T) {
			in := decode(raw)
			recorded := execRequest(t, ctrl.Exec, &in)
			recorded.CodeIs(http.StatusBadRequest)
		})
	}
}
