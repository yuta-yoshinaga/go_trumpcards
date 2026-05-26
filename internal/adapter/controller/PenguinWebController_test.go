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

func mustPenguinOutputJSON(msg string) string {
	out := &controller.PenguinWebOutput{
		Tableau:       [][]*controller.WebOutputCard{},
		FreeCells:     []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPenguinOutputJSON: %v", err))
	}
	return string(b)
}

func TestPenguinWebController(t *testing.T) {
	mockOutput := mustPenguinOutputJSON("ok")

	piMock := new(usecase.MockPenguinInteractor)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("GiveUp").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("AutoComplete").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)
	piMock.On("Undo").Return(mockOutput)
	piMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	piMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	piMock.On("MoveTableauToFreeCell", 0, 5).Return(mockOutput)
	piMock.On("MoveFreeCellToTableau", 1, 3).Return(mockOutput)
	piMock.On("MoveFreeCellToFoundation", 0).Return(mockOutput)

	factory := func() uc.PenguinInteractorIF { return piMock }
	ctrl := controller.NewPenguinWebController(factory)

	expectedBody := mockOutput

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{name: "quit q", body: `{"command":"q","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: mustPenguinOutputJSON("bye.")},
		{name: "quit", body: `{"command":"quit","sessionId":"s1"}`, wantStatus: http.StatusOK, wantBody: mustPenguinOutputJSON("bye.")},
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
		{name: "unsupported command", body: `{"command":"xyz","sessionId":"s1"}`, wantStatus: http.StatusBadRequest, wantBody: mustPenguinOutputJSON("Unsupported command.")},
		{name: "empty command", body: `{"command":"","sessionId":"s1"}`, wantStatus: http.StatusBadRequest, wantBody: mustPenguinOutputJSON("param error.")},
		{name: "empty session", body: `{"command":"r","sessionId":""}`, wantStatus: http.StatusBadRequest, wantBody: mustPenguinOutputJSON("param error.")},
		{name: "session too long", body: fmt.Sprintf(`{"command":"r","sessionId":"%s"}`, strings.Repeat("a", controller.SessionMaxIDLen+1)), wantStatus: http.StatusBadRequest, wantBody: mustPenguinOutputJSON("param error.")},
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
		input      controller.PenguinWebInput
		wantStatus int
		wantBody   string
	}

	moveTests := []moveTest{
		{
			name: "move tableau to tableau",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
				To:           &controller.PenguinWebZone{Zone: "tableau", Col: intPtr(4)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move tableau to foundation",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "tableau", Col: intPtr(1)},
				To:           &controller.PenguinWebZone{Zone: "foundation"},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move tableau to freecell",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "tableau", Col: intPtr(0)},
				To:           &controller.PenguinWebZone{Zone: "freecell", Cell: intPtr(5)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move freecell to tableau",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "freecell", Cell: intPtr(1)},
				To:           &controller.PenguinWebZone{Zone: "tableau", Col: intPtr(3)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move freecell to foundation",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "freecell", Cell: intPtr(0)},
				To:           &controller.PenguinWebZone{Zone: "foundation"},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
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

func TestPenguinWebController_UndoN(t *testing.T) {
	mockOutput := mustPenguinOutputJSON("ok")

	piMock := new(usecase.MockPenguinInteractor)
	piMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.PenguinInteractorIF { return piMock }
	ctrl := controller.NewPenguinWebController(factory)

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		input := controller.PenguinWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		input := controller.PenguinWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestPenguinWebController_MoveErrors(t *testing.T) {
	piMock := new(usecase.MockPenguinInteractor)
	factory := func() uc.PenguinInteractorIF { return piMock }
	ctrl := controller.NewPenguinWebController(factory)

	tests := []struct {
		name  string
		input controller.PenguinWebInput
	}{
		{
			name: "move without from/to",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			},
		},
		{
			name: "move tableau to tableau missing params",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "tableau"},
				To:           &controller.PenguinWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move tableau to foundation no col",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "tableau"},
				To:           &controller.PenguinWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move tableau to freecell missing params",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "tableau"},
				To:           &controller.PenguinWebZone{Zone: "freecell"},
			},
		},
		{
			name: "move freecell to tableau missing params",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "freecell"},
				To:           &controller.PenguinWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move freecell to foundation no cell",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "freecell"},
				To:           &controller.PenguinWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move invalid zones",
			input: controller.PenguinWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.PenguinWebZone{Zone: "invalid"},
				To:           &controller.PenguinWebZone{Zone: "tableau", Col: intPtr(0)},
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

func TestPenguinWebController_Stop(t *testing.T) {
	piMock := new(usecase.MockPenguinInteractor)
	factory := func() uc.PenguinInteractorIF { return piMock }
	ctrl := controller.NewPenguinWebController(factory)
	ctrl.Stop()
	ctrl.Stop() // idempotent
}
