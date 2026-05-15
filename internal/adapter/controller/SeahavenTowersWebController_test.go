//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustSeahavenTowersOutputJSON(msg string) string {
	out := &controller.SeahavenTowersWebOutput{
		Tableau:       [][]*controller.WebOutputCard{},
		ReservedCells: []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSeahavenTowersOutputJSON: %v", err))
	}
	return string(b)
}

func TestSeahavenTowersWebController(t *testing.T) {
	mockOutput := mustSeahavenTowersOutputJSON("ok")

	siMock := new(usecase.MockSeahavenTowersInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("AutoComplete").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("Undo").Return(mockOutput)
	siMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	siMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	siMock.On("MoveTableauToFreeCell", 0, 1).Return(mockOutput)
	siMock.On("MoveFreeCellToTableau", 1, 3).Return(mockOutput)
	siMock.On("MoveFreeCellToFoundation", 0).Return(mockOutput)

	factory := func() uc.SeahavenTowersInteractorIF { return siMock }
	ctrl := controller.NewSeahavenTowersWebController(factory)

	expectedBody := mockOutput

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "quit q", body: `{"command":"q","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: mustSeahavenTowersOutputJSON("bye.")},
		{name: "quit", body: `{"command":"quit","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: mustSeahavenTowersOutputJSON("bye.")},
		{name: "reset r", body: `{"command":"r","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "reset", body: `{"command":"reset","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "giveup g", body: `{"command":"g","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "giveup", body: `{"command":"giveup","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "hint h", body: `{"command":"h","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "hint", body: `{"command":"hint","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "autocomplete ac", body: `{"command":"ac","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "autocomplete", body: `{"command":"autocomplete","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "log", body: `{"command":"log","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "log l", body: `{"command":"l","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "undo u", body: `{"command":"u","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "undo", body: `{"command":"undo","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: expectedBody},
		{name: "unsupported command", body: `{"command":"xyz","sessionId":"s1"}`, wantStatus: http.StatusBadRequest, wantBody: mustSeahavenTowersOutputJSON("Unsupported command.")},
		{name: "empty command", body: `{"command":"","sessionId":"s1"}`, wantStatus: http.StatusBadRequest, wantBody: mustSeahavenTowersOutputJSON("param error.")},
		{name: "empty session", body: `{"command":"r","sessionId":""}`, wantStatus: http.StatusBadRequest, wantBody: mustSeahavenTowersOutputJSON("param error.")},
		{name: "session too long", body: fmt.Sprintf(`{"command":"r","sessionId":"%s"}`, strings.Repeat("a", controller.SessionMaxIDLen+1)), wantStatus: http.StatusBadRequest, wantBody: mustSeahavenTowersOutputJSON("param error.")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := execRequest(t, ctrl.Exec, json.RawMessage(tt.body))
			recorded.CodeIs(tt.wantStatus)
			recorded.BodyIs(tt.wantBody)
		})
	}

	type moveTest struct {
		name       string
		input      controller.SeahavenTowersWebInput
		wantStatus int
		wantBody   string
	}
	moveTests := []moveTest{
		{
			name: "move tableau to tableau",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
				To:           &controller.SeahavenTowersWebZone{Zone: "tableau", Col: intPtr(4)},
			},
			wantStatus: http.StatusOK, wantBody: expectedBody,
		},
		{
			name: "move tableau to foundation",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "tableau", Col: intPtr(1)},
				To:           &controller.SeahavenTowersWebZone{Zone: "foundation"},
			},
			wantStatus: http.StatusOK, wantBody: expectedBody,
		},
		{
			name: "move tableau to reserved",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "tableau", Col: intPtr(0)},
				To:           &controller.SeahavenTowersWebZone{Zone: "reserved", Cell: intPtr(1)},
			},
			wantStatus: http.StatusOK, wantBody: expectedBody,
		},
		{
			name: "move reserved to tableau",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "reserved", Cell: intPtr(1)},
				To:           &controller.SeahavenTowersWebZone{Zone: "tableau", Col: intPtr(3)},
			},
			wantStatus: http.StatusOK, wantBody: expectedBody,
		},
		{
			name: "move reserved to foundation",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "reserved", Cell: intPtr(0)},
				To:           &controller.SeahavenTowersWebZone{Zone: "foundation"},
			},
			wantStatus: http.StatusOK, wantBody: expectedBody,
		},
	}

	for _, tt := range moveTests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := execRequest(t, ctrl.Exec, tt.input)
			recorded.CodeIs(tt.wantStatus)
			recorded.BodyIs(tt.wantBody)
		})
	}
}

func TestSeahavenTowersWebController_UndoN(t *testing.T) {
	mockOutput := mustSeahavenTowersOutputJSON("ok")

	siMock := new(usecase.MockSeahavenTowersInteractor)
	siMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.SeahavenTowersInteractorIF { return siMock }
	ctrl := controller.NewSeahavenTowersWebController(factory)

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		input := controller.SeahavenTowersWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		input := controller.SeahavenTowersWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestSeahavenTowersWebController_MoveErrors(t *testing.T) {
	siMock := new(usecase.MockSeahavenTowersInteractor)
	factory := func() uc.SeahavenTowersInteractorIF { return siMock }
	ctrl := controller.NewSeahavenTowersWebController(factory)

	tests := []struct {
		name  string
		input controller.SeahavenTowersWebInput
	}{
		{
			name: "move without from/to",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			},
		},
		{
			name: "move tableau to tableau missing params",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "tableau"},
				To:           &controller.SeahavenTowersWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move tableau to foundation no col",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "tableau"},
				To:           &controller.SeahavenTowersWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move tableau to reserved missing params",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "tableau"},
				To:           &controller.SeahavenTowersWebZone{Zone: "reserved"},
			},
		},
		{
			name: "move reserved to tableau missing params",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "reserved"},
				To:           &controller.SeahavenTowersWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move reserved to foundation no cell",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "reserved"},
				To:           &controller.SeahavenTowersWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move invalid zones",
			input: controller.SeahavenTowersWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.SeahavenTowersWebZone{Zone: "invalid"},
				To:           &controller.SeahavenTowersWebZone{Zone: "tableau", Col: intPtr(0)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := execRequest(t, ctrl.Exec, tt.input)
			recorded.CodeIs(http.StatusBadRequest)
		})
	}
}

func TestSeahavenTowersWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockSeahavenTowersInteractor)
	factory := func() uc.SeahavenTowersInteractorIF { return siMock }
	ctrl := controller.NewSeahavenTowersWebController(factory)
	ctrl.Stop()
	ctrl.Stop()
}
