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

func mustFreeCellOutputJSON(msg string) string {
	out := &controller.FreeCellWebOutput{
		Tableau:       [][]*controller.WebOutputCard{},
		FreeCells:     []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustFreeCellOutputJSON: %v", err))
	}
	return string(b)
}

func TestFreeCellWebController(t *testing.T) {
	mockOutput := mustFreeCellOutputJSON("ok")

	fiMock := new(usecase.MockFreeCellInteractor)
	fiMock.On("Reset").Return(mockOutput)
	fiMock.On("GiveUp").Return(mockOutput)
	fiMock.On("Hint").Return(mockOutput)
	fiMock.On("AutoComplete").Return(mockOutput)
	fiMock.On("ActionLog").Return(mockOutput)
	fiMock.On("Undo").Return(mockOutput)
	fiMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	fiMock.On("MoveTableauToFoundation", 1).Return(mockOutput)
	fiMock.On("MoveTableauToFreeCell", 0, 1).Return(mockOutput)
	fiMock.On("MoveFreeCellToTableau", 1, 3).Return(mockOutput)
	fiMock.On("MoveFreeCellToFoundation", 0).Return(mockOutput)

	factory := func() uc.FreeCellInteractorIF { return fiMock }
	ctrl := controller.NewFreeCellWebController(factory)

	expectedBody := mockOutput

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "quit q",
			body:       `{"command":"q","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   mustFreeCellOutputJSON("bye."),
		},
		{
			name:       "quit",
			body:       `{"command":"quit","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   mustFreeCellOutputJSON("bye."),
		},
		{
			name:       "reset r",
			body:       `{"command":"r","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "reset",
			body:       `{"command":"reset","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "giveup g",
			body:       `{"command":"g","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "giveup",
			body:       `{"command":"giveup","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "hint h",
			body:       `{"command":"h","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "hint",
			body:       `{"command":"hint","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "autocomplete ac",
			body:       `{"command":"ac","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "autocomplete",
			body:       `{"command":"autocomplete","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "log",
			body:       `{"command":"log","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "log l",
			body:       `{"command":"l","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "undo u",
			body:       `{"command":"u","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "undo",
			body:       `{"command":"undo","sessionId":"s1"}`,
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name:       "unsupported command",
			body:       `{"command":"xyz","sessionId":"s1"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   mustFreeCellOutputJSON("Unsupported command."),
		},
		{
			name:       "empty command",
			body:       `{"command":"","sessionId":"s1"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   mustFreeCellOutputJSON("param error."),
		},
		{
			name:       "empty session",
			body:       `{"command":"r","sessionId":""}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   mustFreeCellOutputJSON("param error."),
		},
		{
			name:       "session too long",
			body:       fmt.Sprintf(`{"command":"r","sessionId":"%s"}`, strings.Repeat("a", controller.SessionMaxIDLen+1)),
			wantStatus: http.StatusBadRequest,
			wantBody:   mustFreeCellOutputJSON("param error."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := execRequest(t, ctrl.Exec, json.RawMessage(tt.body))
			recorded.CodeIs(tt.wantStatus)
			recorded.BodyIs(tt.wantBody)
		})
	}

	// Move tests using struct bodies
	type moveTest struct {
		name       string
		input      controller.FreeCellWebInput
		wantStatus int
		wantBody   string
	}

	moveTests := []moveTest{
		{
			name: "move tableau to tableau",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
				To:           &controller.FreeCellWebZone{Zone: "tableau", Col: intPtr(4)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move tableau to foundation",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "tableau", Col: intPtr(1)},
				To:           &controller.FreeCellWebZone{Zone: "foundation"},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move tableau to freecell",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "tableau", Col: intPtr(0)},
				To:           &controller.FreeCellWebZone{Zone: "freecell", Cell: intPtr(1)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move freecell to tableau",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "freecell", Cell: intPtr(1)},
				To:           &controller.FreeCellWebZone{Zone: "tableau", Col: intPtr(3)},
			},
			wantStatus: http.StatusOK,
			wantBody:   expectedBody,
		},
		{
			name: "move freecell to foundation",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "freecell", Cell: intPtr(0)},
				To:           &controller.FreeCellWebZone{Zone: "foundation"},
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

func TestFreeCellWebController_MoveErrors(t *testing.T) {
	fiMock := new(usecase.MockFreeCellInteractor)
	factory := func() uc.FreeCellInteractorIF { return fiMock }
	ctrl := controller.NewFreeCellWebController(factory)

	tests := []struct {
		name  string
		input controller.FreeCellWebInput
	}{
		{
			name: "move without from/to",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			},
		},
		{
			name: "move tableau to tableau missing params",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "tableau"},
				To:           &controller.FreeCellWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move tableau to foundation no col",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "tableau"},
				To:           &controller.FreeCellWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move tableau to freecell missing params",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "tableau"},
				To:           &controller.FreeCellWebZone{Zone: "freecell"},
			},
		},
		{
			name: "move freecell to tableau missing params",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "freecell"},
				To:           &controller.FreeCellWebZone{Zone: "tableau"},
			},
		},
		{
			name: "move freecell to foundation no cell",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "freecell"},
				To:           &controller.FreeCellWebZone{Zone: "foundation"},
			},
		},
		{
			name: "move invalid zones",
			input: controller.FreeCellWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
				From:         &controller.FreeCellWebZone{Zone: "invalid"},
				To:           &controller.FreeCellWebZone{Zone: "tableau", Col: intPtr(0)},
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

func TestFreeCellWebController_Stop(t *testing.T) {
	fiMock := new(usecase.MockFreeCellInteractor)
	factory := func() uc.FreeCellInteractorIF { return fiMock }
	ctrl := controller.NewFreeCellWebController(factory)
	ctrl.Stop()
	ctrl.Stop() // idempotent
}
